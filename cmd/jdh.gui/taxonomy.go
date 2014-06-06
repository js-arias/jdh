// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

// Taxon gets a taxon.
func taxon(c *cmdapp.Command, db jdh.DB, id string) *jdh.Taxon {
	sc, err := db.Get(jdh.Taxonomy, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	tax := &jdh.Taxon{}
	if err := sc.Scan(tax); err != nil {
		if err == io.EOF {
			return &jdh.Taxon{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return tax
}

type txList struct {
	db   jdh.DB
	tax  *jdh.Taxon
	desc []*jdh.Taxon
	sels []int
	syns bool
}

func newTxList(tax *jdh.Taxon, db jdh.DB, syns bool) *txList {
	ls := &txList{
		db:   db,
		tax:  tax,
		syns: syns,
	}
	id := ""
	if tax == nil {
		ls.tax = &jdh.Taxon{
			Id:   "0",
			Name: "root",
		}
	} else {
		id = tax.Id
	}
	vals := new(jdh.Values)
	vals.Add(jdh.TaxChildren, id)
	pl, err := db.List(jdh.Taxonomy, vals)
	if err != nil {
		return ls
	}
	for {
		d := &jdh.Taxon{}
		if err := pl.Scan(d); err != nil {
			break
		}
		ls.desc = append(ls.desc, d)
	}
	if !syns {
		return ls
	}
	vals.Reset()
	vals.Add(jdh.TaxSynonyms, id)
	pl, err = db.List(jdh.Taxonomy, vals)
	if err != nil {
		return ls
	}
	for {
		s := &jdh.Taxon{}
		if err := pl.Scan(s); err != nil {
			break
		}
		ls.desc = append(ls.desc, s)
	}
	return ls
}

func (ls *txList) Len() int {
	return len(ls.desc)
}

func (ls *txList) Item(i int) string {
	nm := ls.desc[i].Name
	if !ls.desc[i].IsValid {
		nm = "[syn] " + nm
	}
	return nm
}

func (ls *txList) IsSel(i int) bool {
	for _, s := range ls.sels {
		if s == i {
			return true
		}
	}
	return false
}
