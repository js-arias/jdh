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

var txLs = &cmdapp.Command{
	Name: "tx.ls",
	Synopsis: `[-a|--ancs] [-e|--extdb name] [-i|--id value]
	[-m|--machine] [-p|--port value] [-r|--rank name] [-s|--synonym]
	[-v|--verbose] [<name> [<parentname>]]`,
	Short:    "prints a list of taxons",
	IsCommon: true,
	Long: `
Description

Tx.ls prints a list of taxons. With no option, tx.ls prints the taxons attached
to the root of taxonomy.

If a name or -i, --id option is defined, then the descendants of the indicated
taxon will be printed by default. This behaviour can be changed with other
options (e.g. -a, --ancs to show parents).

Options

    -a
    --ancs
      If set, the parents of the indicated taxon will be printed.

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).
    
    -i value
    --id value
      Search for the indicated taxon id.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
      
    -r name
    --rank name
      If indicated, the only taxons at the given rank will be printed. Valid
      values are:
          unranked
          kingdom
          class
          order
          family
          genus
          species

    -s
    --synonym
      If set, the synonyms of the indicated taxon will be printed.

    -v
    --verbose
      If defined, then a large list (including ids and authors) will be
      printed. This option is ignored if -m or --machine option is defined.
      
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
	txLs.Flag.BoolVar(&ancsFlag, "ancs", false, "")
	txLs.Flag.BoolVar(&ancsFlag, "a", false, "")
	txLs.Flag.StringVar(&extDBFlag, "extdb", "", "")
	txLs.Flag.StringVar(&extDBFlag, "e", "", "")
	txLs.Flag.StringVar(&idFlag, "id", "", "")
	txLs.Flag.StringVar(&idFlag, "i", "", "")
	txLs.Flag.BoolVar(&machineFlag, "machine", false, "")
	txLs.Flag.BoolVar(&machineFlag, "m", false, "")
	txLs.Flag.StringVar(&portFlag, "port", "", "")
	txLs.Flag.StringVar(&portFlag, "p", "", "")
	txLs.Flag.StringVar(&rankFlag, "rank", "", "")
	txLs.Flag.StringVar(&rankFlag, "r", "", "")
	txLs.Flag.BoolVar(&synonymFlag, "synonym", false, "")
	txLs.Flag.BoolVar(&synonymFlag, "s", false, "")
	txLs.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	txLs.Flag.BoolVar(&verboseFlag, "v", false, "")
	txLs.Run = txLsRun
}

func txLsRun(c *cmdapp.Command, args []string) {
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}
	var tax *jdh.Taxon
	if len(idFlag) > 0 {
		tax = taxon(c, db, idFlag)
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
		tax = pickTaxName(c, db, args[0], pName)
		if len(tax.Id) == 0 {
			return
		}
	}
	if ancsFlag {
		if tax == nil {
			os.Exit(0)
		}
		vals := new(jdh.Values)
		vals.Add(jdh.TaxParents, tax.Id)
		l, err := db.List(jdh.Taxonomy, vals)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txLsProc(c, l)
		return
	}
	if len(rankFlag) > 0 {
		rank := jdh.GetRank(rankFlag)
		if (rank == jdh.Unranked) && (strings.ToLower(rankFlag) != rank.String()) {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("unknown rank"))
			os.Exit(1)
		}
		txLsRank(c, db, tax, rank)
		return
	}
	if synonymFlag {
		if tax == nil {
			os.Exit(0)
		}
		l := getTaxDesc(c, db, tax.Id, false)
		txLsProc(c, l)
		return
	}
	id := ""
	if tax != nil {
		id = tax.Id
	}
	l := getTaxDesc(c, db, id, true)
	txLsProc(c, l)
}

func txLsProc(c *cmdapp.Command, l jdh.ListScanner) {
	for {
		tax := &jdh.Taxon{}
		if err := l.Scan(tax); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", tax.Id)
			continue
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t%s\t%s %s\t\n", tax.Id, tax.Rank, tax.Name, tax.Authority)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s %s\n", tax.Name, tax.Authority)
	}
}

func txLsRank(c *cmdapp.Command, db jdh.DB, tax *jdh.Taxon, rank jdh.Rank) {
	if (tax == nil) || (len(tax.Id) == 0) {
		txLsRankNav(c, db, "", rank)
		return
	}
	if tax.Rank == rank {
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", tax.Id)
		} else if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t%s\t%s %s\n", tax.Id, tax.Rank, tax.Name, tax.Authority)
		} else {
			fmt.Fprintf(os.Stdout, "%s %s\n", tax.Name, tax.Authority)
		}
		if rank != jdh.Unranked {
			// only continue check if the asked rank is "unranked"
			return
		}
	} else if (tax.Rank > rank) && (rank != jdh.Unranked) {
		// only continue check if the asked rank is "unranked"
		return
	}
	txLsRankNav(c, db, tax.Id, rank)
}

func txLsRankNav(c *cmdapp.Command, db jdh.DB, id string, rank jdh.Rank) {
	args := new(jdh.Values)
	args.Add(jdh.TaxChildren, id)
	l, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if len(desc.Id) == 0 {
			continue
		}
		txLsRank(c, db, desc, rank)
	}
}
