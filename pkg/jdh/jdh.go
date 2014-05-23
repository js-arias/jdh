// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

// Package jdh defines the database used by a jdh client application.
//
// This package provides the general database interface, as well as
// associated types for retrieval, insertion or modification of the
// database.
//
// In client side code, jdh package must be used in conjunction with a
// database driver.
package jdh

import (
	"errors"
	"fmt"
	"strings"
)

// Driver holds the information of a database driver.
type driver struct {
	name string
	open func(param string) (DB, error)
}

// Drivers holds the database drivers.
var drivers []driver

// Register makes a database driver available by the provided name. If the
// name is black, or registred twice, it panics.
func Register(name string, open func(param string) (DB, error)) {
	if open == nil {
		panic("database: open function is nil")
	}
	if len(name) == 0 {
		panic("database: empty driver name")
	}
	for _, d := range drivers {
		if d.name == name {
			panic("database: Register called twice for driver " + name)
		}
	}
	drivers = append(drivers, driver{name, open})
}

// DB is the database handle to be used by a client application.
//
// This interface should be implemented by a valid jdh database driver.
type DB interface {
	// Close closes the database.
	Close() error

	// Driver returns the name of the driver that underlies the database.
	Driver() string

	// Exec executes a query without returning any row (e.g. Add, Delete),
	// usually this executions require modification of the database data.
	// It returns a string result (that depends on the query), or an
	// error, if the query fail.
	Exec(query Query, table Table, param interface{}) (string, error)

	// Get executes a query that is expected to return at most a single
	// value.
	Get(table Table, id string) (Scanner, error)

	// List executes a query that returns a list.
	List(table Table, args *Values) (ListScanner, error)
}

// Open opens a database by its driver.If driver is not present, it returns
// an error.
func Open(name, par string) (DB, error) {
	for _, d := range drivers {
		if d.name == name {
			return d.open(par)
		}
	}
	return nil, fmt.Errorf("driver %s unregistred", name)
}

// Scanner is the result of a get query. It can be read only one time.
type Scanner interface {
	Scan(dest interface{}) error
}

// ListScanner is the result of a list query. It can be read until no more
// elements are in the answer. If the list is abandoned before any error of
// the last element was scanned, it should be closed by the caller.
type ListScanner interface {
	Scanner
	Close()
}

// Query is a query to the database.
type Query string

// Available queries.
const (
	// Add request the addition of an element to the database. The param
	// parameter is the element to be added.
	Add Query = "add"

	// Commit requests the commit of the database.
	Commit = "commit"

	// Close request the closing of the database.
	Close = "close"

	// Delete requests the removal of an element from the database. The
	// param argument is a Values variable that must be include the id
	// of the element to be deleted (all other values will be ignored).
	Delete = "delete"

	// Get request an element form the database.
	Get = "get"

	// List request a list of elements from the database.
	List = "list"

	// Set requests the setting of a value in the database. The param
	// argument is a Values variable that must be included the id, all
	// other values will be modified if they are valid.
	Set = "set"
)

// Table is a request parameter that indicate the "table" in whinch an
// operation should be performed.
type Table string

// KeyValue is a key:value pair used for jdh database queries. Instead of
// this structure most code should use values.
type KeyValue struct {
	Key   Key
	Value []string
}

// Values kepts a list of key:value pairs used for jdh database queries.
// Key values are case-sensitive, and each key can only hold a unique
// value.
type Values struct {
	KV []KeyValue
}

// Add adds a new key:value pair. If the key is already assigned it will
// append the value.
func (v *Values) Add(key Key, value string) {
	key = Key(strings.TrimSpace(string(key)))
	if len(key) == 0 {
		return
	}
	value = strings.Join(strings.Fields(value), " ")
	for i, kv := range v.KV {
		if kv.Key == key {
			if len(value) > 0 {
				v.KV[i].Value = append(v.KV[i].Value, value)
			}
			return
		}
	}
	if len(value) > 0 {
		v.KV = append(v.KV, KeyValue{Key: key, Value: []string{value}})
	} else {
		v.KV = append(v.KV, KeyValue{Key: key})
	}
}

// Set sets a key:value pair. If the key is already assigned, it will deletes
// the old content.
func (v *Values) Set(key Key, value string) {
	key = Key(strings.TrimSpace(string(key)))
	if len(key) == 0 {
		return
	}
	value = strings.Join(strings.Fields(value), " ")
	for i, kv := range v.KV {
		if kv.Key == key {
			if len(value) > 0 {
				v.KV[i].Value = []string{value}
			} else {
				v.KV[i].Value = nil
			}
			return
		}
	}
	if len(value) > 0 {
		v.KV = append(v.KV, KeyValue{Key: key, Value: []string{value}})
	} else {
		v.KV = append(v.KV, KeyValue{Key: key})
	}
}

// Reset cleans the key-values stored in values.
func (v *Values) Reset() {
	v.KV = nil
}

// Key is a key used to query a particular element o field in the jdh
// database.
type Key string

// General Keys used in all or most table's elements.
const (
	// Identifier of an element in a table. Valid values are id and
	// extern ids of the element.
	KeyId Key = "id"

	// Comment field of an element. All free text values are valid.
	// An empty value during a set operation will delete the comment.
	KeyComment = "comment"

	// Extern field of a method. The value is in the form <serv>:<id>,
	// e.g. "gbif:5216933". An empty <id> during a set operation will
	// delete the extern identifier.
	KeyExtern = "extern"

	// A bibliographic reference.
	KeyReference = "reference"
)

// ParseExtern parses an extern identifier. Extern identifiers are of the
// form <service>:<key>, for example "gbif:2423087". It returns the service,
// the key, an error if the extern id is malformed.
//
// To avoid matches like "db" with "dbx", the service is returned including
// the colon, so in this example, it will return unambiguous "db:" or "dbx:".
func ParseExtern(extern string) (string, string, error) {
	p := strings.Index(extern, ":")
	if p < 1 {
		return "", "", errors.New("invalid extern identifier")
	}
	return extern[:p+1], extern[p+1:], nil
}

// ISO 8601 layout for time, as in <http://en.wikipedia.org/wiki/ISO_8601>
const Iso8601 = "2006-01-02T15:04:05+07:00"
