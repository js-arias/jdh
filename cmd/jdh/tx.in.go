// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var txIn = &cmdapp.Command{
	Name: "tx.in",
	Synopsis: `[-a|--anc value] [-f|--format value] [-p|--port value]
	[-r|--rank name] [-s|--synonym] [-v|--verbose] [<file>...]`,
	Short: "imports taxon data",
	Long: `
Description

Tx.in reads taxon data from the indicated files, or the standard input
(if no file is defined), and adds them to the jdh database.

If the format is txt, it is assummed that each line corresponds to a taxon
(lines starting with '#' or ';' will be ignored).

By default, taxons will be added to the root of the taxonomy, valid, and
unranked.

Options

    -a value
    --anc value
      Sets the parent of the added taxons. The value must be a valid id.

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      Set the rank of the added taxon. If the taxon has a parent (the -a,
      --anc options) the parent must be concordant with the given rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

    -s
    --synonym
      If set, the added taxons will be set as synonym of its parent. It
      requires that a valid parent will be defined (-n, --pname, -p or
      --parent options)

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
	txIn.Flag.StringVar(&ancFlag, "anc", "", "")
	txIn.Flag.StringVar(&ancFlag, "a", "", "")
	txIn.Flag.StringVar(&formatFlag, "format", "", "")
	txIn.Flag.StringVar(&formatFlag, "f", "", "")
	txIn.Flag.StringVar(&portFlag, "port", "", "")
	txIn.Flag.StringVar(&portFlag, "p", "", "")
	txIn.Flag.StringVar(&rankFlag, "rank", "", "")
	txIn.Flag.StringVar(&rankFlag, "r", "", "")
	txIn.Flag.BoolVar(&synonymFlag, "synonym", false, "")
	txIn.Flag.BoolVar(&synonymFlag, "s", false, "")
	txIn.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	txIn.Flag.BoolVar(&verboseFlag, "v", false, "")
	txIn.Run = txInRun
}

func txInRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	pId := ""
	if len(ancFlag) > 0 {
		p := taxon(c, localDB, ancFlag)
		if !p.IsValid {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("taxon "+p.Name+" a synonym, can ot be a parent"))
			os.Exit(1)
		}
		pId = p.Id
	}
	valid := true
	if synonymFlag {
		if len(pId) == 0 {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("synonym without defined parent"))
			os.Exit(2)
		}
		valid = false
	}
	rank := jdh.Unranked
	if len(rankFlag) > 0 {
		rank = jdh.GetRank(rankFlag)
	}
	if len(args) > 0 {
		for _, fn := range args {
			if (len(formatFlag) == 0) || (formatFlag == "txt") {
				procTxInTxt(c, fn, pId, rank, valid)
			}
		}
	} else {
		if (len(formatFlag) == 0) || (formatFlag == "txt") {
			procTxInTxt(c, "", pId, rank, valid)
		}
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func procTxInTxt(c *cmdapp.Command, fname, parent string, rank jdh.Rank, valid bool) {
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
		nm := strings.Join(tn, " ")
		if r := []rune(nm); !unicode.IsLetter(r[0]) {
			continue
		}
		// skip names already in the database
		if taxInDB(c, localDB, nm, parent, rank, valid) != nil {
			continue
		}
		pId := parent
		// we know that the first name of a species is the genus.
		if rank == jdh.Species {
			if p := taxInDB(c, localDB, tn[0], parent, jdh.Genus, true); p != nil {
				pId = p.Id
			} else {
				par := &jdh.Taxon{
					Name:    tn[0],
					IsValid: true,
					Parent:  parent,
					Rank:    jdh.Genus,
				}
				pId, err = localDB.Exec(jdh.Add, jdh.Taxonomy, par)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
			}
		}
		tax := &jdh.Taxon{
			Name:    nm,
			IsValid: valid,
			Parent:  pId,
			Rank:    rank,
		}
		id, err := localDB.Exec(jdh.Add, jdh.Taxonomy, tax)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			continue
		}
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "%s\t%s\n", id, tax.Name)
		}
	}
}
