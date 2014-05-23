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

var spInfo = &cmdapp.Command{
	Name: "sp.info",
	Synopsis: `-i|--id value [-e|--extdb name] [-k|--key value]
	[-m|--machine] [-p|--port value]`,
	Short:    "prints general specimen information",
	IsCommon: true,
	Long: `
Description

Sp.info prints general information of an specimen in the database.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    specimens from gbif.

    -i value
    --id value
      Search for the indicated specimen id.
      This option is required.

    -k value
    --key value
      If set, only a particular value of the specimen will be printed.
      Valid keys are:
          basis          Basis of record.
          catalog        Catalog code of the specimen.
          collector      Collector of the specimen.
          comment        A free text comment on the specimen.
          country        Country in which the specimen was collected, using
                         ISO 3166-1 alpha-2.
          county         County (or a similar administration entity) in
                         which the specimen was collected.
          dataset        Dataset that contains the specimen information.
          date           Date of specimen collection, in ISO 8601 format,
                         for example: 2006-01-02T15:04:05+07:00.
          determiner     Person who identify the specimen.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>, for example: gbif:866197949.
          locality       Locality in which the specimen was collected.
          lonlat         Longitude and latitude of the collection point.
          reference      A bibliographic reference to the specimen.
          source         Source of the georeference assignation.
          state          State or province in which the specimen was
                         collected.
          taxon          Id of the taxon assigned to the specimen.
          uncertainty    Uncertainty, in meters, of the georeference
                         assignation.
          validation     Source of the georeference validation.

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
	spInfo.Flag.StringVar(&extDBFlag, "extdb", "", "")
	spInfo.Flag.StringVar(&extDBFlag, "e", "", "")
	spInfo.Flag.StringVar(&idFlag, "id", "", "")
	spInfo.Flag.StringVar(&idFlag, "i", "", "")
	spInfo.Flag.StringVar(&keyFlag, "key", "", "")
	spInfo.Flag.StringVar(&keyFlag, "k", "", "")
	spInfo.Flag.BoolVar(&machineFlag, "machine", false, "")
	spInfo.Flag.BoolVar(&machineFlag, "m", false, "")
	spInfo.Flag.StringVar(&portFlag, "port", "", "")
	spInfo.Flag.StringVar(&portFlag, "p", "", "")
	spInfo.Run = spInfoRun
}

func spInfoRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong specimen id"))
		c.Usage()
	}
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}
	spe := specimen(c, db, idFlag)
	if len(spe.Id) == 0 {
		return
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeTaxon, spe.Taxon)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeBasis, spe.Basis)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyReference, spe.Reference)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDataset, spe.Dataset)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeCatalog, spe.Catalog)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDeterminer, spe.Determiner)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeCollector, spe.Collector)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDate, spe.Date.Format(jdh.Iso8601))
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoCountry, spe.Geography.Country)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoState, spe.Geography.State)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoCounty, spe.Geography.County)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeLocality, spe.Locality)
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "%s=%.8f,%.8f\n", jdh.GeoLonLat, spe.Georef.Point.Lon, spe.Georef.Point.Lat)
				fmt.Fprintf(os.Stdout, "%s=%d\n", jdh.GeoUncertainty, spe.Georef.Uncertainty)
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoSource, spe.Georef.Source)
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoValidation, spe.Georef.Validation)
			} else {
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoLonLat)
				fmt.Fprintf(os.Stdout, "%s=0\n", jdh.GeoUncertainty)
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoSource)
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoValidation)
			}
			for _, e := range spe.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, spe.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.SpeTaxon:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeTaxon, spe.Taxon)
		case jdh.SpeBasis:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeBasis, spe.Basis)
		case jdh.SpeDataset:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDataset, spe.Dataset)
		case jdh.SpeCatalog:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeCatalog, spe.Catalog)
		case jdh.SpeDeterminer:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDeterminer, spe.Determiner)
		case jdh.SpeCollector:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeCollector, spe.Collector)
		case jdh.SpeDate:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeDate, spe.Date.Format(jdh.Iso8601))
		case jdh.GeoCountry:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoCountry, spe.Geography.Country)
		case jdh.GeoState:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoState, spe.Geography.State)
		case jdh.GeoCounty:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoCounty, spe.Geography.County)
		case jdh.SpeLocality:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.SpeLocality, spe.Locality)
		case jdh.GeoLonLat:
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "%s=%.8f,%.8f\n", jdh.GeoLonLat, spe.Georef.Point.Lon, spe.Georef.Point.Lat)
			} else {
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoLonLat)
			}
		case jdh.GeoUncertainty:
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "%s=%d\n", jdh.GeoUncertainty, spe.Georef.Uncertainty)
			} else {
				fmt.Fprintf(os.Stdout, "%s=0\n", jdh.GeoUncertainty)
			}
		case jdh.GeoSource:
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoSource, spe.Georef.Source)
			} else {
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoSource)
			}
		case jdh.GeoValidation:
			if spe.Georef.IsValid() {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.GeoValidation, spe.Georef.Validation)
			} else {
				fmt.Fprintf(os.Stdout, "%s=\n", jdh.GeoValidation)
			}
		case jdh.KeyExtern:
			for _, e := range spe.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, spe.Comment)
		case jdh.KeyReference:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyReference, spe.Reference)
		}
		return
	}
	if len(keyFlag) == 0 {
		tax := taxon(c, db, spe.Taxon)
		var set *jdh.Dataset
		if len(spe.Dataset) > 0 {
			set = dataset(c, db, spe.Dataset)
		}
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", spe.Id)
		fmt.Fprintf(os.Stdout, "%-16s %s %s [id: %s]\n", "Taxon:", tax.Name, tax.Authority, spe.Taxon)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Basis:", spe.Basis)
		if len(spe.Reference) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Reference:", spe.Reference)
		}
		if set != nil {
			fmt.Fprintf(os.Stdout, "%-16s [id: %s]\n%s\n", "Dataset:", set.Id, set.Title)
		}
		if len(spe.Catalog) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Catalog:", spe.Catalog)
		}
		if len(spe.Determiner) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Determiner:", spe.Determiner)
		}
		if len(spe.Collector) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Collector:", spe.Collector)
		}
		if !spe.Date.IsZero() {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Date:", spe.Date.Format(jdh.Iso8601))
		}
		if len(spe.Geography.Country) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Country:", spe.Geography.Country.Name())
		}
		if len(spe.Geography.State) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "State:", spe.Geography.State)
		}
		if len(spe.Geography.County) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "County:", spe.Geography.County)
		}
		if len(spe.Locality) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Locality:", spe.Locality)
		}
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "%-16s %.8f, %.8f\n", "LonLat:", spe.Georef.Point.Lon, spe.Georef.Point.Lat)
			if spe.Georef.Uncertainty != 0 {
				fmt.Fprintf(os.Stdout, "%-16s %d\n", "Uncertainty:", spe.Georef.Uncertainty)
			}
			if len(spe.Georef.Source) > 0 {
				fmt.Fprintf(os.Stdout, "%-16s %s\n", "Source:", spe.Georef.Source)
			}
			if len(spe.Georef.Validation) > 0 {
				fmt.Fprintf(os.Stdout, "%-16s %s\n", "Validation:", spe.Georef.Validation)
			}
		}
		if len(spe.Extern) > 0 {
			fmt.Fprintf(os.Stdout, "Extern ids:\n")
			for _, e := range spe.Extern {
				fmt.Fprintf(os.Stdout, "\t%s\n", e)
			}
		}
		if len(spe.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", spe.Comment)
		}
		if set != nil {
			if len(set.Citation) > 0 {
				fmt.Fprintf(os.Stdout, "Citation:\n%s\n", set.Citation)
			}
			if len(set.License) > 0 {
				fmt.Fprintf(os.Stdout, "License:\n%s\n", set.License)
			}
		}
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.SpeTaxon:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Taxon)
	case jdh.SpeBasis:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Basis)
	case jdh.SpeDataset:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Dataset)
	case jdh.SpeCatalog:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Catalog)
	case jdh.SpeDeterminer:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Determiner)
	case jdh.SpeCollector:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Collector)
	case jdh.SpeDate:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Date.Format(jdh.Iso8601))
	case jdh.GeoCountry:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Geography.Country)
	case jdh.GeoState:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Geography.State)
	case jdh.GeoCounty:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Geography.County)
	case jdh.SpeLocality:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Locality)
	case jdh.GeoLonLat:
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "%.8f\t%.8f\n", spe.Georef.Point.Lon, spe.Georef.Point.Lat)
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	case jdh.GeoUncertainty:
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "%d\n", spe.Georef.Uncertainty)
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	case jdh.GeoSource:
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "%s\n", spe.Georef.Source)
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	case jdh.GeoValidation:
		if spe.Georef.IsValid() {
			fmt.Fprintf(os.Stdout, "%s\n", spe.Georef.Validation)
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	case jdh.KeyExtern:
		for _, e := range spe.Extern {
			fmt.Fprintf(os.Stdout, "%s\n", e)
		}
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Comment)
	case jdh.KeyReference:
		fmt.Fprintf(os.Stdout, "%s\n", spe.Reference)
	}
}
