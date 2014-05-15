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

var dsLs = &cmdapp.Command{
	Name: "ds.ls",
	Synopsis: `[-c|--citation] [-e|--extdb name] [-l|--license]
	[-m|--machine] [-p|--port value] [-u|--url] [-v|--verbose]`,
	Short: "prints a list of datasets",
	Long: `
Description

Ds.ls prints a list of datasets. With no option, ds.ls prints all the datasets
in the database.

Options

    -c
    --citation
      If set, citation information will be printed.
      
    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    datasets from gbif.
      
    -l
    --license
      If set, license information will be printed.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
    -u
    --url
      If set the url of the dataset will be printed.

    -v
    --verbose
      If defined, then a large list will be printed. This option is ignored
      if -m or --machine option is defined.
	`,
}

func init() {
	dsLs.Flag.BoolVar(&citFlag, "citation", false, "")
	dsLs.Flag.BoolVar(&citFlag, "c", false, "")
	dsLs.Flag.StringVar(&extDBFlag, "extdb", "", "")
	dsLs.Flag.StringVar(&extDBFlag, "e", "", "")
	dsLs.Flag.BoolVar(&licFlag, "license", false, "")
	dsLs.Flag.BoolVar(&licFlag, "l", false, "")
	dsLs.Flag.BoolVar(&machineFlag, "machine", false, "")
	dsLs.Flag.BoolVar(&machineFlag, "m", false, "")
	dsLs.Flag.StringVar(&portFlag, "port", "", "")
	dsLs.Flag.StringVar(&portFlag, "p", "", "")
	dsLs.Flag.BoolVar(&urlFlag, "url", false, "")
	dsLs.Flag.BoolVar(&urlFlag, "u", false, "")
	dsLs.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	dsLs.Flag.BoolVar(&verboseFlag, "v", false, "")
	dsLs.Run = dsLsRun
}

func dsLsRun(c *cmdapp.Command, args []string) {
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}
	l, err := db.List(jdh.Datasets, new(jdh.Values))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		set := &jdh.Dataset{}
		if err := l.Scan(set); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", set.Id)
			continue
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t%s\t%sn", set.Id, set.Title, set.Url)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s", set.Title)
		if urlFlag {
			fmt.Fprintf(os.Stdout, "\t%s", set.Url)
		}
		if citFlag {
			fmt.Fprintf(os.Stdout, "\t%s", set.Citation)
		}
		if licFlag {
			fmt.Fprintf(os.Stdout, "\t%s", set.License)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}
}
