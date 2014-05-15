// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package gbif

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

type spAnswer struct {
	Offset, Limit int64
	EndOfRecords  bool
	Results       []*species
}

type species struct {
	Key, NubKey, AcceptedKey int64  // id
	CanonicalName            string // name
	Authorship               string // author
	Rank                     string // rank
	Synonym                  bool   // valid
	AccordingTo              string // source
	ParentKey                int64  // parent

	//parents (it seems that gbif knows nothing about arrays)
	KingdomKey int64
	PhylumKey  int64
	ClassKey   int64
	OrderKey   int64
	FamilyKey  int64
	GenusKey   int64

	Kingdom string
	Phylum  string
	Clazz   string
	Order   string
	Family  string
	Genus   string
}

// returns a copy of species
func (sp *species) copy() *jdh.Taxon {
	if sp.Key == 0 {
		return &jdh.Taxon{}
	}
	rank := jdh.GetRank(strings.TrimSpace(sp.Rank))
	tax := &jdh.Taxon{
		Id:        strconv.FormatInt(sp.Key, 10),
		Name:      strings.Join(strings.Fields(sp.CanonicalName), " "),
		Authority: strings.Join(strings.Fields(sp.Authorship), " "),
		Rank:      rank,
		IsValid:   !sp.Synonym,
		Comment:   strings.Join(strings.Fields(sp.AccordingTo), " "),
	}
	if sp.Synonym {
		tax.Parent = strconv.FormatInt(sp.AcceptedKey, 10)
	} else if sp.ParentKey > 0 {
		tax.Parent = strconv.FormatInt(sp.ParentKey, 10)
	}
	return tax
}

// isDesc returns true if p is an ancestor of a given species.
func (sp *species) isDesc(p int64) bool {
	if p == 0 {
		return false
	}
	if sp.KingdomKey == p {
		return true
	}
	if sp.PhylumKey == p {
		return true
	}
	if sp.ClassKey == p {
		return true
	}
	if sp.OrderKey == p {
		return true
	}
	if sp.FamilyKey == p {
		return true
	}
	if sp.GenusKey == p {
		return true
	}
	return false
}

// hasParentName returns true if p has a parent of a given name.
func (sp *species) hasParentName(p string) bool {
	if len(p) == 0 {
		return false
	}
	if strings.ToLower(sp.Kingdom) == p {
		return true
	}
	if strings.ToLower(sp.Phylum) == p {
		return true
	}
	if strings.ToLower(sp.Clazz) == p {
		return true
	}
	if strings.ToLower(sp.Order) == p {
		return true
	}
	if strings.ToLower(sp.Family) == p {
		return true
	}
	if strings.ToLower(sp.Genus) == p {
		return true
	}
	return false
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
			if len(kv.Value) == 0 {
				l.Close()
				ok = true
				break
			}
			id := strings.TrimSpace(kv.Value[0])
			if (len(id) == 0) || (id == "0") {
				return nil, errors.New("taxon without identification")
			}
			go db.synonyms(l, id)
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
			break
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

var kingdoms = []string{
	"1", // Animalia
	"2", // Archaea
	"3", // Bacteria
	"4", // Chromista
	"5", // Fungi
	"6", // Plantae
	"7", // Protozoa
	"8", // Viruses
}

// childs search for a list with the children of a taxon.
func (db *DB) childs(l *listScanner, id string) {
	if (len(id) == 0) || (id == "0") {
		for _, k := range kingdoms {
			sp, err := db.getSpecies(k)
			if err != nil {
				l.setErr(err)
				return
			}
			select {
			case l.c <- sp.copy():
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
	for off := int64(0); ; {
		request := wsHead + "species/" + id + "/children"
		if off > 0 {
			request += "?offset=" + strconv.FormatInt(off, 10)
		}
		an := new(spAnswer)
		if err := db.listRequest(request, an); err != nil {
			l.setErr(err)
			return
		}
		for _, sp := range an.Results {
			if sp.Key != sp.NubKey {
				continue
			}
			select {
			case l.c <- sp.copy():
			case <-l.end:
				return
			}
		}
		if an.EndOfRecords {
			break
		}
		off += an.Limit
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// parents searcg for a list of a parents of a taxon
func (db *DB) parents(l *listScanner, id string) {
	// check if the name is a synonym (in gbif synonyms are not
	// attached to their senior synonym via parents.
	sp, err := db.getSpecies(id)
	if err != nil {
		l.setErr(err)
		return
	}
	if sp.Synonym {
		sp, err = db.getSpecies(strconv.FormatInt(sp.AcceptedKey, 10))
		if err != nil {
			l.setErr(err)
			return
		}
		select {
		case l.c <- sp.copy():
		case <-l.end:
			return
		}
	}
	request := wsHead + "species/" + id + "/parents"
	db.request <- request
	a := <-db.answer
	switch answer := a.(type) {
	case error:
		l.setErr(answer)
		return
	case *http.Response:
		defer answer.Body.Close()
		d := json.NewDecoder(answer.Body)
		var pl []species
		if err := d.Decode(&pl); err != nil {
			l.setErr(err)
			return
		}
		for i := len(pl) - 1; i >= 0; i-- {
			select {
			case l.c <- pl[i].copy():
			case <-l.end:
				return
			}
		}
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// searchTaxon searchs for taxon name in gbif.
func (db *DB) searchTaxon(l *listScanner, name string, kvs []jdh.KeyValue) {
	nm := strings.Join(strings.Fields(name), " ")
	if i := strings.Index(nm, "*"); i > 0 {
		l.setErr(errors.New("current gbif api does not support partial lookups"))
		return
	}
	var pId int64
	var pName string
	var rank jdh.Rank
	for _, kv := range kvs {
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.TaxParent:
			pId, _ = strconv.ParseInt(kv.Value[0], 10, 64)
		case jdh.TaxRank:
			rank = jdh.GetRank(kv.Value[0])
		case jdh.TaxParentName:
			pName = strings.ToLower(strings.Join(strings.Fields(kv.Value[0]), " "))
		}
	}
	vals := url.Values{}
	vals.Add("name", nm)
	for off := int64(0); ; {
		if off > 0 {
			vals.Set("offset", strconv.FormatInt(off, 10))
		}
		request := wsHead + "species?" + vals.Encode()
		an := new(spAnswer)
		if err := db.listRequest(request, an); err != nil {
			l.setErr(err)
			return
		}
		for _, sp := range an.Results {
			if sp.Key != sp.NubKey {
				continue
			}
			if pId != 0 {
				if !sp.isDesc(pId) {
					continue
				}
			}
			if rank != jdh.Unranked {
				if rank != jdh.GetRank(strings.TrimSpace(sp.Rank)) {
					continue
				}
			}
			if len(pName) > 0 {
				if !sp.hasParentName(pName) {
					continue
				}
			}
			select {
			case l.c <- sp.copy():
			case <-l.end:
				return
			}
		}
		if an.EndOfRecords {
			break
		}
		off += an.Limit
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// childs search for a list with the children of a taxon.
func (db *DB) synonyms(l *listScanner, id string) {
	if (len(id) == 0) || (id == "0") {
		select {
		case l.c <- nil:
		case <-l.end:
		}
		return
	}
	for off := int64(0); ; {
		request := wsHead + "species/" + id + "/synonyms"
		if off > 0 {
			request += "?offset=" + strconv.FormatInt(off, 10)
		}
		an := new(spAnswer)
		if err := db.listRequest(request, an); err != nil {
			l.setErr(err)
			return
		}
		for _, sp := range an.Results {
			if sp.Key != sp.NubKey {
				continue
			}
			select {
			case l.c <- sp.copy():
			case <-l.end:
				return
			}
		}
		if an.EndOfRecords {
			break
		}
		off += an.Limit
	}
	select {
	case l.c <- nil:
	case <-l.end:
	}
}

// taxon returns a jdh scanner with a taxon.
func (db *DB) taxon(id string) (jdh.Scanner, error) {
	if len(id) == 0 {
		return nil, errors.New("taxon without identification")
	}
	sp, err := db.getSpecies(id)
	if err != nil {
		return nil, err
	}
	return &getScanner{val: sp.copy()}, nil
}

func (db *DB) getSpecies(id string) (*species, error) {
	sp := &species{}
	for {
		request := wsHead + "species/" + id
		if err := db.simpleRequest(request, sp); err != nil {
			return nil, err
		}
		if sp.Key == sp.NubKey {
			break
		}
		id = strconv.FormatInt(sp.NubKey, 10)
	}
	return sp, nil
}
