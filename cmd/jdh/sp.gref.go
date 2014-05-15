// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/geography"
	_ "github.com/js-arias/jdh/pkg/geography/geolocate"
	"github.com/js-arias/jdh/pkg/jdh"
)

var spGref = &cmdapp.Command{
	Name: "sp.gref",
	Synopsis: `[-a|--add] [-c|--correct] [-d|--delete] [-p|--port value]
	[-t|--taxon value] [-u|--uncert value] [-v|--verbose]
	[<name> [<parentname>]]`,
	Short: "validate and add specimen georeferences",
	Long: `
Description

Sp.gref uses a gazatteer service (geolocate web service 
<http://www.museum.tulane.edu/geolocate/>) to validate or add a georeference
of the specimens in the database.

Specimens without a set country and locality will indicated as not validated,
but not corrected, or deleted.

By default, it just print the specimens that fail the validation. With -c,
--correct option, it will try to correct the georeference, if possible (check
for flips in latitude and longitude, for example). With -d, --delete option,
georeferences from specimens that fail the validation, and unable to be
corrected will be eliminated.

With -a, --add option, non georeferences specimens will be searched, and if a
valid location is found (under a given bound, defined by -u, --uncert option),
the value will be used to set the point. If this option is not set, the point
will be indicated as not validates, but not corrected, or deleted.

Options

    -a
    --add
      Check non-georeferenced records, and, if they are found, they will be
      added to the database.
    
    -c
    --correct
      If set, it will try to correct invalid georeferences. It will try it
      by flipping lon, lat values, and changing lon, lat values sings.
    
    -d
    --delete
      If set, if an specimen can not be corrected, then the georeference (but
      not the specimen) will be removed from the database.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      If set, only specimens in selected taxon id will be searched.

    -u value
    --uncert value
      Set valid uncertainty (in meters), values below the given uncertainty
      will be scored as validated, or added, if -a, --add option is defined.
      Default value is 110000, which is roughly, about 1º at the equator. With
      0, the uncertainty values defined in each specimen will be used.

    -v
    --verbose
      If set, details on the error (if available), will be printed.


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
	spGref.Flag.BoolVar(&addFlag, "add", false, "")
	spGref.Flag.BoolVar(&addFlag, "a", false, "")
	spGref.Flag.BoolVar(&corrFlag, "correct", false, "")
	spGref.Flag.BoolVar(&corrFlag, "c", false, "")
	spGref.Flag.BoolVar(&delFlag, "delete", false, "")
	spGref.Flag.BoolVar(&delFlag, "d", false, "")
	spGref.Flag.StringVar(&portFlag, "port", "", "")
	spGref.Flag.StringVar(&portFlag, "p", "", "")
	spGref.Flag.StringVar(&taxonFlag, "taxon", "", "")
	spGref.Flag.StringVar(&taxonFlag, "t", "", "")
	spGref.Flag.IntVar(&uncertFlag, "uncert", 110000, "")
	spGref.Flag.IntVar(&uncertFlag, "u", 110000, "")
	spGref.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	spGref.Flag.BoolVar(&verboseFlag, "v", false, "")
	spGref.Run = spGrefRun
}

func spGrefRun(c *cmdapp.Command, args []string) {
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
		tax = &jdh.Taxon{}
	}
	gzt, err := geography.OpenGazetter("geolocate", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	spGrefProc(c, tax, gzt)
	if addFlag || corrFlag || delFlag {
		localDB.Exec(jdh.Commit, "", nil)
	}
}

func spGrefProc(c *cmdapp.Command, tax *jdh.Taxon, gzt geography.Gazetter) {
	defer func() {
		l := getTaxDesc(c, localDB, tax.Id, true)
		spGrefNav(c, l, gzt)
		l = getTaxDesc(c, localDB, tax.Id, false)
		spGrefNav(c, l, gzt)
	}()
	if len(tax.Id) == 0 {
		return
	}
	vals := new(jdh.Values)
	vals.Add(jdh.SpeTaxon, tax.Id)
	if !addFlag {
		vals.Add(jdh.LocGeoRef, "true")
	}
	l := speList(c, localDB, vals)
	defer l.Close()
	for {
		spe := &jdh.Specimen{}
		if err := l.Scan(spe); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if len(spe.Location.GeoRef.Validation) > 0 {
			continue
		}
		if !spe.Location.IsValid() {
			fmt.Fprintf(os.Stdout, "%s: location without country or locality\n", spe.Id)
			continue
		}
		if !spe.Location.GeoRef.Point.IsValid() {
			if !addFlag {
				fmt.Fprintf(os.Stdout, "%s: invalid georeference\n", spe.Id)
				continue
			}
			p, err := gzt.Locate(&spe.Location, uint(uncertFlag))
			if err != nil {
				fmt.Fprintf(os.Stdout, "%s: unable to add: %v\n", spe.Id, err)
				if verboseFlag {
					if err == geography.ErrAmbiguous {
						pts, err := gzt.List(&spe.Location, uint(uncertFlag))
						if err != nil {
							fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
							continue
						}
						for _, p := range pts {
							fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t%d\n", p.Point.Lon, p.Point.Lat, p.Uncertainty)
						}
					}
				}
				continue
			}
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, spe.Id)
			vals.Add(jdh.LocLonLat, strconv.FormatFloat(p.Point.Lon, 'g', -1, 64)+","+strconv.FormatFloat(p.Point.Lat, 'g', -1, 64))
			vals.Add(jdh.LocUncertainty, strconv.FormatInt(int64(p.Uncertainty), 10))
			vals.Add(jdh.LocSource, p.Source)
			vals.Add(jdh.LocValidation, p.Validation)
			localDB.Exec(jdh.Set, jdh.Specimens, vals)
			continue
		}
		u := uint(uncertFlag)
		if u == 0 {
			u = spe.Location.GeoRef.Uncertainty
		}
		pts, err := gzt.List(&spe.Location, u)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s: %v\n", spe.Id, err)
			continue
		}
		if len(pts) == 0 {
			fmt.Fprintf(os.Stdout, "%s: location not found\n", spe.Id, err)
			continue
		}
		lon, lat := spe.Location.GeoRef.Point.Lon, spe.Location.GeoRef.Point.Lat
		if (len(pts) > 0) && (u == 0) {
			fmt.Fprintf(os.Stdout, "%s: ambiguous location\n", spe.Id, err)
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t\t[current georeference]\n", lon, lat)
				for _, p := range pts {
					fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t%d\t%d\n", p.Point.Lon, p.Point.Lat, p.Uncertainty, p.Point.Distance(lon, lat))
				}
			}
			continue
		}
		val := false
		gr := geography.Georeference{
			Point:       geography.InvalidPoint(),
			Uncertainty: 50000000, // a distance large enough
		}
		for _, p := range pts {
			d := p.Point.Distance(lon, lat)
			if d <= u {
				val = true
				if (p.Uncertainty > 0) && (p.Uncertainty < gr.Uncertainty) {
					gr = p
				}
			}
		}
		if val {
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, spe.Id)
			if spe.Location.GeoRef.Uncertainty == 0 {
				vals.Add(jdh.LocUncertainty, strconv.FormatInt(int64(gr.Uncertainty), 10))
			}
			vals.Add(jdh.LocSource, gr.Source)
			vals.Add(jdh.LocValidation, gr.Validation)
			localDB.Exec(jdh.Set, jdh.Specimens, vals)
			continue
		}
		if !(corrFlag || delFlag) {
			fmt.Fprintf(os.Stdout, "%s: location not found\n", spe.Id, err)
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t\t[current georeference]\n", lon, lat)
				for _, p := range pts {
					fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t%d\t%d\n", p.Point.Lon, p.Point.Lat, p.Uncertainty, p.Point.Distance(lon, lat))
				}
			}
			continue
		}
		gr = geography.Georeference{
			Point:       geography.InvalidPoint(),
			Uncertainty: 50000000, // a distance large enough
		}
		val = false
		for _, p := range pts {
			// invert lon-lat
			lt, ln := lon, lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}

			// lon with wrong sign
			ln, lt = -lon, lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}

			// lat with wrong sing
			ln, lt = lon, -lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}

			// invert lon-lat, wrong sings
			lt, ln = -lon, -lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}

			// invert lon-lat, lon with wrong sing
			lt, ln = lon, -lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}

			// invert lon-lat, lat with wrong sing
			lt, ln = -lon, lat
			if geography.IsLon(ln) && geography.IsLat(lt) {
				d := p.Point.Distance(lon, lat)
				if d <= u {
					val = true
					gr = p
					gr.Point = geography.Point{Lon: ln, Lat: lt}
					break
				}
			}
		}
		if val {
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, spe.Id)
			vals.Add(jdh.LocLonLat, strconv.FormatFloat(gr.Point.Lon, 'g', -1, 64)+","+strconv.FormatFloat(gr.Point.Lat, 'g', -1, 64))
			vals.Add(jdh.LocUncertainty, strconv.FormatInt(int64(gr.Uncertainty), 10))
			vals.Add(jdh.LocSource, gr.Source)
			vals.Add(jdh.LocValidation, gr.Validation)
			localDB.Exec(jdh.Set, jdh.Specimens, vals)
			continue
		}
		if delFlag {
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, spe.Id)
			vals.Add(jdh.LocLonLat, "")
			localDB.Exec(jdh.Set, jdh.Specimens, vals)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s: not corrected\n", spe.Id)
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t\t[current georeference]\n", lon, lat)
			for _, p := range pts {
				fmt.Fprintf(os.Stderr, "\t%.5f %.5f\t%d\t%d\n", p.Point.Lon, p.Point.Lat, p.Uncertainty, p.Point.Distance(lon, lat))
			}
		}
	}
}

func spGrefNav(c *cmdapp.Command, l jdh.ListScanner, gzt geography.Gazetter) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		spGrefProc(c, desc, gzt)
	}
}
