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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

// Datasets holds the information of extern datasets.
type datasets struct {
	db      *DB
	ids     map[string]*setData // map of id:dataset
	ls      *list.List          // list of datasets
	changed bool                // if true, the database has changed
	next    uint64              // next valid id
}

// SetData holds the dataset information.
type setData struct {
	data *jdh.Dataset
	elem *list.Element
}

// dataset file
const dsetFile = "datasets"

// OpenDatasets open dataset data.
func openDatasets(db *DB) *datasets {
	d := &datasets{
		db:   db,
		ids:  make(map[string]*setData),
		ls:   list.New(),
		next: 1,
	}
	p := filepath.Join(db.path, dsetFile)
	f, err := os.Open(p)
	if err != nil {
		return d
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		set := &jdh.Dataset{}
		if err := dec.Decode(set); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-datasets: error: %v\n", err)
			break
		}
		d.setNext(set.Id)
		if err := d.validate(set); err != nil {
			log.Printf("db-datasets: error: %v\n", err)
			continue
		}
		d.addSet(set)
	}
	return d
}

// SetNext sets the value of the next id,
func (d *datasets) setNext(id string) {
	v, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return
	}
	if v >= d.next {
		d.next = v + 1
	}
}

// Validates that a dataset is valid in the database. When posible canonical
// values will be set. If the dataset is invalid it will return an error.
func (d *datasets) validate(set *jdh.Dataset) error {
	set.Id = strings.TrimSpace(set.Id)
	set.Title = strings.Join(strings.Fields(set.Title), " ")
	if (len(set.Id) == 0) || (len(set.Title) == 0) {
		return errors.New("dataset without identification")
	}
	if _, ok := d.ids[set.Id]; ok {
		return fmt.Errorf("dataset id %s already in use", set.Id)
	}
	set.Citation = strings.Join(strings.Fields(set.Citation), " ")
	set.License = strings.Join(strings.Fields(set.License), " ")
	if u, err := url.Parse(set.Url); err != nil {
		set.Url = ""
	} else {
		set.Url = u.String()
	}
	ext := set.Extern
	set.Extern = nil
	for _, e := range ext {
		serv, id, err := jdh.ParseExtern(e)
		if err != nil {
			continue
		}
		if len(id) == 0 {
			continue
		}
		add := true
		for _, ex := range set.Extern {
			if strings.HasPrefix(ex, serv) {
				add = false
				break
			}
		}
		if !add {
			continue
		}
		if _, ok := d.ids[e]; !ok {
			set.Extern = append(set.Extern, e)
		}
	}
	return nil
}

// AddSet adds a new dataset to the database.
func (d *datasets) addSet(set *jdh.Dataset) {
	sd := &setData{
		data: set,
	}
	sd.elem = d.ls.PushBack(sd)
	d.ids[set.Id] = sd
	for _, e := range set.Extern {
		d.ids[e] = sd
	}
}

// Add adds a new dataset to the database.
func (d *datasets) add(set *jdh.Dataset) (string, error) {
	id := strconv.FormatUint(d.next, 10)
	set.Id = id
	if err := d.validate(set); err != nil {
		return "", err
	}
	d.addSet(set)
	d.next++
	d.changed = true
	return id, nil
}

// Commit saves the datasets to the hard disk.
func (d *datasets) commit(e chan error) {
	if !d.changed {
		e <- nil
		return
	}
	p := filepath.Join(d.db.path, dsetFile)
	f, err := os.Create(p)
	if err != nil {
		e <- err
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for e := d.ls.Front(); e != nil; e = e.Next() {
		sd := e.Value.(*setData)
		enc.Encode(sd.data)
	}
	d.changed = false
	e <- nil
}

// Delete deletes a collection from the database.
func (d *datasets) delete(vals []jdh.KeyValue) error {
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
		return errors.New("dataset without identification")
	}
	sd, ok := d.ids[id]
	if !ok {
		return nil
	}
	delete(d.ids, sd.data.Id)
	for _, e := range sd.data.Extern {
		delete(d.ids, e)
	}
	d.ls.Remove(sd.elem)
	sd.elem = nil
	sd.data = nil
	d.changed = true
	return nil
}

// Get returns a dataset with a given id.
func (d *datasets) get(id string) (*jdh.Dataset, error) {
	if len(id) == 0 {
		return nil, errors.New("dataset without identification")
	}
	sd, ok := d.ids[id]
	if !ok {
		return nil, nil
	}
	return sd.data, nil
}

// List returns a list of datasets.
func (d *datasets) list(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	// creates the list
	for e := d.ls.Front(); e != nil; e = e.Next() {
		sd := e.Value.(*setData)
		l.PushBack(sd.data)
	}

	// filter the list
	for _, kv := range vals {
		if l.Len() == 0 {
			break
		}
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.DataLicense:
			for e := l.Front(); e != nil; {
				nx := e.Next()
				set := e.Value.(*jdh.Dataset)
				remove := true
				for _, v := range kv.Value {
					lc := strings.Join(strings.Fields(v), " ")
					if len(lc) == 0 {
						continue
					}
					if i := strings.Index(lc, "*"); i > 0 {
						lc = lc[:i]
					}
					if strings.HasPrefix(set.License, lc) {
						remove = false
						break
					}
				}
				if remove {
					l.Remove(e)
				}
				e = nx
			}
		case jdh.DataTitle:
			for e := l.Front(); e != nil; {
				nx := e.Next()
				set := e.Value.(*jdh.Dataset)
				remove := true
				for _, v := range kv.Value {
					tl := strings.Join(strings.Fields(v), " ")
					if len(tl) == 0 {
						continue
					}
					if i := strings.Index(tl, "*"); i > 0 {
						tl = tl[:i]
					}
					if strings.HasPrefix(set.Title, tl) {
						remove = false
						break
					}
				}
				if remove {
					l.Remove(e)
				}
				e = nx
			}
		}
	}
	return l, nil
}

// Set sets one or more values of a dataset element.
func (d *datasets) set(vals []jdh.KeyValue) error {
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
		return errors.New("dataset without identification")
	}
	sd, ok := d.ids[id]
	if !ok {
		return nil
	}
	set := sd.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if set.Comment == v {
				continue
			}
			set.Comment = v
		case jdh.KeyExtern:
			ok := false
			for _, v := range kv.Value {
				v = strings.TrimSpace(v)
				if len(v) == 0 {
					continue
				}
				serv, ext, err := jdh.ParseExtern(v)
				if err != nil {
					continue
				}
				if len(ext) == 0 {
					if !d.delExtern(sd, serv) {
						continue
					}
					ok = true
					continue
				}
				if d.addExtern(sd, v) != nil {
					continue
				}
				ok = true
			}
			if !ok {
				continue
			}
		case jdh.DataCitation:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if set.Citation == v {
				continue
			}
			set.Citation = v
		case jdh.DataLicense:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if set.License == v {
				continue
			}
			set.License = v
		case jdh.DataTitle:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if set.Title == v {
				continue
			}
			set.Title = v
		case jdh.DataUrl:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if len(v) == 0 {
				if len(set.Url) == 0 {
					continue
				}
				set.Url = ""
				break
			}
			u, err := url.Parse(v)
			if err != nil {
				continue
			}
			if set.Url == u.String() {
				continue
			}
			set.Url = u.String()
		}
		d.changed = true
	}
	return nil
}

// AddExtern adds an extern id to a dataset.
func (d *datasets) addExtern(sd *setData, extern string) error {
	serv, id, err := jdh.ParseExtern(extern)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return nil
	}
	if ot, ok := d.ids[extern]; ok {
		return fmt.Errorf("extern id %s of %s alredy in use by %s", extern, sd.data.Id, ot.data.Id)
	}
	// the service is already assigned, then overwrite
	for i, e := range sd.data.Extern {
		if strings.HasPrefix(e, serv) {
			delete(d.ids, e)
			sd.data.Extern[i] = extern
			d.ids[extern] = sd
			return nil
		}
	}
	sd.data.Extern = append(sd.data.Extern, extern)
	d.ids[extern] = sd
	return nil
}

// DelExtern deletes an extern id of a dataset.
func (d *datasets) delExtern(sd *setData, service string) bool {
	for i, e := range sd.data.Extern {
		if strings.HasPrefix(e, service) {
			delete(d.ids, e)
			copy(sd.data.Extern[i:], sd.data.Extern[i+1:])
			sd.data.Extern[len(sd.data.Extern)-1] = ""
			sd.data.Extern = sd.data.Extern[:len(sd.data.Extern)-1]
			return true
		}
	}
	return false
}

// IsInDB returns true if the dataset is the database.
func (d *datasets) isInDB(id string) bool {
	if len(id) == 0 {
		return false
	}
	if _, ok := d.ids[id]; ok {
		return true
	}
	return false
}
