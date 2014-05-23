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

var raSet = &cmdapp.Command{
	Name:     "ra.set",
	Synopsis: `-i|--id value [-p|--port value] [<key=value>...]`,
	Short:    "sets a value in a rasterized distribution",
	Long: `
Description

Ra.set sets a particular value for a rasterized d in the database. Use this
command to edit the rasterized distribution database, instead of manual
edition.

If no key is defined, the key values will be read from the standard input, 
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the raster to be set. It is a required option.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "basis=preserved specimen"
      Valid keys are:
          comment        A free text comment on the raster.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>.
          pixel          Sets a pixel value in the raster, in the form
                         "X,Y,Val", in which Val is an int.
          reference      A bibliographic reference to the raster.
          source         Source of the raster.
          taxon          Id of the taxon assigned to the raster.
	`,
}

func init() {
	raSet.Flag.StringVar(&idFlag, "id", "", "")
	raSet.Flag.StringVar(&idFlag, "i", "", "")
	raSet.Flag.StringVar(&portFlag, "port", "", "")
	raSet.Flag.StringVar(&portFlag, "p", "", "")
	raSet.Run = raSetRun
}

func raSetRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong specimen id"))
		c.Usage()
	}
	openLocal(c)
	ras := raster(c, localDB, idFlag)
	if len(ras.Id) == 0 {
		return
	}
	vals := valsFromArgs(ras.Id, args)
	localDB.Exec(jdh.Set, jdh.RasDistros, vals)
	localDB.Exec(jdh.Commit, "", nil)
}
