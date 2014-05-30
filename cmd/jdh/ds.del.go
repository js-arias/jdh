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

var dsDel = &cmdapp.Command{
	Name:     "ds.del",
	Synopsis: `-i|--id value [-p|--port value]`,
	Short:    "deletes a dataset",
	Long: `
Description

Ds.del removes a dataset, but not the specimens associated with it, from the
database.

Options

    -i value
    --id value
      Search for the indicated dataset id. It is a required option.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	dsDel.Flag.StringVar(&idFlag, "id", "", "")
	dsDel.Flag.StringVar(&idFlag, "i", "", "")
	dsDel.Flag.StringVar(&portFlag, "port", "", "")
	dsDel.Flag.StringVar(&portFlag, "p", "", "")
	dsDel.Run = dsDelRun
}

func dsDelRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong dataset id"))
		c.Usage()
	}
	openLocal(c)
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, idFlag)
	if _, err := localDB.Exec(jdh.Delete, jdh.Datasets, vals); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	localDB.Exec(jdh.Commit, "", nil)
}
