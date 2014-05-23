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

var txInfo = &cmdapp.Command{
	Name: "tx.info",
	Synopsis: `[-e|--extdb name] [-i|--id value] [-k|--key value]
	[-m|--machine] [-p|--port value] [<name> [<parentname>]]`,
	Short:    "prints general taxon information",
	IsCommon: true,
	Long: `
Description

Tx.info prints general information of a taxon in the database. For the
list of parents, descendants of synonyms of the taxon, use 'jdh tx.ls'.

Options

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
    
    -k value
    --key value
      If set, only a particular value of the taxon will be printed.
      Valid keys are:
          authority      Authorship of the taxon.
          comment        A free text comment on the taxon.
          extern         Extern identifiers of the taxon, in the form
                         <service>:<key>, for example: gbif:5216933.
          name           Name of the taxon.
          parent         Id of the new parent.
          rank           The taxon rank.
          synonym        Prints the parent of the taxon, if it is a synonym.
                         If the taxon is valid, the valid string will be printed.
          valid          See synonym.

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.
      
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
	txInfo.Flag.StringVar(&extDBFlag, "extdb", "", "")
	txInfo.Flag.StringVar(&extDBFlag, "e", "", "")
	txInfo.Flag.StringVar(&idFlag, "id", "", "")
	txInfo.Flag.StringVar(&idFlag, "i", "", "")
	txInfo.Flag.StringVar(&keyFlag, "key", "", "")
	txInfo.Flag.StringVar(&keyFlag, "k", "", "")
	txInfo.Flag.BoolVar(&machineFlag, "machine", false, "")
	txInfo.Flag.BoolVar(&machineFlag, "m", false, "")
	txInfo.Flag.StringVar(&portFlag, "port", "", "")
	txInfo.Flag.StringVar(&portFlag, "p", "", "")
	txInfo.Run = txInfoRun
}

func txInfoRun(c *cmdapp.Command, args []string) {
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
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong taxon name or id"))
		c.Usage()
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxName, tax.Name)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxAuthority, tax.Authority)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxRank, tax.Rank)
			if tax.IsValid {
				fmt.Fprintf(os.Stdout, "%s=true\n", jdh.TaxValid)
			} else {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxSynonym, tax.Parent)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxParent, tax.Parent)
			for _, e := range tax.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, tax.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.TaxName:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxName, tax.Name)
		case jdh.TaxAuthority:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxAuthority, tax.Authority)
		case jdh.TaxRank:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxRank, tax.Rank)
		case jdh.TaxValid, jdh.TaxSynonym:
			if tax.IsValid {
				fmt.Fprintf(os.Stdout, "%s=true\n", jdh.TaxValid)
			} else {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxSynonym, tax.Parent)
			}
		case jdh.TaxParent:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TaxParent, tax.Parent)
		case jdh.KeyExtern:
			for _, e := range tax.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, tax.Comment)
		}
		return
	}
	if len(keyFlag) == 0 {
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", tax.Id)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Name:", tax.Name)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Authority:", tax.Authority)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Rank:", tax.Rank)
		if tax.IsValid {
			fmt.Fprintf(os.Stdout, "%-16s true\n", "Valid:")
		} else {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Synonym of:", tax.Parent)
		}
		if len(tax.Parent) > 0 {
			p := taxon(c, db, tax.Parent)
			fmt.Fprintf(os.Stdout, "%-16s %s %s [id: %s]\n", "Parent:", p.Name, p.Authority, p.Id)
		}
		if len(tax.Extern) > 0 {
			fmt.Fprintf(os.Stdout, "Extern ids:\n")
			for _, e := range tax.Extern {
				fmt.Fprintf(os.Stdout, "\t%s\n", e)
			}
		}
		if len(tax.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", tax.Comment)
		}
		txInfoList(c, db, tax.Id, true)
		txInfoList(c, db, tax.Id, false)
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.TaxName:
		fmt.Fprintf(os.Stdout, "%s\n", tax.Name)
	case jdh.TaxAuthority:
		fmt.Fprintf(os.Stdout, "%s\n", tax.Authority)
	case jdh.TaxRank:
		fmt.Fprintf(os.Stdout, "%s\n", tax.Rank)
	case jdh.TaxValid, jdh.TaxSynonym:
		if tax.IsValid {
			fmt.Fprintf(os.Stdout, "true\n")
		} else {
			fmt.Fprintf(os.Stdout, "synonym of %s\n", tax.Parent)
		}
	case jdh.TaxParent:
		fmt.Fprintf(os.Stdout, "%s\n", tax.Parent)
	case jdh.KeyExtern:
		for _, e := range tax.Extern {
			fmt.Fprintf(os.Stdout, "%s\n", e)
		}
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", tax.Comment)
	}
}

func txInfoList(c *cmdapp.Command, db jdh.DB, pId string, valid bool) {
	l := getTaxDesc(c, db, pId, valid)
	first := true
	for {
		tax := &jdh.Taxon{}
		if err := l.Scan(tax); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if first {
			if valid {
				fmt.Fprintf(os.Stdout, "Children taxa:\n")
			} else {
				fmt.Fprintf(os.Stdout, "Synonyms:\n")
			}
			first = false
		}
		fmt.Fprintf(os.Stdout, "\t%s %s [id: %s]\n", tax.Name, tax.Authority, tax.Id)
	}
}
