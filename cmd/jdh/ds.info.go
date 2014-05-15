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

var dsInfo = &cmdapp.Command{
	Name: "ds.info",
	Synopsis: `-i|--id value [-e|--extdb name] [-k|--key value]
	[-m|--machine] [-p|--port value]`,
	Short: "prints dataset information",
	Long: `
Description

Ds.info prints general information of a dataset in the database.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    datasets from gbif.
    
    -i value
    --id value
      Search for the indicated dataset id. It is a required option.
    
    -k value
    --key value
      If set, only a particular value of the taxon will be printed.
      Valid keys are:
          citation       Preferred citation of the dataset.
          comment        A free text comment on the dataset.
          license        License of use of the data.
          extern         Extern identifiers of the dataset, in the form
                         <service>:<key>.
          title          Title of the dataset.
          url            Url of the dataset.

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
	dsInfo.Flag.StringVar(&extDBFlag, "extdb", "", "")
	dsInfo.Flag.StringVar(&extDBFlag, "e", "", "")
	dsInfo.Flag.StringVar(&idFlag, "id", "", "")
	dsInfo.Flag.StringVar(&idFlag, "i", "", "")
	dsInfo.Flag.StringVar(&keyFlag, "key", "", "")
	dsInfo.Flag.StringVar(&keyFlag, "k", "", "")
	dsInfo.Flag.BoolVar(&machineFlag, "machine", false, "")
	dsInfo.Flag.BoolVar(&machineFlag, "m", false, "")
	dsInfo.Flag.StringVar(&portFlag, "port", "", "")
	dsInfo.Flag.StringVar(&portFlag, "p", "", "")
	dsInfo.Run = dsInfoRun
}

func dsInfoRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong dataset id"))
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
	set := dataset(c, db, idFlag)
	if len(set.Id) == 0 {
		return
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataTitle, set.Title)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataCitation, set.Citation)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataLicense, set.License)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataUrl, set.Url)
			for _, e := range set.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, set.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.DataTitle:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataTitle, set.Title)
		case jdh.DataCitation:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataCitation, set.Citation)
		case jdh.DataLicense:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataLicense, set.License)
		case jdh.DataUrl:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.DataUrl, set.Url)
		case jdh.KeyExtern:
			for _, e := range set.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, set.Comment)
		}
		return
	}
	if len(keyFlag) == 0 {
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", set.Id)
		fmt.Fprintf(os.Stdout, "Title:\n%s\n", set.Title)
		if len(set.Url) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Url:", set.Url)
		}
		if len(set.Citation) > 0 {
			fmt.Fprintf(os.Stdout, "Citation:\n%s\n", set.Citation)
		}
		if len(set.License) > 0 {
			fmt.Fprintf(os.Stdout, "License:\n%s\n", set.License)
		}
		if len(set.Extern) > 0 {
			fmt.Fprintf(os.Stdout, "Extern ids:\n")
			for _, e := range set.Extern {
				fmt.Fprintf(os.Stdout, "\t%s\n", e)
			}
		}
		if len(set.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", set.Comment)
		}
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.DataTitle:
		fmt.Fprintf(os.Stdout, "%s\n", set.Title)
	case jdh.DataCitation:
		fmt.Fprintf(os.Stdout, "%s\n", set.Citation)
	case jdh.DataLicense:
		fmt.Fprintf(os.Stdout, "%s\n", set.License)
	case jdh.DataUrl:
		fmt.Fprintf(os.Stdout, "%s\n", set.Url)
	case jdh.KeyExtern:
		for _, e := range set.Extern {
			fmt.Fprintf(os.Stdout, "%s\n", e)
		}
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", set.Comment)
	}
}
