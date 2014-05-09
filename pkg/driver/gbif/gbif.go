// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package gbif implements a jdh driver for gbif.
package gbif

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/js-arias/jdh/pkg/jdh"
)

const driver = "gbif"

// DB implements the GBIF connection jdh DB interface.
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

// Close always produce an error (DB interface).
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

// Executable query can not be done in gbif: it is a read only database.
func (db *DB) Exec(query jdh.Query, table jdh.Table, param interface{}) (string, error) {
	return "", errors.New("gbif is a read only database")
}

// Get returns an element data from gbif database.
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
	if len(args.KV) == 0 {
		return nil, errors.New("empty argument list")
	}
	switch table {
	case jdh.Taxonomy:
		return db.taxonList(args.KV)
	}
	return nil, errors.New("list not implemented for table " + string(table))
}

const wsHead = "http://api.gbif.org/v0.9/"

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

func (db *DB) listRequest(request string, an interface{}) error {
	db.request <- request
	a := <-db.answer
	switch answer := a.(type) {
	case error:
		return answer
	case *http.Response:
		defer answer.Body.Close()
		d := json.NewDecoder(answer.Body)
		if err := d.Decode(an); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) simpleRequest(request string, an interface{}) error {
	db.request <- request
	a := <-db.answer
	switch answer := a.(type) {
	case error:
		return answer
	case *http.Response:
		defer answer.Body.Close()
		d := json.NewDecoder(answer.Body)
		if err := d.Decode(an); err != nil {
			return err
		}
	}
	return nil
}
