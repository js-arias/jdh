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

var spLs = &cmdapp.Command{
	Name: "sp.ls",
	Synopsis: `[-c|--children] [-e|--extdb name] [-g|--georef]
	[-m|--machine] [-n|--nonref] [-p|--port value] [-r|--country name]
	[-t|--taxon value] [-v|--verbose] [<name> [<parentname>]]`,
	Short:    "prints a list of specimens",
	IsCommon: true,
	Long: `
Description

Sp.ls prints a list of specimens associated with a taxon.

By default, all records will be printed, use -g, --georef, -n, or --nogeoref
to modify this behavior.

Options

    -c
    --children
      If set, the speciemens associated with the indicated taxon, as well
      as the ones from its descendants, will be printed.
    
    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    specimens from gbif.

    -g
    --georef
      If set, only georeferenced records will be printed.
      
    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.
    
    -n
    --nonref
      If defined, only records without a georeference will be printed. It
      will be ignored if -g, --georef is defined.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
    -r name
    --country name
      If set, only specimens reported to the given country will be printed.
    
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
	spLs.Flag.BoolVar(&childFlag, "children", false, "")
	spLs.Flag.BoolVar(&childFlag, "c", false, "")
	spLs.Flag.StringVar(&extDBFlag, "extdb", "", "")
	spLs.Flag.StringVar(&extDBFlag, "e", "", "")
	spLs.Flag.BoolVar(&geoRefFlag, "georef", false, "")
	spLs.Flag.BoolVar(&geoRefFlag, "g", false, "")
	spLs.Flag.BoolVar(&machineFlag, "machine", false, "")
	spLs.Flag.BoolVar(&machineFlag, "m", false, "")
	spLs.Flag.BoolVar(&noRefFlag, "noref", false, "")
	spLs.Flag.BoolVar(&noRefFlag, "n", false, "")
	spLs.Flag.StringVar(&portFlag, "port", "", "")
	spLs.Flag.StringVar(&portFlag, "p", "", "")
	spLs.Flag.StringVar(&countryFlag, "country", "", "")
	spLs.Flag.StringVar(&countryFlag, "r", "", "")
	spLs.Flag.StringVar(&taxonFlag, "taxon", "", "")
	spLs.Flag.StringVar(&taxonFlag, "t", "", "")
	spLs.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	spLs.Flag.BoolVar(&verboseFlag, "v", false, "")
	spLs.Run = spLsRun
}

func spLsRun(c *cmdapp.Command, args []string) {
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}
	var tax *jdh.Taxon
	if len(taxonFlag) > 0 {
		tax = taxon(c, db, taxonFlag)
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
	vals := new(jdh.Values)
	if childFlag {
		vals.Add(jdh.SpeTaxonParent, tax.Id)
	} else {
		vals.Add(jdh.SpeTaxon, tax.Id)
	}
	if len(countryFlag) > 0 {
		vals.Add(jdh.GeoCountry, countryFlag)
	}
	if geoRefFlag {
		vals.Add(jdh.SpeGeoref, "true")
	} else if noRefFlag {
		vals.Add(jdh.SpeGeoref, "false")
	}
	l := speList(c, db, vals)
	defer l.Close()
	ct := tax
	for {
		spe := &jdh.Specimen{}
		if err := l.Scan(spe); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", spe.Id)
			continue
		}
		if spe.Taxon != ct.Id {
			ct = taxon(c, db, spe.Taxon)
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s %s %s\t%s %s", ct.Id, ct.Name, ct.Authority, spe.Id, spe.Catalog)
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "\t%.5f %.5f %d", spe.Georef.Point.Lon, spe.Georef.Point.Lat, spe.Georef.Uncertainty)
			}
			fmt.Fprintf(os.Stdout, "\n")
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\t%s", ct.Name, spe.Catalog)
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "\t%.5f %.5f", spe.Georef.Point.Lon, spe.Georef.Point.Lat)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}
}
