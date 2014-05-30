// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package native

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/radix"
)

// Taxonomy holds the taxonomy of the database.
type taxonomy struct {
	db      *DB
	root    *taxon            // root of the taxonomy
	ids     map[string]*taxon // map of id:taxon
	names   *radix.Radix      // names of the taxonomy
	changed bool              // if true, the database has changed
	next    int64             // next valid id
}

// Taxon holds the taxon information.
type taxon struct {
	data *jdh.Taxon

	// relations
	parent *taxon
	childs []*taxon
}

// root id
const rootId = "0"

// taxonomy file
const taxFile = "taxonomy"

// OpenTaxonomy opens taxonomy data.
func openTaxonomy(db *DB) *taxonomy {
	t := &taxonomy{
		db:    db,
		root:  &taxon{},
		ids:   make(map[string]*taxon),
		names: radix.New(),
		next:  1,
	}
	p := filepath.Join(db.path, taxFile)
	f, err := os.Open(p)
	if err != nil {
		return t
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		tax := &jdh.Taxon{}
		if err := dec.Decode(tax); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-taxonomy: error: %v\n", err)
			break
		}
		t.setNext(tax.Id)
		if err := t.validate(tax); err != nil {
			log.Printf("db-taxonomy: error: %v\n", err)
			continue
		}
		t.addTaxon(tax)
	}
	return t
}

// SetNext sets the value of the next id.
func (t *taxonomy) setNext(id string) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}
	if v >= t.next {
		t.next = v + 1
	}
}

// Validate validates that a taxon is valid in the database, and set some
// canonical values. It returns an error if the taxon is not valid.
func (t *taxonomy) validate(tax *jdh.Taxon) error {
	tax.Id = strings.TrimSpace(tax.Id)
	tax.Name = strings.Join(strings.Fields(tax.Name), " ")
	if (len(tax.Id) == 0) || (len(tax.Name) == 0) {
		return errors.New("taxon without identification")
	}
	if _, ok := t.ids[tax.Id]; ok {
		return fmt.Errorf("taxon id %s already in use", tax.Id)
	}
	p := t.root
	if tax.Rank > jdh.Species {
		tax.Rank = jdh.Unranked
	}
	if len(tax.Parent) > 0 {
		var ok bool
		if p, ok = t.ids[tax.Parent]; !ok {
			return fmt.Errorf("taxon %s parent [%s] not in database", tax.Name, tax.Parent)
		}
	}
	if p == t.root {
		if !tax.IsValid {
			return fmt.Errorf("taxon %s is a synonym without a parent", tax.Name)
		}
	} else if !p.data.IsValid {
		return fmt.Errorf("taxon %s parent [%s] is a synonym", tax.Name, p.data.Name)
	}
	if !p.isDescValid(tax.Rank, tax.IsValid) {
		return fmt.Errorf("taxon %s rank incompatible with database hierarchy", tax.Name)
	}
	ext := tax.Extern
	tax.Extern = nil
	for _, e := range ext {
		serv, id, err := jdh.ParseExtern(e)
		if err != nil {
			continue
		}
		if len(id) == 0 {
			continue
		}
		add := true
		for _, ex := range tax.Extern {
			if strings.HasPrefix(ex, serv) {
				add = false
				break
			}
		}
		if !add {
			continue
		}
		if _, ok := t.ids[e]; !ok {
			tax.Extern = append(tax.Extern, e)
		}
	}
	return nil
}

// IsDescValid check if a taxon with a given rank and status can be inserted
// safetly in the database hierarchy.
func (tx *taxon) isDescValid(rank jdh.Rank, valid bool) bool {
	if rank == 0 {
		return true
	}
	for p := tx; p.data != nil; p = p.parent {
		if p.data.Rank == 0 {
			continue
		}
		if (!valid) && (p.data.Rank <= rank) {
			return true
		}
		if p.data.Rank < rank {
			return true
		}
		return false
	}
	return true
}

// AddTaxon adds a new taxon to the database.
func (t *taxonomy) addTaxon(tax *jdh.Taxon) {
	p := t.root
	if len(tax.Parent) > 0 {
		p = t.ids[tax.Parent]
	}
	tx := &taxon{
		data:   tax,
		parent: p,
	}
	nmLow := strings.ToLower(tax.Name)
	v := t.names.Lookup(nmLow)
	var nm []*taxon
	if v == nil {
		nm = []*taxon{tx}
	} else {
		nm = append(v.([]*taxon), tx)
	}
	t.names.Set(nmLow, nm)
	t.ids[tax.Id] = tx
	for _, e := range tax.Extern {
		t.ids[e] = tx
	}
	p.childs = append(p.childs, tx)
}

// Add adds a new taxon to the database.
func (t *taxonomy) add(tax *jdh.Taxon) (string, error) {
	id := strconv.FormatInt(t.next, 10)
	tax.Id = id
	if err := t.validate(tax); err != nil {
		return "", err
	}
	t.addTaxon(tax)
	t.next++
	t.changed = true
	return id, nil
}

// Commit saves the taxonomy into hard disk.
func (t *taxonomy) commit(e chan error) {
	if !t.changed {
		e <- nil
		return
	}
	p := filepath.Join(t.db.path, taxFile)
	f, err := os.Create(p)
	if err != nil {
		e <- err
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, d := range t.root.childs {
		d.encode(enc)
	}
	t.changed = false
	e <- nil
}

// Encode encodes a taxon into a json blob in the database.
func (tx *taxon) encode(enc *json.Encoder) {
	enc.Encode(tx.data)
	for _, d := range tx.childs {
		d.encode(enc)
	}
}

// Delete deletes a taxon (and all its descendants) from the database.
func (t *taxonomy) delete(vals []jdh.KeyValue) error {
	id := ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			id = kv.Value[0]
			break
		}
	}
	if len(id) == 0 {
		return errors.New("taxon without identification")
	}
	tx, ok := t.ids[id]
	if !ok {
		return nil
	}
	t.delTaxon(tx)
	t.changed = true
	return nil
}

// DelTaxon recursively removes taxons from the database.
func (t *taxonomy) delTaxon(tx *taxon) {
	for len(tx.childs) > 0 {
		t.delTaxon(tx.childs[0])
	}
	// removes specimens
	if t.db.s != nil {
		t.db.s.delTaxon(tx.data.Id)
	}
	// removes rasters
	if t.db.rd != nil {
		t.db.rd.delTaxon(tx.data.Id)
	}
	tx.childs = nil
	nmLow := strings.ToLower(tx.data.Name)
	v := t.names.Lookup(nmLow).([]*taxon)
	v = delTaxFromList(v, tx)
	if len(v) > 0 {
		t.names.Set(nmLow, v)
	} else {
		t.names.Delete(nmLow)
	}
	tx.parent.childs = delTaxFromList(tx.parent.childs, tx)
	tx.parent = nil
	delete(t.ids, tx.data.Id)
	for _, e := range tx.data.Extern {
		delete(t.ids, e)
	}
	tx.data = nil
}

// DelTaxFromList removes a taxon pointer from a list of taxons
func delTaxFromList(ls []*taxon, tx *taxon) []*taxon {
	if len(ls) == 0 {
		return nil
	}
	if len(ls) == 1 {
		if ls[0] == tx {
			return nil
		}
		return ls
	}
	for i, ot := range ls {
		if ot == tx {
			copy(ls[i:], ls[i+1:])
			ls[len(ls)-1] = nil
			return ls[:len(ls)-1]
		}
	}
	return ls
}

// Get returns a taxon with a given id.
func (t *taxonomy) get(id string) (*jdh.Taxon, error) {
	if len(id) == 0 {
		return nil, errors.New("taxon without identification")
	}
	tx, ok := t.ids[id]
	if !ok {
		return nil, nil
	}
	return tx.data, nil
}

// List returns a list of taxons.
func (t *taxonomy) list(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	nameList := false
	noVal := true
	// creates the list
	for _, kv := range vals {
		switch kv.Key {
		case jdh.TaxChildren:
			tx := t.root
			if len(kv.Value) > 0 {
				v := strings.TrimSpace(kv.Value[0])
				if len(v) > 0 {
					var ok bool
					if tx, ok = t.ids[v]; !ok {
						return nil, fmt.Errorf("taxon %s not in database", kv.Value)
					}
				}
			}
			for _, d := range tx.childs {
				if d.data.IsValid {
					l.PushBack(d.data)
				}
			}
			noVal = false
		case jdh.TaxSynonyms:
			if len(kv.Value) == 0 {
				return l, nil
			}
			tx := t.root
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				return l, nil
			}
			var ok bool
			if tx, ok = t.ids[v]; !ok {
				return nil, fmt.Errorf("taxon %s not in database", kv.Value)
			}
			for _, d := range tx.childs {
				if !d.data.IsValid {
					l.PushBack(d.data)
				}
			}
			noVal = false
		case jdh.TaxParents:
			if len(kv.Value) == 0 {
				return l, nil
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				return l, nil
			}
			tx, ok := t.ids[v]
			if !ok {
				return nil, fmt.Errorf("taxon %s not in database", kv.Value)
			}
			for p := tx.parent; p != t.root; p = p.parent {
				l.PushBack(p.data)
			}
			noVal = false
		case jdh.TaxName:
			if len(kv.Value) == 0 {
				return nil, errors.New("taxon without identification")
			}
			nm := strings.ToLower(strings.Join(strings.Fields(kv.Value[0]), " "))
			if len(nm) == 0 {
				return nil, errors.New("taxon without identification")
			}
			i := strings.Index(nm, "*")
			if i == 0 {
				return nil, errors.New("taxon without identification")
			}
			var ls *list.List
			if i > 0 {
				prefix := nm[:i]
				ls = t.names.Prefix(prefix)
			} else {
				ls = list.New()
				if v := t.names.Lookup(nm); v != nil {
					ls.PushBack(v)
				}
			}
			for e := ls.Front(); e != nil; e = e.Next() {
				tl := e.Value.([]*taxon)
				for _, tx := range tl {
					l.PushBack(tx.data)
				}
			}
			noVal = false
			nameList = true
		}
		if !noVal {
			break
		}
	}
	if noVal {
		return nil, errors.New("taxon without identification")
	}
	if !nameList {
		return l, nil
	}

	// filters of the list.
	for _, kv := range vals {
		if l.Len() == 0 {
			break
		}
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.TaxParent:
			pId := strings.TrimSpace(kv.Value[0])
			if len(pId) == 0 {
				continue
			}
			for e := l.Front(); e != nil; {
				nx := e.Next()
				tax := e.Value.(*jdh.Taxon)
				if !t.isDesc(tax.Id, pId) {
					l.Remove(e)
				}
				e = nx
			}
		case jdh.TaxParentName:
			p := strings.Join(strings.Fields(kv.Value[0]), " ")
			if len(p) == 0 {
				continue
			}
			for e := l.Front(); e != nil; {
				nx := e.Next()
				tax := e.Value.(*jdh.Taxon)
				if !t.hasParentName(tax.Id, p) {
					l.Remove(e)
				}
				e = nx
			}
		case jdh.TaxRank:
			rank := jdh.GetRank(strings.TrimSpace(kv.Value[0]))
			for e := l.Front(); e != nil; {
				nx := e.Next()
				tax := e.Value.(*jdh.Taxon)
				if tax.Rank != rank {
					l.Remove(e)
				}
				e = nx
			}
		}
	}
	return l, nil
}

// IsDesc returns true if a taxon is descendant of parent in the taxonomy.
func (t *taxonomy) isDesc(id, parent string) bool {
	if (len(id) == 0) || (len(parent) == 0) {
		return false
	}
	tx, ok := t.ids[id]
	if !ok {
		return false
	}
	for p := tx.parent; p != t.root; p = p.parent {
		if p.data.Id == parent {
			return true
		}
	}
	return false
}

// HasParentName returns true if a taxon has a parent with a given name.
func (t *taxonomy) hasParentName(id, parent string) bool {
	if (len(id) == 0) || (len(parent) == 0) {
		return false
	}
	tx, ok := t.ids[id]
	if !ok {
		return false
	}
	for p := tx.parent; p.data != nil; p = p.parent {
		if p.data.Name == parent {
			return true
		}
	}
	return false
}

// Set sets a value of a taxon in the taxonomy.
func (t *taxonomy) set(vals []jdh.KeyValue) error {
	id := ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			id = kv.Value[0]
			break
		}
	}
	if len(id) == 0 {
		return errors.New("taxon without identification")
	}
	tx, ok := t.ids[id]
	if !ok {
		return nil
	}
	tax := tx.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if tax.Comment == v {
				continue
			}
			tax.Comment = v
		case jdh.KeyExtern:
			ok := false
			for _, v := range kv.Value {
				v := strings.TrimSpace(v)
				if len(v) == 0 {
					continue
				}
				serv, ext, err := jdh.ParseExtern(v)
				if err != nil {
					continue
				}
				if len(ext) == 0 {
					if !t.delExtern(tx, serv) {
						continue
					}
					ok = true
					continue
				}
				if t.addExtern(tx, v) != nil {
					continue
				}
				ok = true
			}
			if !ok {
				continue
			}
		case jdh.TaxAuthority:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if tax.Authority == v {
				continue
			}
			tax.Authority = v
		case jdh.TaxName:
			nm := ""
			if len(kv.Value) > 0 {
				nm = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if len(nm) == 0 {
				return fmt.Errorf("new name for %s undefined", tx.data.Name)
			}
			if tax.Name == nm {
				continue
			}
			// remove the old name
			nmLow := strings.ToLower(nm)
			v := t.names.Lookup(nmLow).([]*taxon)
			v = delTaxFromList(v, tx)
			if len(v) > 0 {
				t.names.Set(nmLow, v)
			} else {
				t.names.Delete(nmLow)
			}
			tax.Name = nm
			// add the new name
			nmLow = strings.ToLower(tax.Name)
			vi := t.names.Lookup(nmLow)
			var txLs []*taxon
			if vi == nil {
				txLs = []*taxon{tx}
			} else {
				txLs = append(vi.([]*taxon), tx)
			}
			t.names.Set(nmLow, txLs)
		case jdh.TaxParent:
			p := t.root
			if len(kv.Value) > 0 {
				v := strings.TrimSpace(kv.Value[0])
				if len(v) == 0 {
					continue
				}
				var ok bool
				p, ok = t.ids[v]
				if !ok {
					return fmt.Errorf("new parent [%s] for taxon %s not in database", kv.Value, tax.Name)
				}
			}
			if tx.parent == p {
				continue
			}
			if (!tax.IsValid) && (p == t.root) {
				return fmt.Errorf("taxon %s is a synonym, it requires a parent", tax.Name)
			}
			if (p != t.root) && (!p.data.IsValid) {
				return fmt.Errorf("new parent [%s] for taxon %s is a synonym", p.data.Name, tax.Name)
			}
			if !p.isDescValid(tax.Rank, tax.IsValid) {
				return fmt.Errorf("taxon %s rank incompatible with new parent [%s] hierarchy", tax.Name, p.data.Name)
			}
			tx.parent.childs = delTaxFromList(tx.parent.childs, tx)
			p.childs = append(p.childs, tx)
			tx.parent = p
			tax.Parent = p.data.Id
		case jdh.TaxRank:
			v := jdh.Unranked
			if len(kv.Value) > 0 {
				v = jdh.GetRank(strings.TrimSpace(kv.Value[0]))
			}
			if tax.Rank == v {
				continue
			}
			if !tx.parent.isDescValid(v, tax.IsValid) {
				return fmt.Errorf("new rank [%s] for taxon %s incompatible with taxonomy hierarchy", v, tax.Name)
			}
			tax.Rank = v
		case jdh.TaxSynonym:
			p := t.root
			if len(kv.Value) > 0 {
				v := strings.TrimSpace(kv.Value[0])
				if len(v) == 0 {
					continue
				}
				var ok bool
				p, ok = t.ids[v]
				if !ok {
					return fmt.Errorf("new parent [%s] for taxon %s not in database", kv.Value, tax.Name)
				}
			}
			if (p == tx.parent) && (!tax.IsValid) {
				continue
			}
			if p == t.root {
				return fmt.Errorf("taxon %s can not be a synonym: no new parent defined", tax.Name)
			}
			if p != tx.parent {
				if !p.isDescValid(tax.Rank, false) {
					return fmt.Errorf("taxon %s rank incompatible with new parent [%s] hierarchy", tax.Name, p.data.Name)
				}
				tx.parent.childs = delTaxFromList(tx.parent.childs, tx)
				tx.parent = p
				tax.Parent = p.data.Id
				p.childs = append(p.childs, tx)
			}
			for _, d := range tx.childs {
				d.parent = p
				d.data.Parent = p.data.Id
				p.childs = append(p.childs, d)
			}
			tx.childs = nil
			tax.IsValid = false
		case jdh.TaxValid:
			if tax.IsValid {
				continue
			}
			tx.parent.childs = delTaxFromList(tx.parent.childs, tx)
			tx.parent = tx.parent.parent
			tx.parent.childs = append(tx.parent.childs, tx)
			tax.IsValid = true
		default:
			continue
		}
		t.changed = true
	}
	return nil
}

// AddExtern adds an extern id to a taxon.
func (t *taxonomy) addExtern(tx *taxon, extern string) error {
	serv, id, err := jdh.ParseExtern(extern)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return nil
	}
	if ot, ok := t.ids[extern]; ok {
		return fmt.Errorf("extern id %s of %s alredy in use by %s", extern, tx.data.Name, ot.data.Name)
	}
	// the service is already assigned, then overwrite
	for i, e := range tx.data.Extern {
		if strings.HasPrefix(e, serv) {
			delete(t.ids, e)
			tx.data.Extern[i] = extern
			t.ids[extern] = tx
			return nil
		}
	}
	tx.data.Extern = append(tx.data.Extern, extern)
	t.ids[extern] = tx
	return nil
}

// DelExtern deletes an extern id of a taxon.
func (t *taxonomy) delExtern(tx *taxon, service string) bool {
	for i, e := range tx.data.Extern {
		if strings.HasPrefix(e, service) {
			delete(t.ids, e)
			copy(tx.data.Extern[i:], tx.data.Extern[i+1:])
			tx.data.Extern[len(tx.data.Extern)-1] = ""
			tx.data.Extern = tx.data.Extern[:len(tx.data.Extern)-1]
			return true
		}
	}
	return false
}

// IsInDB returns true if the taxon is the taxonomy.
func (t *taxonomy) isInDB(id string) bool {
	if len(id) == 0 {
		return false
	}
	if _, ok := t.ids[id]; ok {
		return true
	}
	return false
}
