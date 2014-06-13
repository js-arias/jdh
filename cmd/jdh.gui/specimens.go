// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

// SpeList returns an specimen list scanner.
func speList(c *cmdapp.Command, db jdh.DB, vals *jdh.Values) jdh.ListScanner {
	l, err := db.List(jdh.Specimens, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return l
}

type spInfo struct {
	tax *jdh.Taxon
	spe *jdh.Specimen
	set *jdh.Dataset
}

type spList struct {
	db  jdh.DB
	tax *jdh.Taxon
	spe []*jdh.Specimen
	sel int
}

func newSpList(tax *jdh.Taxon, db jdh.DB) *spList {
	ls := &spList{
		db:  db,
		sel: -1,
		tax: tax,
	}
	if taxonRank(cmd, db, tax) < jdh.Species {
		return ls
	}
	vals := new(jdh.Values)
	vals.Add(jdh.SpeTaxon, tax.Id)
	l, err := db.List(jdh.Specimens, vals)
	if err != nil {
		return ls
	}
	for {
		spe := &jdh.Specimen{}
		if err := l.Scan(spe); err != nil {
			break
		}
		ls.spe = append(ls.spe, spe)
	}
	return ls
}

func (ls *spList) Len() int {
	return len(ls.spe)
}

func (ls *spList) Item(i int) string {
	if len(ls.spe[i].Catalog) > 0 {
		return ls.spe[i].Catalog
	}
	return "[id: " + ls.spe[i].Id + "]"
}

func (ls *spList) IsSel(i int) bool {
	if ls.sel == i {
		return true
	}
	return false
}
