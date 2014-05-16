// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/geography"
	"github.com/js-arias/jdh/pkg/jdh"
)

var spPop = &cmdapp.Command{
	Name: "sp.pop",
	Synopsis: `-e|--extdb name [-p|--port value] [-r|--rank name]
	[-t|--taxon value] [<name> [<parentname>]]`,
	Short:    "add specimens from an extern database",
	IsCommon: true,
	Long: `
Description

Sp.pop uses an extern database to populate the local database with specimens.

When the options -t, --taxon or a name are used, the effect of the command
will only affect the indicated taxon and its descendants.

Options

    -e name
    --extdb name
      Set the extern database.
      Valid values are:
          gbif    specimens from gbif.
      This parameter is required.
    
    -g
    --georef
      If set, only the specimens georeferenced in the extern database
      will be added.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
    -r name
    --rank name
      If set, only taxons below the indicated rank will be populated.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -t value
    --taxon value
      Search for the indicated taxon id.

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
	spPop.Flag.StringVar(&extDBFlag, "extdb", "", "")
	spPop.Flag.StringVar(&extDBFlag, "e", "", "")
	spPop.Flag.BoolVar(&geoRefFlag, "georef", false, "")
	spPop.Flag.BoolVar(&geoRefFlag, "g", false, "")
	spPop.Flag.StringVar(&portFlag, "port", "", "")
	spPop.Flag.StringVar(&portFlag, "p", "", "")
	spPop.Flag.StringVar(&rankFlag, "rank", "", "")
	spPop.Flag.StringVar(&rankFlag, "r", "", "")
	spPop.Flag.StringVar(&taxonFlag, "taxon", "", "")
	spPop.Flag.StringVar(&taxonFlag, "t", "", "")
	spPop.Run = spPopRun
}

func spPopRun(c *cmdapp.Command, args []string) {
	if len(extDBFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong '--extdb' option"))
		c.Usage()
	}
	openLocal(c)
	openExt(c, extDBFlag, "")
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
		tax = &jdh.Taxon{}
	}
	rank := jdh.Kingdom
	if len(rankFlag) > 0 {
		rank = jdh.GetRank(rankFlag)
		if rank == jdh.Unranked {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("invalid rank"))
			os.Exit(1)
		}
	}
	spPopFetch(c, tax, jdh.Kingdom, rank)
	localDB.Exec(jdh.Commit, "", nil)
}

func spPopFetch(c *cmdapp.Command, tax *jdh.Taxon, prevRank, rank jdh.Rank) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
	defer func() {
		l := getTaxDesc(c, localDB, tax.Id, true)
		spPopNav(c, l, r, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		spPopNav(c, l, r, rank)
	}()
	if len(tax.Id) == 0 {
		return
	}
	if r < rank {
		return
	}
	eid := searchExtern(extDBFlag, tax.Extern)
	if len(eid) == 0 {
		return
	}
	vals := new(jdh.Values)
	vals.Add(jdh.SpeTaxon, eid)
	if geoRefFlag {
		vals.Add(jdh.SpeGeoref, "true")
	}
	l := speList(c, extDB, vals)
	for {
		spe := &jdh.Specimen{}
		if err := l.Scan(spe); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if osp := specimen(c, localDB, extDBFlag+":"+spe.Id); len(osp.Id) > 0 {
			continue
		}
		if len(spe.Catalog) > 0 {
			if osp := specimen(c, localDB, spe.Catalog); len(osp.Id) > 0 {
				exsp := searchExtern(extDBFlag, osp.Extern)
				if exsp == spe.Id {
					continue
				}
				fmt.Fprintf(os.Stderr, "specimen %s already in database as %s [duplicated in %s]\n", spe.Catalog, osp.Id, extDBFlag)
				continue
			}
		}
		addToSpecimens(c, spe, tax.Id)
	}
}

func addToSpecimens(c *cmdapp.Command, src *jdh.Specimen, tax string) string {
	dest := &jdh.Specimen{}
	*dest = *src
	dest.Id = ""
	dest.Taxon = tax
	dest.Extern = []string{extDBFlag + ":" + src.Id}
	if !src.Georef.IsValid() {
		dest.Georef = geography.InvalidGeoref()
	}
	if len(src.Dataset) > 0 {
		dest.Dataset = getValidDataset(c, src.Dataset)
	}
	id, err := localDB.Exec(jdh.Add, jdh.Specimens, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return id
}

func spPopNav(c *cmdapp.Command, l jdh.ListScanner, prevRank, rank jdh.Rank) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		spPopFetch(c, desc, prevRank, rank)
	}
}

func getValidDataset(c *cmdapp.Command, extId string) string {
	set := dataset(c, localDB, extDBFlag+":"+extId)
	if len(set.Id) > 0 {
		return set.Id
	}
	set = dataset(c, extDB, extId)
	if len(set.Id) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(fmt.Errorf("unable to found collection %s in %s", extId, extDB)))
		os.Exit(1)
	}
	dest := &jdh.Dataset{}
	*dest = *set
	dest.Id = ""
	dest.Extern = []string{extDBFlag + ":" + set.Id}
	id, err := localDB.Exec(jdh.Add, jdh.Datasets, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return id
}
