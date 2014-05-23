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

var raInfo = &cmdapp.Command{
	Name: "ra.info",
	Synopsis: `-i|--id value [-k|--key value] [-m|--machine]
	[-p|--port value]`,
	Short:    "prints information about a rasterized distribution",
	IsCommon: true,
	Long: `
Description

Ra.info prints general information of a rasterized distribution in the
database.

Options

    -i value
    --id value
      Search for the indicated rasterized distribution id.
      This option is required.

    -k value
    --key value
      If set, only a particular value of the specimen will be printed.
      Valid keys are:
          column         Number of columns in the raster.
          comment        A free text comment on the raster.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>.
          pixel          Sets a pixel value in the raster, in the form
                         "X,Y,Val", in which Val is an int.
          reference      A bibliographic reference to the raster.
          source         Source of the raster.
          taxon          Id of the taxon assigned to the raster.

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.
      
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	raInfo.Flag.StringVar(&idFlag, "id", "", "")
	raInfo.Flag.StringVar(&idFlag, "i", "", "")
	raInfo.Flag.StringVar(&keyFlag, "key", "", "")
	raInfo.Flag.StringVar(&keyFlag, "k", "", "")
	raInfo.Flag.BoolVar(&machineFlag, "machine", false, "")
	raInfo.Flag.BoolVar(&machineFlag, "m", false, "")
	raInfo.Flag.StringVar(&portFlag, "port", "", "")
	raInfo.Flag.StringVar(&portFlag, "p", "", "")
	raInfo.Run = raInfoRun
}

func raInfoRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong raster id"))
		c.Usage()
	}
	openLocal(c)
	ras := raster(c, localDB, idFlag)
	if len(ras.Id) == 0 {
		return
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisTaxon, ras.Taxon)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisSource, ras.Source)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyReference, ras.Reference)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisCols, ras.Cols)
			for _, e := range ras.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, ras.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.RDisTaxon:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisTaxon, ras.Taxon)
		case jdh.RDisSource:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisSource, ras.Source)
		case jdh.RDisCols:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.RDisCols, ras.Cols)
		case jdh.KeyExtern:
			for _, e := range ras.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, ras.Comment)
		case jdh.KeyReference:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyReference, ras.Reference)
		}
		return
	}
	if len(keyFlag) == 0 {
		tax := taxon(c, localDB, ras.Taxon)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", ras.Id)
		fmt.Fprintf(os.Stdout, "%-16s %s %s [id: %s]\n", "Taxon:", tax.Name, tax.Authority, ras.Taxon)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Source:", ras.Source)
		fmt.Fprintf(os.Stdout, "%-16s %dx%d\n", "Dimensions:", ras.Cols, ras.Cols/2)
		if len(ras.Reference) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Reference:", ras.Reference)
		}
		if len(ras.Extern) > 0 {
			fmt.Fprintf(os.Stdout, "Extern ids:\n")
			for _, e := range ras.Extern {
				fmt.Fprintf(os.Stdout, "\t%s\n", e)
			}
		}
		if len(ras.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", ras.Comment)
		}
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.RDisTaxon:
		fmt.Fprintf(os.Stdout, "%s\n", ras.Taxon)
	case jdh.RDisSource:
		fmt.Fprintf(os.Stdout, "%s\n", ras.Source)
	case jdh.RDisCols:
		fmt.Fprintf(os.Stdout, "%s\n", ras.Cols)
	case jdh.KeyExtern:
		for _, e := range ras.Extern {
			fmt.Fprintf(os.Stdout, "%s\n", e)
		}
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", ras.Comment)
	case jdh.KeyReference:
		fmt.Fprintf(os.Stdout, "%s\n", ras.Reference)
	}
}
