// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

// Package native implements the driver of native jdh databases.
package native

import (
	"encoding/json"
	"errors"
	"net"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

const driver = "native"

func init() {
	jdh.Register(driver, open)
}

// Port is the default jdh port.
const Port = ":16917"

// A Request is a database request.
type Request struct {
	Query jdh.Query
	Table jdh.Table
	Kvs   []jdh.KeyValue
}

// An Answer is an answer from the database. If the Message field is
// different of "success", an error has occurred, and the Message contains
// further details about the error.
type Answer struct {
	Message string // answer mensage
}

// GetMessage returns the message from a database Answer.
func (a *Answer) GetMessage() (string, error) {
	msg := strings.Split(a.Message, ": ")
	if msg[0] != "success" {
		if len(msg) == 1 {
			return "", errors.New("unknown server error")
		}
		return "", errors.New(strings.Join(msg[1:], ":"))
	}
	m := ""
	if len(msg) > 1 {
		m = strings.TrimSpace(strings.Join(msg[1:], ":"))
	}
	return m, nil
}

// Success returns a success answer.
func Success(msg string) *Answer {
	return &Answer{Message: "success: " + msg}
}

// ErrAnswer retruns an error answer.
func ErrAnswer(msg string) *Answer {
	return &Answer{Message: "error: " + msg}
}

// DB holds the information of the native database.
type DB struct {
	port string
}

// Open creates a new database connection.
func open(port string) (jdh.DB, error) {
	if len(port) == 0 {
		port = Port
	}
	if i := strings.Index(port, ":"); i == 0 {
		port = "localhost" + port
	} else if i < 0 {
		port = "localhost:" + port
	}
	return &DB{port}, nil
}

// Close closes the database.
func (db *DB) Close() error {
	return db.simple(jdh.Close)
}

// Driver returns the driver name.
func (db *DB) Driver() string {
	return driver
}

// Exec executes a query on the database.
func (db *DB) Exec(query jdh.Query, table jdh.Table, param interface{}) (string, error) {
	switch query {
	case jdh.Add:
		if param == nil {
			return "", errors.New("empty element")
		}
		conn, err := net.Dial("tcp", db.port)
		if err != nil {
			return "", err
		}
		defer conn.Close()
		enc := json.NewEncoder(conn)
		req := &Request{
			Query: jdh.Add,
			Table: table,
		}
		enc.Encode(req)
		enc.Encode(param)
		dec := json.NewDecoder(conn)
		ans := &Answer{}
		if err := dec.Decode(ans); err != nil {
			return "", err
		}
		if id, err := ans.GetMessage(); err != nil {
			return "", err
		} else {
			return id, nil
		}
		return "", nil
	case jdh.Commit:
		return "", db.simple(query)
	case jdh.Delete:
		if param == nil {
			return "", errors.New("empty argument list")
		}
		kvs := param.(*jdh.Values)
		if len(kvs.KV) == 0 {
			return "", errors.New("empty argument list")
		}
		conn, err := net.Dial("tcp", db.port)
		if err != nil {
			return "", err
		}
		defer conn.Close()
		enc := json.NewEncoder(conn)
		req := &Request{
			Query: jdh.Delete,
			Table: table,
			Kvs:   kvs.KV,
		}
		enc.Encode(req)
		dec := json.NewDecoder(conn)
		ans := &Answer{}
		if err := dec.Decode(ans); err != nil {
			return "", err
		}
		if _, err := ans.GetMessage(); err != nil {
			return "", err
		}
		return "", nil
	case jdh.Set:
		if param == nil {
			return "", errors.New("empty argument list")
		}
		kvs := param.(*jdh.Values)
		if len(kvs.KV) == 0 {
			return "", errors.New("empty argument list")
		}
		conn, err := net.Dial("tcp", db.port)
		if err != nil {
			return "", err
		}
		defer conn.Close()
		enc := json.NewEncoder(conn)
		req := &Request{
			Query: jdh.Set,
			Table: table,
			Kvs:   kvs.KV,
		}
		enc.Encode(req)
		dec := json.NewDecoder(conn)
		ans := &Answer{}
		if err := dec.Decode(ans); err != nil {
			return "", err
		}
		if _, err := ans.GetMessage(); err != nil {
			return "", err
		}
		return "", nil
	}
	return "", errors.New("invalid query")
}

// Get request a single element from the database.
func (db *DB) Get(table jdh.Table, id string) (jdh.Scanner, error) {
	conn, err := net.Dial("tcp", db.port)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(conn)
	req := &Request{
		Query: jdh.Get,
		Table: table,
		Kvs:   []jdh.KeyValue{jdh.KeyValue{Key: jdh.KeyId, Value: []string{id}}},
	}
	enc.Encode(req)
	dec := json.NewDecoder(conn)
	ans := &Answer{}
	if err := dec.Decode(ans); err != nil {
		return nil, err
	}
	if _, err := ans.GetMessage(); err != nil {
		return nil, err
	}
	return &getScanner{c: conn, d: dec}, nil
}

// List executes a query that returns a list.
func (db *DB) List(table jdh.Table, args *jdh.Values) (jdh.ListScanner, error) {
	if args == nil {
		return nil, errors.New("empty argument list")
	}
	conn, err := net.Dial("tcp", db.port)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(conn)
	req := &Request{
		Query: jdh.List,
		Table: table,
		Kvs:   args.KV,
	}
	enc.Encode(req)
	dec := json.NewDecoder(conn)
	ans := &Answer{}
	if err := dec.Decode(ans); err != nil {
		return nil, err
	}
	if _, err := ans.GetMessage(); err != nil {
		return nil, err
	}
	return &listScanner{c: conn, d: dec}, nil
}

// sends a simple request
func (db *DB) simple(query jdh.Query) error {
	conn, err := net.Dial("tcp", db.port)
	if err != nil {
		return err
	}
	defer conn.Close()
	enc := json.NewEncoder(conn)
	req := &Request{Query: query}
	enc.Encode(req)
	dec := json.NewDecoder(conn)
	ans := &Answer{}
	if err := dec.Decode(ans); err != nil {
		return err
	}
	if _, err := ans.GetMessage(); err != nil {
		return err
	}
	return nil
}
