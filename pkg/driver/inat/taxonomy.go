// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package inat

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
	id   string
	name string
	rank string
}

// returns a copy of a taxon
func (tx *taxon) copy() *jdh.Taxon {
	tax := &jdh.Taxon{
		Id:   strings.TrimSpace(tx.id),
		Name: strings.Join(strings.Fields(tx.name), " "),
		Rank: jdh.GetRank(strings.TrimSpace(tx.rank)),
	}
	return tax
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
			id := ""
			if len(kv.Value) > 0 {
				id = strings.TrimSpace(kv.Value[0])
			}
			go db.childs(l, id)
			ok = true
		case jdh.TaxParents:
			if len(kv.Value) == 0 {
				l.Close()
				ok = true
				break
			}
			id := strings.TrimSpace(kv.Value[0])
			if (len(id) == 0) || (id == "0") {
				return nil, errors.New("taxon without identification")
			}
			go db.parents(l, id)
			ok = true
		case jdh.TaxSynonyms:
			l.Close()
			ok = true
		case jdh.TaxName:
			if len(kv.Value) == 0 {
				return nil, errors.New("taxon without identification")
			}
			nm := strings.Join(strings.Fields(kv.Value[0]), " ")
			if len(nm) == 0 {
				return nil, errors.New("taxon without identification")
			}
			go db.searchTaxon(l, nm, kvs)
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

var kingdoms = []taxon{
	taxon{id: "48222", name: "Chromista", rank: "kingdom"},
	taxon{id: "1", name: "Animalia", rank: "kingdom"},
	taxon{id: "47126", name: "Plantae", rank: "kingdom"},
	taxon{id: "47686", name: "Protozoa", rank: "kingdom"},
	taxon{id: "47170", name: "Fungi", rank: "kingdom"},
}

// childs returns a list with the children of a taxon.
func (db *DB) childs(l *listScanner, id string) {
	if (len(id) == 0) || (id == "48460") {
		for _, k := range kingdoms {
			select {
			case l.c <- k.copy():
			case <-l.end:
				return
			}
		}
		select {
		case l.c <- nil:
		case <-l.end:
		}
		return
	}
	ls, err := db.txList(id)
	if err != nil {
		l.setErr(err)
		return
	}
	pars := true
	for _, tx := range ls {
		if tx.id == id {
			pars = false
			continue
		}
		if pars {
			continue
		}
		select {
		case l.c <- tx.copy():
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
	ls, err := db.txList(id)
	if err != nil {
		l.setErr(err)
		return
	}
	for _, tx := range ls {
		if tx.id == id {
			break
		}
		if tx.id == "48460" {
			continue
		}
		select {
		case l.c <- tx.copy():
		case <-l.end:
			return
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// SearchTaxon search for a taxon name in iNaturalist.
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
		nm = nm[:i]
	}
	pId := ""
	var rank jdh.Rank
	pName := ""
	for _, kv := range kvs {
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.TaxParent:
			pId = kv.Value[0]
		case jdh.TaxRank:
			rank = jdh.GetRank(kv.Value[0])
		case jdh.TaxParentName:
			pName = strings.Join(strings.Fields(kv.Value[0]), " ")
		}
	}
	vals := url.Values{}
	vals.Add("q", name)
	vals.Add("utf8", "✓")
	for next := 1; ; next++ {
		if next > 1 {
			vals.Set("page", strconv.FormatInt(int64(next), 10))
		}
		ls, nx, err := db.txSearch(inatHead + "taxa/search?" + vals.Encode())
		if err != nil {
			l.setErr(err)
			return
		}
		for _, tx := range ls {
			tax := tx.copy()
			if len(prefix) > 0 {
				if strings.HasPrefix(tax.Name, prefix) {
					if len(pId) > 0 {
						pl, err := db.txList(tx.id)
						if err != nil {
							l.setErr(err)
							return
						}
						isP := false
						for _, pt := range pl {
							if pt.id == tx.id {
								break
							}
							if pt.id == pId {
								isP = true
								break
							}
						}
						if !isP {
							continue
						}
					}
					if len(pName) > 0 {
						pl, err := db.txList(tx.id)
						if err != nil {
							l.setErr(err)
							return
						}
						isP := false
						for _, pt := range pl {
							if pt.id == tx.id {
								break
							}
							if strings.Join(strings.Fields(pt.name), " ") == pName {
								isP = true
								break
							}
						}
						if !isP {
							continue
						}
					}
					if rank != jdh.Unranked {
						if rank != tax.Rank {
							continue
						}
					}
					select {
					case l.c <- tax:
					case <-l.end:
						return
					}
				}
				continue
			}
			if tax.Name == nm {
				if len(pId) > 0 {
					pl, err := db.txList(tx.id)
					if err != nil {
						l.setErr(err)
						return
					}
					isP := false
					for _, pt := range pl {
						if pt.id == tx.id {
							break
						}
						if pt.id == pId {
							isP = true
							break
						}
					}
					if !isP {
						continue
					}
				}
				if len(pName) > 0 {
					pl, err := db.txList(tx.id)
					if err != nil {
						l.setErr(err)
						return
					}
					isP := false
					for _, pt := range pl {
						if pt.id == tx.id {
							break
						}
						if strings.Join(strings.Fields(pt.name), " ") == pName {
							isP = true
							break
						}
					}
					if !isP {
						continue
					}
				}
				if rank != jdh.Unranked {
					if rank != tax.Rank {
						continue
					}
				}
				select {
				case l.c <- tax:
				case <-l.end:
					return
				}
			}
		}
		if !nx {
			break
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// taxon returns a jdh scanner with a taxon.
func (db *DB) taxon(id string) (jdh.Scanner, error) {
	if (len(id) == 0) || (id == "48460") {
		return nil, errors.New("taxon without identification")
	}
	ls, err := db.txList(id)
	if err != nil {
		return nil, err
	}
	var tax *jdh.Taxon
	for _, tx := range ls {
		if tx.id == id {

			tax = tx.copy()
			break
		}
	}
	if tax == nil {
		tax = &jdh.Taxon{}
	}
	return &getScanner{val: tax}, nil
}

func (db *DB) txList(id string) ([]taxon, error) {
	request := inatHead + "taxa/" + id
	db.request <- request
	a := <-db.answer
	var ls []taxon
	switch answer := a.(type) {
	case error:
		return nil, answer
	case *http.Response:
		defer answer.Body.Close()
		dec := xml.NewDecoder(answer.Body)
		dec.Strict = false
		dec.AutoClose = xml.HTMLAutoClose
		dec.Entity = xml.HTMLEntity
		end := false
		for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
			if err != nil {
				return nil, err
			}
			switch t := tk.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "ul":
					for _, at := range t.Attr {
						if (at.Name.Local == "class") && (at.Value == "taxonomic_tree leafylist") {
							ls, err = readLeafList(dec)
							if err != nil {
								return nil, err
							}
							break
						}
						if (at.Name.Local == "id") && (at.Value == "more_children") {
							ot, err := readLeafList(dec)
							if err != nil {
								return nil, err
							}
							ls = append(ls, ot...)
							break
						}
					}
				case "div":
					for _, at := range t.Attr {
						if (at.Name.Local == "id") && (at.Value == "extras") {
							end = true
						}
					}
				}
			}
			if end {
				break
			}
		}
	}
	return ls, nil
}

func readLeafList(dec *xml.Decoder) ([]taxon, error) {
	var tx []taxon
	tax := taxon{}
	rank := false
	sciname := false
	for tk, err := dec.Token(); ; tk, err = dec.Token() {
		if err != nil {
			return nil, err
		}
		switch t := tk.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "ul":
				if len(tax.id) > 0 {
					tx = append(tx, tax)
					tax = taxon{}
				}
			case "li":
				tax = taxon{}
			case "span":
				for _, at := range t.Attr {
					if at.Name.Local == "class" {
						cf := strings.Fields(at.Value)
						if len(cf) == 0 {
							break
						}
						switch cf[0] {
						case "taxon":
							if len(cf) < 2 {
								break
							}
							ids := strings.Split(cf[1], "-")
							if len(ids) < 2 {
								break
							}
							tax.id = ids[1]
							if len(cf) < 3 {
								break
							}
							tax.rank = cf[2]
						case "rank":
							rank = true
						case "sciname":
							sciname = true
						}
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "ul":
				return tx, nil
			case "li":
				if len(tax.id) > 0 {
					tx = append(tx, tax)
					tax = taxon{}
				}
			case "span":
				if rank {
					rank = false
					break
				}
				if sciname {
					sciname = false
				}
			}
		case xml.CharData:
			if sciname && (!rank) {
				tax.name += string(t)
			}
		}
	}
}

func (db *DB) txSearch(request string) ([]taxon, bool, error) {
	db.request <- request
	a := <-db.answer
	var tx []taxon
	next := false
	switch answer := a.(type) {
	case error:
		return nil, false, answer
	case *http.Response:
		defer answer.Body.Close()
		dec := xml.NewDecoder(answer.Body)
		dec.Strict = false
		dec.AutoClose = xml.HTMLAutoClose
		dec.Entity = xml.HTMLEntity
		tax := taxon{}
		info := false
		pag := false
		rank := false
		sciname := false
		for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
			if err != nil {
				return nil, false, err
			}
			switch t := tk.(type) {
			case xml.StartElement:
				switch t.Name.Local {
				case "div":
					for _, at := range t.Attr {
						if at.Name.Local == "class" {
							switch at.Value {
							case "info":
								info = true
								tax = taxon{}
							case "pagination":
								pag = true
							}
						}
					}
				case "span":
					for _, at := range t.Attr {
						if at.Name.Local == "class" {
							cf := strings.Fields(at.Value)
							if len(cf) == 0 {
								break
							}
							switch cf[0] {
							case "taxon":
								if len(cf) < 2 {
									break
								}
								ids := strings.Split(cf[1], "-")
								if len(ids) < 2 {
									break
								}
								tax.id = ids[1]
								if len(cf) < 3 {
									break
								}
								tax.rank = cf[2]
							case "rank":
								rank = true
							case "sciname":
								sciname = true
							}
						}
					}
				case "a":
					if !pag {
						continue
					}
					for _, at := range t.Attr {
						if (at.Name.Local == "class") && (at.Value == "next_page") {
							next = true
							break
						}
					}
				}
			case xml.EndElement:
				switch t.Name.Local {
				case "div":
					if info && (len(tax.id) > 0) {
						tx = append(tx, tax)
					}
					info, pag = false, false
				case "span":
					if rank {
						rank = false
						break
					}
					if sciname {
						sciname = false
					}
				}
			case xml.CharData:
				if sciname && (!rank) {
					tax.name += string(t)
				}
			}
		}
	}
	return tx, next, nil
}
