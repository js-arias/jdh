// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package ncbi

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

type taxon struct {
	tax *jdh.Taxon
	syn []string // synonyms
	par []string // parent ids
	lin []string // lineage
}

// taxon list returns a list scanner with a list of taxons.
func (db *DB) taxonList(kvs []jdh.KeyValue) (jdh.ListScanner, error) {
	l := &listScanner{
		c:   make(chan interface{}, 20),
		end: make(chan struct{}),
	}
	ok := false
	for _, kv := range kvs {
		switch kv.Key {
		case jdh.TaxChildren:
			go db.childs(l, kv.Value)
			ok = true
		case jdh.TaxParents:
			if (len(kv.Value) == 0) || (kv.Value == "0") {
				return nil, errors.New("taxon without identification")
			}
			go db.parents(l, kv.Value)
			ok = true
		case jdh.TaxSynonyms:
			go db.synonyms(l, kv.Value)
			ok = true
		case jdh.TaxName:
			go db.searchTaxon(l, kv.Value, kvs)
			ok = true
		}
		if ok {
			break
		}
	}
	if !ok {
		return nil, errors.New("invalid argument list")
	}
	return l, nil
}

// childs returns a list with the children of a taxon.
func (db *DB) childs(l *listScanner, id string) {
	if (len(id) == 0) || (id == "0") {
		id = "1"
	}
	idv := strings.Split(id, ".")
	if len(idv) > 1 {
		select {
		case l.c <- nil:
		case <-l.end:
		}
		return
	}
	var lt []string
	request := emblHead + "Taxon:" + id + "&display=xml"
	db.request <- request
	a := <-db.answer
	switch answer := a.(type) {
	case error:
		l.setErr(answer)
		return
	case *http.Response:
		defer answer.Body.Close()
		dec := xml.NewDecoder(answer.Body)
		for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
			if err != nil {
				l.setErr(err)
				return
			}
			switch t := tk.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "children":
					lt, err = readTaxonList(dec)
					if err != nil {
						l.setErr(err)
						return
					}
				}
			}
		}
	}
	for _, nid := range lt {
		tax, err := db.getTaxon(nid)
		if err != nil {
			l.setErr(err)
			return
		}
		select {
		case l.c <- tax:
		case <-l.end:
			return
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// synonyms returns a list with the synonyms of a taxon.
func (db *DB) synonyms(l *listScanner, id string) {
	if (len(id) == 0) || (id == "0") {
		id = "1"
	}
	idv := strings.Split(id, ".")
	if len(idv) > 1 {
		select {
		case l.c <- nil:
		case <-l.end:
		}
		return
	}
	tx, err := readTaxon(idv[0])
	if err != nil {
		l.setErr(err)
		return
	}
	if len(tx.syn) == 0 {
		select {
		case l.c <- nil:
		case <-l.end:
		}
		return
	}
	for i, s := range tx.syn {
		st := &jdh.Taxon{
			Id:      idv[0] + "." + strconv.FormatInt(int64(i)+1, 10),
			Name:    s,
			IsValid: false,
			Parent:  idv[0],
			Rank:    tx.tax.Rank,
		}
		select {
		case l.c <- st:
		case <-l.end:
			return
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// parents returns a list with the parents of a taxon.
func (db *DB) parents(l *listScanner, id string) {
	var lt []string
	idv := strings.Split(id, ".")
	request := emblHead + "Taxon:" + idv[0] + "&display=xml"
	db.request <- request
	a := <-db.answer
	switch answer := a.(type) {
	case error:
		l.setErr(answer)
		return
	case *http.Response:
		defer answer.Body.Close()
		dec := xml.NewDecoder(answer.Body)
		if len(idv) > 1 {
			lt = append(lt, idv[0])
		}
		for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
			if err != nil {
				l.setErr(err)
				return
			}
			switch t := tk.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "lineage":
					lx, err := readTaxonList(dec)
					if err != nil {
						l.setErr(err)
						return
					}
					if len(lt) == 0 {
						lt = lx
					} else {
						lt = append(lt, lx...)
					}
				}
			}
		}
	}
	for _, nid := range lt {
		tax, err := db.getTaxon(nid)
		if err != nil {
			l.setErr(err)
			return
		}
		select {
		case l.c <- tax:
		case <-l.end:
			return
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// SearchTaxon search for a taxon name in ncbi.
func (db *DB) searchTaxon(l *listScanner, name string, kvs []jdh.KeyValue) {
	nm := strings.Join(strings.Fields(name), " ")
	i := strings.Index(nm, "*")
	prefix := ""
	if i == 0 {
		l.setErr(errors.New("taxon without identification"))
		return
	}
	if i > 0 {
		prefix = nm[:i]
	}
	pId := ""
	var rank jdh.Rank
	pName := ""
	for _, kv := range kvs {
		switch kv.Key {
		case jdh.TaxParent:
			pId = kv.Value
		case jdh.TaxRank:
			rank = jdh.GetRank(kv.Value)
		case jdh.TaxParentName:
			pName = strings.Join(strings.Fields(name), " ")
		}
	}
	for next := 0; ; {
		vals := url.Values{}
		vals.Add("db", "taxonomy")
		vals.Add("term", nm)
		if next > 0 {
			vals.Add("RetStart", strconv.FormatInt(int64(next), 10))
		}
		request := ncbiHead + "esearch.fcgi?" + vals.Encode()
		nl, nx, err := db.search(request)
		if err != nil {
			l.setErr(err)
			return
		}
		for _, id := range nl {
			tx, err := readTaxon(id)
			if err != nil {
				l.setErr(err)
				return
			}
			if len(prefix) > 0 {
				if strings.HasPrefix(tx.tax.Name, prefix) {
					if len(pId) > 0 {
						isP := false
						for _, p := range tx.par {
							if pId == p {
								isP = true
								break
							}
						}
						if !isP {
							continue
						}
					}
					if rank != jdh.Unranked {
						if rank != tx.tax.Rank {
							continue
						}
					}
					if len(pName) > 0 {
						isP := false
						for _, p := range tx.lin {
							if pName == p {
								isP = true
								break
							}
						}
						if !isP {
							continue
						}
					}
					select {
					case l.c <- tx.tax:
					case <-l.end:
						return
					}
					continue
				}
				for i, s := range tx.syn {
					if strings.HasPrefix(s, prefix) {
						if len(pId) > 0 {
							isP := false
							for _, p := range tx.par {
								if pId == p {
									isP = true
									break
								}
							}
							if !isP {
								continue
							}
						}
						if rank != jdh.Unranked {
							if rank != tx.tax.Rank {
								continue
							}
						}
						if len(pName) > 0 {
							isP := false
							for _, p := range tx.lin {
								if pName == p {
									isP = true
									break
								}
							}
							if !isP {
								continue
							}
						}
						st := &jdh.Taxon{
							Id:      tx.tax.Id + "." + strconv.FormatInt(int64(i)+1, 10),
							Name:    s,
							IsValid: false,
							Parent:  tx.tax.Id,
							Rank:    tx.tax.Rank,
						}
						select {
						case l.c <- st:
						case <-l.end:
							return
						}
						break
					}
				}
				continue
			}
			if tx.tax.Name == nm {
				if len(pId) > 0 {
					isP := false
					for _, p := range tx.par {
						if pId == p {
							isP = true
							break
						}
					}
					if !isP {
						continue
					}
				}
				if rank != jdh.Unranked {
					if rank != tx.tax.Rank {
						continue
					}
				}
				if len(pName) > 0 {
					isP := false
					for _, p := range tx.lin {
						if pName == p {
							isP = true
							break
						}
					}
					if !isP {
						continue
					}
				}
				select {
				case l.c <- tx.tax:
				case <-l.end:
					return
				}
				continue
			}
			for i, s := range tx.syn {
				if s == nm {
					if len(pId) > 0 {
						isP := false
						for _, p := range tx.par {
							if pId == p {
								isP = true
								break
							}
						}
						if !isP {
							continue
						}
					}
					if rank != jdh.Unranked {
						if rank != tx.tax.Rank {
							continue
						}
					}
					st := &jdh.Taxon{
						Id:      tx.tax.Id + "." + strconv.FormatInt(int64(i)+1, 10),
						Name:    s,
						IsValid: false,
						Parent:  tx.tax.Id,
						Rank:    tx.tax.Rank,
					}
					select {
					case l.c <- st:
					case <-l.end:
						return
					}
					break
				}
			}
		}
		if nx == 0 {
			break
		}
		next = nx
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

func readTaxon(id string) (*taxon, error) {
	if (len(id) == 0) || (id == "0") {
		return nil, errors.New("taxon without identification")
	}
	request := emblHead + "Taxon:" + id + "&display=xml"
	answer, err := http.Get(request)
	if err != nil {
		return nil, err
	}
	defer answer.Body.Close()
	dec := xml.NewDecoder(answer.Body)
	tx := &taxon{
		tax: &jdh.Taxon{},
	}
	for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
		if err != nil {
			return nil, err
		}
		switch t := tk.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "taxon":
				for _, at := range t.Attr {
					switch at.Name.Local {
					case "scientificName":
						tx.tax.Name = at.Value
						tx.tax.IsValid = true
					case "taxId":
						tx.tax.Id = at.Value
					case "parentTaxId":
						tx.tax.Parent = at.Value
					case "rank":
						tx.tax.Rank = jdh.GetRank(at.Value)
					}
				}
			case "lineage":
				if err = readTaxonLineage(dec, tx); err != nil {
					return nil, err
				}
			case "children":
				skip(dec, t.Name.Local)
			case "synonym":
				nm := ""
				isSyn := false
				for _, at := range t.Attr {
					switch at.Name.Local {
					case "type":
						if at.Value == "synonym" {
							isSyn = true
						}
					case "name":
						nm = at.Value
					}
				}
				if isSyn {
					tx.syn = append(tx.syn, nm)
				}
			}
		}
	}
	return tx, nil
}

// taxon returns a jdh scanner with a taxon.
func (db *DB) taxon(id string) (jdh.Scanner, error) {
	if (len(id) == 0) || (id == "0") {
		return nil, errors.New("taxon without identification")
	}
	tax, err := db.getTaxon(id)
	if err != nil {
		return nil, err
	}
	return &getScanner{val: tax}, nil
}

// getTaxon returns a jdh taxon.
func (db *DB) getTaxon(id string) (*jdh.Taxon, error) {
	idv := strings.Split(id, ".")
	tx, err := readTaxon(idv[0])
	if err != nil {
		return nil, err
	}
	if len(idv) > 1 {
		seek, err := strconv.ParseInt(idv[1], 10, 64)
		if err != nil {
			return nil, err
		}
		if int(seek) > len(tx.syn) {
			return &jdh.Taxon{}, nil
		}
		tx.tax.Name = tx.syn[int(seek)-1]
		tx.tax.Id += "." + idv[1]
		tx.tax.Parent = idv[0]
		tx.tax.IsValid = false
	}
	return tx.tax, nil
}

func readTaxonList(dec *xml.Decoder) ([]string, error) {
	var l []string
	for tk, err := dec.Token(); ; tk, err = dec.Token() {
		if err != nil {
			return nil, err
		}
		switch t := tk.(type) {
		case xml.StartElement:
			if t.Name.Local == "taxon" {
				for _, at := range t.Attr {
					if at.Name.Local == "taxId" {
						l = append(l, at.Value)
						break
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local != "taxon" {
				return l, nil
			}
		}
	}
}

func readTaxonLineage(dec *xml.Decoder, tx *taxon) error {
	for tk, err := dec.Token(); ; tk, err = dec.Token() {
		if err != nil {
			return err
		}
		switch t := tk.(type) {
		case xml.StartElement:
			if t.Name.Local == "taxon" {
				for _, at := range t.Attr {
					switch at.Name.Local {
					case "taxId":
						tx.par = append(tx.par, at.Value)
					case "scientificName":
						tx.lin = append(tx.lin, at.Value)
					}

				}
			}
		case xml.EndElement:
			if t.Name.Local != "taxon" {
				return nil
			}
		}
	}
}
