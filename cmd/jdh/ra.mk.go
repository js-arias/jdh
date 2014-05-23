// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"io"
	"os"
	"strconv"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
	pixlist "github.com/js-arias/jdh/pkg/raster"
)

var raMk = &cmdapp.Command{
	Name: "ra.mk",
	Synopsis: `[-e|--extdb name] [-p|--port value] [-r|--rank name]
	[-s|--size value] [-t|--taxon value] [-d|--validated]
	[<name> [<parentname>]]`,
	Short:    "creates raster distributions from specimen data",
	IsCommon: true,
	Long: `
Description

Ra.mk uses specimen data in the database (or an extern database, defined with
-e, --extdb option) to creates a precense-absence rasterized distributions.
Only valid taxons will be rasterized (although their sinonyms will be used to
create the raster).

When the options -t, --taxon or a name are used, the effect of the command
will only affect the indicated taxon and its descendants.

When the -r, --rank option is used, only the taxons at or below the indicated
rank will be rasterized.

Options

    -e name
    --extdb name
      Sets the a extern database to extract distribution data.
    
    -d
    --validated
      If set, only specimens with validated georeferences will be used to
      build the raster.

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

    -s value
    --size value
      Sets the size of a pixel side (pixels, or cells are assumed as squared),
      in terms of arc degrees. By default, the value is 1. It must be a value
      between 0 and 360 (not inclusive). The value will be arranged to make
      the number of colums fit well in the 360 degrees.
      
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
	raMk.Flag.BoolVar(&validFlag, "validate", false, "")
	raMk.Flag.BoolVar(&validFlag, "d", false, "")
	raMk.Flag.StringVar(&extDBFlag, "extdb", "", "")
	raMk.Flag.StringVar(&extDBFlag, "e", "", "")
	raMk.Flag.StringVar(&portFlag, "port", "", "")
	raMk.Flag.StringVar(&portFlag, "p", "", "")
	raMk.Flag.StringVar(&rankFlag, "rank", "", "")
	raMk.Flag.StringVar(&rankFlag, "r", "", "")
	raMk.Flag.Float64Var(&sizeFlag, "size", 1, "")
	raMk.Flag.Float64Var(&sizeFlag, "s", 1, "")
	raMk.Flag.StringVar(&taxonFlag, "taxon", "", "")
	raMk.Flag.StringVar(&taxonFlag, "t", "", "")
	raMk.Run = raMkRun
}

func raMkRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	var spDB jdh.DB
	if len(extDBFlag) > 0 {
		openExt(c, extDBFlag, "")
		spDB = extDB
	} else {
		spDB = localDB
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
	if (sizeFlag <= 0) || (sizeFlag >= 360) {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("invalid size option"))
		os.Exit(1)
	}
	cols := uint(360 / sizeFlag)
	if int(float64(cols)*sizeFlag) < 360 {
		cols++
	}
	size := 360 / float64(cols)
	raMkFetch(c, spDB, tax, jdh.Kingdom, rank, size)
	localDB.Exec(jdh.Commit, "", nil)
}

func raMkFetch(c *cmdapp.Command, spDB jdh.DB, tax *jdh.Taxon, prevRank, rank jdh.Rank, size float64) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
	defer func() {
		l := getTaxDesc(c, localDB, tax.Id, true)
		raMkNav(c, l, spDB, r, rank, size)
	}()
	if len(tax.Id) == 0 {
		return
	}
	if r < rank {
		return
	}
	vals := new(jdh.Values)
	if spDB != localDB {
		eid := searchExtern(extDBFlag, tax.Extern)
		if len(eid) == 0 {
			return
		}
		vals.Add(jdh.SpeTaxonParent, eid)
	} else {
		vals.Add(jdh.SpeTaxonParent, tax.Id)
	}
	vals.Add(jdh.SpeGeoref, "true")
	inDB := true
	ras := raMkGetRas(c, tax.Id, size)
	if len(ras.Id) == 0 {
		inDB = false
		ras = &jdh.Raster{
			Taxon:  tax.Id,
			Cols:   uint(360 / size),
			Source: jdh.ExplicitPoints,
		}
	}
	l := speList(c, spDB, vals)
	for {
		spe := &jdh.Specimen{}
		if err := l.Scan(spe); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if !spe.Georef.IsValid() {
			continue
		}
		if validFlag && (len(spe.Georef.Validation) == 0) {
			continue
		}
		azm, inc := spe.Georef.Point.Lon+180, 90-spe.Georef.Point.Lat
		x, y := int(azm/size), int(inc/size)
		if inDB {
			vals.Reset()
			vals.Add(jdh.KeyId, ras.Id)
			vals.Add(jdh.RasPixel, fmt.Sprintf("%d,%d,1", x, y))
			localDB.Exec(jdh.Set, jdh.RasDistros, vals)
			continue
		}
		if ras.Raster == nil {
			ras.Raster = pixlist.NewPixList()
		}
		pt := image.Pt(x, y)
		ras.Raster.Set(pt, 1)
	}
	if inDB || (ras.Raster == nil) {
		return
	}
	localDB.Exec(jdh.Add, jdh.RasDistros, ras)
}

func raMkNav(c *cmdapp.Command, l jdh.ListScanner, spDB jdh.DB, prevRank, rank jdh.Rank, size float64) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		raMkFetch(c, spDB, desc, prevRank, rank, size)
	}
}

func raMkGetRas(c *cmdapp.Command, id string, size float64) *jdh.Raster {
	vals := new(jdh.Values)
	vals.Add(jdh.RDisTaxon, id)
	vals.Add(jdh.RDisCols, strconv.FormatInt(int64(360/size), 10))
	vals.Add(jdh.RDisSource, jdh.ExplicitPoints.String())
	l := rasList(c, localDB, vals)
	defer l.Close()
	for {
		ras := &jdh.Raster{}
		if err := l.Scan(ras); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		return ras
	}
	return &jdh.Raster{}
}
