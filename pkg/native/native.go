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
	t *taxonomy

	lock sync.Mutex
}

// Open opens a database in a given path.
func Open(path string) *DB {
	db := &DB{path: path}
	db.t = openTaxonomy(db)
	return db
}

// Add adds a new element to the database.
func (db *DB) Add(table jdh.Table, dec *json.Decoder) (string, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	switch table {
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
		doCommit(db.t, &done, ec)
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
	case jdh.Taxonomy:
		return db.t.set(vals)
	}
	return errors.New("set not implemented for table " + string(table))
}
