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

// Dataset gets a dataset.
func dataset(c *cmdapp.Command, db jdh.DB, id string) *jdh.Dataset {
	sc, err := db.Get(jdh.Datasets, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	set := &jdh.Dataset{}
	if err := sc.Scan(set); err != nil {
		if err == io.EOF {
			return &jdh.Dataset{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return set
}
