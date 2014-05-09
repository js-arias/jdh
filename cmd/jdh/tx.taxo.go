// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var txTaxo = &cmdapp.Command{
	Name: "tx.taxo",
	Synopsis: `[-e|--extdb name] [-f|--format name]	[-i|--id value]
	[-p|--port value] [-s|--simple]	[<name> [<parentname>]]`,
	Short: "prints taxonomy",
	Long: `
Description

Tx.taxo prints the taxonomy of the indicated taxon in the format of a
taxonomic catalog.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          ncbi    taxonomy from ncbi (genbank).
    
    -f
    --format
      Sets the output format, by default it will use txt format.
      Valid values are:
          txt     text format
          html    html format 

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
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
	txTaxo.Flag.StringVar(&extDBFlag, "extdb", "", "")
	txTaxo.Flag.StringVar(&extDBFlag, "e", "", "")
	txTaxo.Flag.StringVar(&formatFlag, "format", "", "")
	txTaxo.Flag.StringVar(&formatFlag, "f", "", "")
	txTaxo.Flag.StringVar(&idFlag, "id", "", "")
	txTaxo.Flag.StringVar(&idFlag, "i", "", "")
	txTaxo.Flag.StringVar(&portFlag, "port", "", "")
	txTaxo.Flag.StringVar(&portFlag, "p", "", "")
	txTaxo.Run = txTaxoRun
}

func txTaxoRun(c *cmdapp.Command, args []string) {
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}
	var tax *jdh.Taxon
	if len(idFlag) > 0 {
		tax = taxon(c, db, idFlag)
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
		tax = &jdh.Taxon{}
	}
	if len(formatFlag) > 0 {
		switch formatFlag {
		case "txt":
		case "html":
			fmt.Fprintf(os.Stdout, "<html>\n")
			fmt.Fprintf(os.Stdout, "<head><meta http-equiv=\"Content-Type\" content=\"text/html\" charset=utf-8\" /></head>\n")
			fmt.Fprintf(os.Stdout, "<body bgcolor=\"white\">\n<font face=\"sans-serif\"><pre>\n")
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("unknown format"))
			os.Exit(1)
		}
	} else {
		formatFlag = "txt"
	}
	txTaxoProc(c, db, tax, jdh.Kingdom)
	if formatFlag == "html" {
		fmt.Fprintf(os.Stdout, "</pre></font>\n</body>\n</html>\n")
	}
}

func txTaxoProc(c *cmdapp.Command, db jdh.DB, tax *jdh.Taxon, prevRank jdh.Rank) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
	if len(tax.Id) != 0 {
		txTaxoPrint(c, db, tax, r)
	}
	l := getTaxDesc(c, db, tax.Id, true)
	defer l.Close()
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txTaxoProc(c, db, desc, r)
	}
}

func txTaxoPrint(c *cmdapp.Command, db jdh.DB, tax *jdh.Taxon, prevRank jdh.Rank) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
	serv := extDBFlag
	if len(serv) == 0 {
		serv = "jdh"
	}
	serv += ":"
	l := getTaxDesc(c, db, tax.Id, false)
	defer l.Close()
	if r < jdh.Species {
		nm := strings.ToTitle(tax.Name)
		fmt.Fprintf(os.Stdout, "\n")
		switch formatFlag {
		case "html":
			if tax.Rank != jdh.Unranked {
				fmt.Fprintf(os.Stdout, "%s <strong>%s</strong> %s [%s]\n", html.EscapeString(strings.Title(tax.Rank.String())), html.EscapeString(nm), html.EscapeString(tax.Authority), html.EscapeString(serv+tax.Id))
			} else {
				fmt.Fprintf(os.Stdout, "<strong>%s</strong> %s [%s]\n", html.EscapeString(nm), html.EscapeString(tax.Authority), html.EscapeString(serv+tax.Id))
			}
			for {
				s := &jdh.Taxon{}
				if err := l.Scan(s); err != nil {
					if err == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					os.Exit(1)
				}
				fmt.Fprintf(os.Stdout, "<font color=\"gray\">%s %s [%s]</font>\n", html.EscapeString(s.Name), html.EscapeString(s.Authority), html.EscapeString(serv+s.Id))
			}
		case "txt":
			if tax.Rank != jdh.Unranked {
				fmt.Fprintf(os.Stdout, "%s %s %s [%s]\n", strings.Title(tax.Rank.String()), nm, tax.Authority, serv+tax.Id)
			} else {
				fmt.Fprintf(os.Stdout, "%s %s [%s]\n", nm, tax.Authority, serv+tax.Id)
			}
			for {
				s := &jdh.Taxon{}
				if err := l.Scan(s); err != nil {
					if err == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					os.Exit(1)
				}
				fmt.Fprintf(os.Stdout, "%s %s [%s]\n", s.Name, s.Authority, serv+s.Id)
			}
		}
		fmt.Fprintf(os.Stdout, "\n")
		return
	}
	// species or below
	switch formatFlag {
	case "html":
		fmt.Fprintf(os.Stdout, "\t<i>%s</i> %s [%s]\n", html.EscapeString(tax.Name), html.EscapeString(tax.Authority), html.EscapeString(serv+tax.Id))
		for {
			s := &jdh.Taxon{}
			if err := l.Scan(s); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "\t\t<font color=\"gray\"><i>%s</i> %s [%s]</font>\n", html.EscapeString(s.Name), html.EscapeString(s.Authority), html.EscapeString(serv+s.Id))
		}
	case "txt":
		fmt.Fprintf(os.Stdout, "%s %s [%s]\n", tax.Name, tax.Authority, serv+tax.Id)
		for {
			s := &jdh.Taxon{}
			if err := l.Scan(s); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "%s %s [%s]\n", s.Name, s.Authority, serv+s.Id)
		}
	}
}
