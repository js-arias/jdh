// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package ncbi implements a jdh driver for genbank using ncbi and embl.
package ncbi

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/js-arias/jdh/pkg/jdh"
)

const driver = "ncbi"

// DB implements the NCBI connection jdh DB interface.
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
	return "", errors.New("ncbi is a read only database")
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

const (
	ncbiHead = "http://eutils.ncbi.nlm.nih.gov/entrez/eutils/"
	emblHead = "http://www.ebi.ac.uk/ena/data/view/"
)

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

func (db *DB) search(request string) ([]string, int, error) {
	db.request <- request
	a := <-db.answer
	var ids []string
	switch answer := a.(type) {
	case error:
		return nil, 0, answer
	case *http.Response:
		defer answer.Body.Close()
		dec := xml.NewDecoder(answer.Body)
		prev := ""
		var start, count, max int64
		for tk, err := dec.Token(); err != io.EOF; tk, err = dec.Token() {
			if err != nil {
				return nil, 0, err
			}
			switch t := tk.(type) {
			case xml.StartElement:
				prev = t.Name.Local
			case xml.EndElement:
				prev = ""
			case xml.CharData:
				switch prev {
				case "Id":
					ids = append(ids, string(t))
				case "Count":
					count, err = strconv.ParseInt(string(t), 10, 0)
					if err != nil {
						return nil, 0, err
					}
				case "RetMax":
					max, err = strconv.ParseInt(string(t), 10, 0)
					if err != nil {
						return nil, 0, err
					}
				case "RetStart":
					start, err = strconv.ParseInt(string(t), 10, 0)
					if err != nil {
						return nil, 0, err
					}
				}
			}
		}
		if next := start + max; next < count {
			return ids, int(next), nil
		}
	}
	return ids, 0, nil
}
