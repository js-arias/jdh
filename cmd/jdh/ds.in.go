// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var dsIn = &cmdapp.Command{
	Name: "ds.in",
	Synopsis: `[-f|--format value] [-p|--port value] [-v|--verbose]
	[<file>...]`,
	Short: "imports dataset data",
	Long: `
Description

Ds.in reads dataset data form the indicated files, or standard input (if
no file is defined), and adds them to the jdh database.

Default input format is txt. If the format is txt, it is assummed that each 
line corresponds to a dataset (lines starting with '#' or ';' will be ignored).

Options

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -v
    --verbose
      If set, the name and id of each added taxon will be print in the
      standard output.
    
    <file>
      One or more files to be proccessed by tx.in. If no file is given
      then the information is expected to be from the standard input.
	`,
}

func init() {
	dsIn.Flag.StringVar(&formatFlag, "format", "", "")
	dsIn.Flag.StringVar(&formatFlag, "f", "", "")
	dsIn.Flag.StringVar(&portFlag, "port", "", "")
	dsIn.Flag.StringVar(&portFlag, "p", "", "")
	dsIn.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	dsIn.Flag.BoolVar(&verboseFlag, "v", false, "")
	dsIn.Run = dsInRun
}

func dsInRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	format := "txt"
	if len(formatFlag) > 0 {
		format = formatFlag
	}
	if len(args) > 0 {
		switch format {
		case "txt":
			for _, fname := range args {
				dsInTxt(c, fname)
			}
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	} else {
		switch format {
		case "txt":
			dsInTxt(c, "")
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func dsInTxt(c *cmdapp.Command, fname string) {
	var in *bufio.Reader
	if len(fname) > 0 {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			return
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}
	for {
		tn, err := readLine(in)
		if err != nil {
			break
		}
		ln := strings.Join(tn, " ")
		set := &jdh.Dataset{
			Title: ln,
		}
		if _, err = localDB.Exec(jdh.Add, jdh.Datasets, set); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			continue
		}
	}
}
