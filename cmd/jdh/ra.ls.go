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

var raLs = &cmdapp.Command{
	Name: "ra.ls",
	Synopsis: `[-c|--children] [-m|--machine] [-p|--port value] 
	[-t|--taxon value] [-v|--verbose] [<name> [<parentname>]]`,
	Short:    "prints a list of rasterized distributions",
	IsCommon: true,
	Long: `
Description

Ra.ls prints a list of rasterized distributions associated with a taxon.

Options

    -c
    --children
      If set, the rastes associated with the indicated taxon, as well
      as the ones from its descendants, will be printed.
    
    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
    -t value
    --taxon value
      Search for the indicated taxon id.

    -v
    --verbose
      If defined, then a large list (including ids) will be printed. This 
      option is ignored if -m or --machine option is defined.
      
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
	raLs.Flag.BoolVar(&childFlag, "children", false, "")
	raLs.Flag.BoolVar(&childFlag, "c", false, "")
	raLs.Flag.BoolVar(&machineFlag, "machine", false, "")
	raLs.Flag.BoolVar(&machineFlag, "m", false, "")
	raLs.Flag.StringVar(&portFlag, "port", "", "")
	raLs.Flag.StringVar(&portFlag, "p", "", "")
	raLs.Flag.StringVar(&taxonFlag, "taxon", "", "")
	raLs.Flag.StringVar(&taxonFlag, "t", "", "")
	raLs.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	raLs.Flag.BoolVar(&verboseFlag, "v", false, "")
	raLs.Run = raLsRun
}

func raLsRun(c *cmdapp.Command, args []string) {
	openLocal(c)
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
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong taxon name or id"))
		c.Usage()
	}
	vals := new(jdh.Values)
	if childFlag {
		vals.Add(jdh.RDisTaxonParent, tax.Id)
	} else {
		vals.Add(jdh.RDisTaxon, tax.Id)
	}
	l := rasList(c, localDB, vals)
	defer l.Close()
	ct := tax
	for {
		ras := &jdh.Raster{}
		if err := l.Scan(ras); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", ras.Id)
			continue
		}
		if ras.Taxon != ct.Id {
			ct = taxon(c, localDB, ras.Taxon)
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s %s %s\t%s\t%dx%d\n", ct.Id, ct.Name, ct.Authority, ras.Id, ras.Cols, ras.Cols/2)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\t%s\t%dx%d\n", ct.Name, ras.Id, ras.Cols, ras.Cols/2)
	}
}
