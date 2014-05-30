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
