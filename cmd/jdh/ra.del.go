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

var raDel = &cmdapp.Command{
	Name: "ra.del",
	Synopsis: `[-i|--id value] [-p|--port value] [-t|--taxon value]
	[<name> [<parentname>]]`,
	Short: "deletes rasterized distributions",
	Long: `
Description

Ra.del removes a rasterized distribution from the database, or, if option -t
or --taxon is defined, or taxon name is given, all the rasterized
distributions associated with the indicated taxon. This option deletes the
rasters, neither specimens or taxons, use command sp.del or tx.del to perform
that operations.

Operations

    -i value
    --id value
      Search for the indicated raster id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      Search for the indicated taxon id.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.
    
    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon 
      are defined.      
	`,
}

func init() {
	raDel.Flag.StringVar(&idFlag, "id", "", "")
	raDel.Flag.StringVar(&idFlag, "i", "", "")
	raDel.Flag.StringVar(&portFlag, "port", "", "")
	raDel.Flag.StringVar(&portFlag, "p", "", "")
	raDel.Flag.StringVar(&taxonFlag, "taxon", "", "")
	raDel.Flag.StringVar(&taxonFlag, "t", "", "")
	raDel.Run = raDelRun
}

func raDelRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	vals := new(jdh.Values)
	if len(idFlag) > 0 {
		vals.Add(jdh.KeyId, idFlag)
		if _, err := localDB.Exec(jdh.Delete, jdh.RasDistros, vals); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		localDB.Exec(jdh.Commit, "", nil)
		return
	}
	var tax *jdh.Taxon
	if len(taxonFlag) > 0 {
		tax = taxon(c, localDB, taxonFlag)
		if len(tax.Id) == 0 {
			return
		}
	} else if len(args) > 0 {
		if len(args) > 2 {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("too many arguments"))
			os.Exit(1)
		}
		pName := ""
		if len(args) > 1 {
			pName = args[1]
		}
		tax = pickTaxName(c, localDB, args[0], pName)
		if len(tax.Id) == 0 {
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong raster id or taxon name or id"))
		c.Usage()
	}
	vals.Add(jdh.RDisTaxon, tax.Id)
	if _, err := localDB.Exec(jdh.Delete, jdh.RasDistros, vals); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	localDB.Exec(jdh.Commit, "", nil)
}
