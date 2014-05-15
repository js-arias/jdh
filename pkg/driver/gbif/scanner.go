// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package gbif

import (
	"io"

	"github.com/js-arias/jdh/pkg/jdh"
)

// GetScanner scans a single value.
type getScanner struct {
	val interface{}
	err error
}

func (g *getScanner) Scan(dest interface{}) error {
	if g.err != nil {
		return g.err
	}
	switch v := dest.(type) {
	case *jdh.Dataset:
		*v = *g.val.(*jdh.Dataset)
	case *jdh.Specimen:
		*v = *g.val.(*jdh.Specimen)
	case *jdh.Taxon:
		*v = *g.val.(*jdh.Taxon)
	}
	g.err = io.EOF
	return nil
}

// ListScanner scans a list of values.
type listScanner struct {
	c   chan interface{}
	end chan struct{}
	err error
}

func (l *listScanner) Scan(dest interface{}) error {
	if l.err != nil {
		return l.err
	}
	var val interface{}
	select {
	case <-l.end:
		return l.err
	case val = <-l.c:
		if val == nil {
			l.Close()
			return io.EOF
		}
		switch v := dest.(type) {
		case *jdh.Dataset:
			*v = *val.(*jdh.Dataset)
		case *jdh.Specimen:
			*v = *val.(*jdh.Specimen)
		case *jdh.Taxon:
			*v = *val.(*jdh.Taxon)
		}
	}
	return nil
}

func (l *listScanner) Close() {
	if l.err != nil {
		return
	}
	close(l.end)
	l.err = io.EOF
}

func (l *listScanner) setErr(err error) {
	if l.err != nil {
		return
	}
	close(l.end)
	l.err = err
}
