// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"

	_ "github.com/js-arias/jdh/pkg/driver/gbif"
	_ "github.com/js-arias/jdh/pkg/driver/inat"
	_ "github.com/js-arias/jdh/pkg/driver/native"
	_ "github.com/js-arias/jdh/pkg/driver/ncbi"
)

// databases
var (
	localDB jdh.DB // local database
	extDB   jdh.DB // extern database
)

// openLocal opens the local database.
func openLocal(c *cmdapp.Command) {
	localDB = openDB(c, "native", portFlag)
}

// openExt opens the extern database.
func openExt(c *cmdapp.Command, driver, par string) {
	extDB = openDB(c, driver, par)
}

// openDB opens a database.
func openDB(c *cmdapp.Command, driver, par string) jdh.DB {
	db, err := jdh.Open(driver, par)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return db
}
