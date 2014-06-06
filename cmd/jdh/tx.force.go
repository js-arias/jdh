// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var txForce = &cmdapp.Command{
	Name: "tx.force",
	Synopsis: `[-i|--id value] [-p|--port value] [-r|--rank name]
	[<name> [<parentname>]]`,
	Short: "enforces a ranked taxonomy",
	Long: `
Description

Tx.force enforces the database to be ranked, synonymizing all rankless taxa
with their most inmmediate ranked taxon.

Options

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set search only for taxons below the indicated rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

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
	txForce.Flag.StringVar(&idFlag, "id", "", "")
	txForce.Flag.StringVar(&idFlag, "i", "", "")
	txForce.Flag.StringVar(&portFlag, "port", "", "")
	txForce.Flag.StringVar(&portFlag, "p", "", "")
	txForce.Flag.StringVar(&rankFlag, "rank", "", "")
	txForce.Flag.StringVar(&rankFlag, "r", "", "")
	txForce.Run = txForceRun
}

func txForceRun(c *cmdapp.Command, args []string) {
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
		tax = &jdh.Taxon{}
	}
	rank := jdh.Kingdom
	if len(rankFlag) > 0 {
		rank := jdh.GetRank(rankFlag)
		if rank.String() != strings.ToLower(rankFlag) {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("invalid rank"))
			os.Exit(1)
		}
	}
	txForceProc(c, tax, jdh.Kingdom, rank)
	localDB.Exec(jdh.Commit, "", nil)
}

func txForceProc(c *cmdapp.Command, tax *jdh.Taxon, prevRank, rank jdh.Rank) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
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
		txForceProc(c, desc, r, rank)
	}

	if len(tax.Id) == 0 {
		return
	}
	if tax.Rank != jdh.Unranked {
		return
	}
	if r < rank {
		return
	}
	if len(tax.Parent) == 0 {
		return
	}
	args := new(jdh.Values)
	args.Add(jdh.KeyId, tax.Id)
	args.Add(jdh.TaxSynonym, tax.Parent)
	localDB.Exec(jdh.Set, jdh.Taxonomy, args)
}
