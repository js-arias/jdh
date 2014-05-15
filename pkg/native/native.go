// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package native implements the jdh native database.
package native

import (
	"container/list"
	"encoding/json"
	"errors"
	"sync"

	"github.com/js-arias/jdh/pkg/jdh"
)

// DB holds the database information.
type DB struct {
	path string // path of the database

	//database tables
	d *datasets
	t *taxonomy
	s *specimens

	lock sync.Mutex
}

// Open opens a database in a given path.
func Open(path string) *DB {
	db := &DB{path: path}
	db.d = openDatasets(db)
	db.t = openTaxonomy(db)
	db.s = openSpecimens(db)
	return db
}

// Add adds a new element to the database.
func (db *DB) Add(table jdh.Table, dec *json.Decoder) (string, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
	case jdh.Datasets:
		set := &jdh.Dataset{}
		if err := dec.Decode(set); err != nil {
			return "", err
		}
		return db.d.add(set)
	case jdh.Specimens:
		spe := &jdh.Specimen{}
		if err := dec.Decode(spe); err != nil {
			return "", err
		}
		return db.s.add(spe)
	case jdh.Taxonomy:
		tax := &jdh.Taxon{}
		if err := dec.Decode(tax); err != nil {
			return "", err
		}
		return db.t.add(tax)
	}
	return "", errors.New("add not implemented for table " + string(table))
}

// commiter is a type that commits its data.
type commiter interface {
	commit(chan error)
}

// DoCommit commit with a commiter.
func doCommit(c commiter, done *sync.WaitGroup, e chan error) {
	done.Add(1)
	go func() {
		c.commit(e)
		done.Done()
	}()
}

// Commit commits the database.
func (db *DB) Commit() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	ec := make(chan error)
	go func() {
		var done sync.WaitGroup
		doCommit(db.d, &done, ec)
		doCommit(db.t, &done, ec)
		doCommit(db.s, &done, ec)
		done.Wait()
		close(ec)
	}()
	var err error
	for e := range ec {
		if (e != nil) && (err == nil) {
			err = e
		}
	}
	return err
}

// Delete removes an element from the database.
func (db *DB) Delete(table jdh.Table, vals []jdh.KeyValue) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
	case jdh.Datasets:
		return db.d.delete(vals)
	case jdh.Specimens:
		return db.s.delete(vals)
	case jdh.Taxonomy:
		return db.t.delete(vals)
	}
	return errors.New("delete not implemented for table " + string(table))
}

// Get returns an element from the database.
func (db *DB) Get(table jdh.Table, id string) (interface{}, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
	case jdh.Datasets:
		return db.d.get(id)
	case jdh.Specimens:
		return db.s.get(id)
	case jdh.Taxonomy:
		return db.t.get(id)
	}
	return nil, errors.New("get not implemented for table " + string(table))
}

// List returns a list of elements from the database.
func (db *DB) List(table jdh.Table, vals []jdh.KeyValue) (*list.List, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
	case jdh.Datasets:
		return db.d.list(vals)
	case jdh.Specimens:
		return db.s.list(vals)
	case jdh.Taxonomy:
		return db.t.list(vals)
	}
	return nil, errors.New("list not implemented for table " + string(table))
}

// Set sets one or more values for a given element in the database.
func (db *DB) Set(table jdh.Table, vals []jdh.KeyValue) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
	case jdh.Datasets:
		return db.d.set(vals)
	case jdh.Specimens:
		return db.s.set(vals)
	case jdh.Taxonomy:
		return db.t.set(vals)
	}
	return errors.New("set not implemented for table " + string(table))
}
