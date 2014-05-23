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

// RasList returns a raster distribution list scanner.
func rasList(c *cmdapp.Command, db jdh.DB, vals *jdh.Values) jdh.ListScanner {
	l, err := localDB.List(jdh.RasDistros, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return l
}

// Raster gets the raster distribution.
func raster(c *cmdapp.Command, db jdh.DB, id string) *jdh.Raster {
	sc, err := db.Get(jdh.RasDistros, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	ras := &jdh.Raster{}
	if err := sc.Scan(ras); err != nil {
		if err == io.EOF {
			return &jdh.Raster{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return ras
}
