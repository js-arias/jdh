// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package inat implements a jdh driver for i-Naturalist.
package inat

import (
	"encoding/xml"
	"errors"
	"net/http"
	"time"

	"github.com/js-arias/jdh/pkg/jdh"
)

const driver = "inat"

// DB implements the i-Naturalist connection jdh DB interface.
type DB struct {
	isClosed bool
	request  chan string
	answer   chan interface{}
}

func init() {
	jdh.Register(driver, open)
}

// open creates a new database.
func open(port string) (jdh.DB, error) {
	db := &DB{
		isClosed: false,
		request:  make(chan string),
		answer:   make(chan interface{}),
	}
	go db.req()
	return db, nil
}

// Close closes the database.
func (db *DB) Close() error {
	if db.isClosed {
		return errors.New("database already closed")
	}
	close(db.request)
	return nil
}

// Driver returns the driver name.
func (db *DB) Driver() string {
	return driver
}

// Executable query can not be done in inat: it is a read only database.
func (db *DB) Exec(query jdh.Query, table jdh.Table, param interface{}) (string, error) {
	return "", errors.New("ncbi is a read only database")
}

// Get returns an element data from inat database.
func (db *DB) Get(table jdh.Table, id string) (jdh.Scanner, error) {
	if db.isClosed {
		return nil, errors.New("database already closed")
	}
	switch table {
	case jdh.Taxonomy:
		return db.taxon(id)
	}
	return nil, errors.New("get not implemented for table " + string(table))
}

// List executes a query that returns a list.
func (db *DB) List(table jdh.Table, args *jdh.Values) (jdh.ListScanner, error) {
	if db.isClosed {
		return nil, errors.New("database already closed")
	}
	if args == nil {
		return nil, errors.New("empty argument list")
	}
	switch table {
	case jdh.Taxonomy:
		return db.taxonList(args.KV)
	}
	return nil, errors.New("list not implemented for table " + string(table))
}

const inatHead = "http://www.inaturalist.org/"

// process requests
func (db *DB) req() {
	for r := range db.request {
		answer, err := http.Get(r)
		if err != nil {
			db.answer <- err
			continue
		}
		db.answer <- answer

		// this is set to not overload the gbif server...
		// I'm afraid of being baned!
		time.Sleep(100 * time.Millisecond)
	}
}

func skip(dec *xml.Decoder, end string) error {
	for tk, err := dec.Token(); ; tk, err = dec.Token() {
		if err != nil {
			return err
		}
		switch t := tk.(type) {
		case xml.EndElement:
			if t.Name.Local == end {
				return nil
			}
		}
	}
}
