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

var txDel = &cmdapp.Command{
	Name: "tx.del",
	Synopsis: `[-c|--collapse] [-i|--id value] [-p|--port value]
	[<name> [<parentname>]]`,
	Short: "deletes a taxon",
	Long: `
Description

Tx.del removes a taxon from the database, by default it deletes the taxon,
and all of its descendants. If the option -c, --collapse is defined, then
the taxon will be "collapsed": all of its descendants will be assigned to
the the taxon's parent, and then, it will be deleted.

Options

    -c
    --collapse
      Collapse the taxon: all the descendants (including synonyms) will be
      assigned to the ancestor of the taxon. If the taxon has no ancestor,
      then the valid descendants will be assigned to the root of the
      taxonomy and synonyms will be deleted with the taxon.

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.
    
    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id 
      are defined.   
	`,
}

func init() {
	txDel.Flag.BoolVar(&colpFlag, "collapse", false, "")
	txDel.Flag.BoolVar(&colpFlag, "c", false, "")
	txDel.Flag.StringVar(&idFlag, "id", "", "")
	txDel.Flag.StringVar(&idFlag, "i", "", "")
	txDel.Flag.StringVar(&portFlag, "port", "", "")
	txDel.Flag.StringVar(&portFlag, "p", "", "")
	txDel.Run = txDelRun
}

func txDelRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	var tax *jdh.Taxon
	if len(idFlag) > 0 {
		tax = taxon(c, localDB, idFlag)
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
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong taxon name or id"))
		c.Usage()
	}
	vals := new(jdh.Values)
	if colpFlag {
		if len(tax.Parent) > 0 {
			vals.Add(jdh.KeyId, tax.Id)
			vals.Add(jdh.TaxSynonym, "")
			if _, err := localDB.Exec(jdh.Set, jdh.Taxonomy, vals); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
		} else {
			l := getTaxDesc(c, localDB, tax.Id, true)
			for {
				desc := &jdh.Taxon{}
				if err := l.Scan(desc); err != nil {
					if err == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					os.Exit(1)
				}
				vals.Reset()
				vals.Add(jdh.KeyId, desc.Id)
				vals.Add(jdh.TaxParent, "")
				if _, err := localDB.Exec(jdh.Set, jdh.Taxonomy, vals); err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					os.Exit(1)
				}
			}
		}
	}
	vals.Reset()
	vals.Add(jdh.KeyId, tax.Id)
	if _, err := localDB.Exec(jdh.Delete, jdh.Taxonomy, vals); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	localDB.Exec(jdh.Commit, "", nil)
}
