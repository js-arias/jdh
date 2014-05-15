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

var dsSet = &cmdapp.Command{
	Name:     "ds.set",
	Synopsis: `-i|--id value [-p|--port value] [<key=value>...]`,
	Short:    "sets a dataset value",
	Long: `
Description

Ds.set sets a particular value for a dataset in the database. Use this
command to edit the dataset database, instead of manual edition.

If no key is defined, the key values will be read from the standard input, 
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the dataset to be set. It is a required option.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "title=Global Biodiversity Information Facility".
      Valid keys are:
          citation       Preferred citation of the dataset.
          comment        A free text comment on the dataset.
          license        License of use of the data.
          extern         Extern identifiers of the dataset, in the form
                         <service>:<key>.
          title          Title of the dataset.
          url            Url of the dataset.
	`,
}

func init() {
	dsSet.Flag.StringVar(&idFlag, "id", "", "")
	dsSet.Flag.StringVar(&idFlag, "i", "", "")
	dsSet.Flag.StringVar(&portFlag, "port", "", "")
	dsSet.Flag.StringVar(&portFlag, "p", "", "")
	dsSet.Run = dsSetRun
}

func dsSetRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong dataset id"))
		c.Usage()
	}
	openLocal(c)
	set := dataset(c, localDB, idFlag)
	if len(set.Id) == 0 {
		return
	}
	vals := valsFromArgs(set.Id, args)
	localDB.Exec(jdh.Set, jdh.Datasets, vals)
	localDB.Exec(jdh.Commit, "", nil)
}
