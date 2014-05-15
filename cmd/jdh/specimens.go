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

// Specimen gets an specimen.
func specimen(c *cmdapp.Command, db jdh.DB, id string) *jdh.Specimen {
	sc, err := db.Get(jdh.Specimens, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	spe := &jdh.Specimen{}
	if err := sc.Scan(spe); err != nil {
		if err == io.EOF {
			return &jdh.Specimen{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return spe
}

// SpeList returns an specimen list scanner.
func speList(c *cmdapp.Command, db jdh.DB, vals *jdh.Values) jdh.ListScanner {
	l, err := db.List(jdh.Specimens, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return l
}
