// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package native

import (
	"encoding/json"
	"io"
	"net"
)

// GetScanner scans a single value.
type getScanner struct {
	c   net.Conn
	d   *json.Decoder
	err error
}

func (g *getScanner) Scan(dest interface{}) error {
	if g.err != nil {
		return g.err
	}
	defer g.c.Close()
	if err := g.d.Decode(dest); err != nil {
		g.err = err
		return err
	}
	g.err = io.EOF
	return nil
}

// ListScanner scans a list of values.
type listScanner struct {
	c   net.Conn
	d   *json.Decoder
	err error
}

func (l *listScanner) Scan(dest interface{}) error {
	if l.err != nil {
		return l.err
	}
	if err := l.d.Decode(dest); err != nil {
		l.err = err
		l.c.Close()
		return err
	}
	return nil
}

func (l *listScanner) Close() {
	if l.err != nil {
		return
	}
	l.c.Close()
	l.err = io.EOF
}
