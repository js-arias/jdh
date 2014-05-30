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

var trLs = &cmdapp.Command{
	Name: "tr.ls",
	Synopsis: `[-a|--ancs] [-n|--node value] [-m|--machine]
	[-p|--port value] [-v|--verbose]`,
	Short:    "prints a list of trees or nodes",
	IsCommon: true,
	Long: `
Description

Tr.ls prints the list of trees, or the nodes of a tree, in the database. With
no option, tr.ls will print the list of trees, with the -i, --id option it
will print the descendants, or with the -a, --anc option the ancestors, of a
particular node.

Options

    -a
    --ancs
      If set, and the option -i, --id is set, the parents of the indicated
      node will be printed.
    
    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -n value
    --node value
      Search for the indicated node id.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -v
    --verbose
      If defined, then a large list will be printed. This option is ignored
      if -m or --machine option is defined.
	`,
}

func init() {
	trLs.Flag.BoolVar(&ancsFlag, "ancs", false, "")
	trLs.Flag.BoolVar(&ancsFlag, "a", false, "")
	trLs.Flag.StringVar(&nodeFlag, "node", "", "")
	trLs.Flag.StringVar(&nodeFlag, "n", "", "")
	trLs.Flag.BoolVar(&machineFlag, "machine", false, "")
	trLs.Flag.BoolVar(&machineFlag, "m", false, "")
	trLs.Flag.StringVar(&portFlag, "port", "", "")
	trLs.Flag.StringVar(&portFlag, "p", "", "")
	trLs.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	trLs.Flag.BoolVar(&verboseFlag, "v", false, "")
	trLs.Run = trLsRun
}

func trLsRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if len(nodeFlag) > 0 {
		trLsNodes(c)
		return
	}
	l, err := localDB.List(jdh.Trees, new(jdh.Values))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		phy := &jdh.Phylogeny{}
		if err := l.Scan(phy); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", phy.Id)
			continue
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t%s\troot: %s\n", phy.Id, phy.Name, phy.Root)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\troot: %s\n", phy.Name, phy.Root)
	}
}

func trLsNodes(c *cmdapp.Command) {
	vals := new(jdh.Values)
	if ancsFlag {
		vals.Add(jdh.NodParent, nodeFlag)
	} else {
		vals.Add(jdh.NodChildren, nodeFlag)
	}
	l, err := localDB.List(jdh.Nodes, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		nod := &jdh.Node{}
		if err := l.Scan(nod); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if machineFlag {
			fmt.Fprintf(os.Stdout, "%s\n", nod.Id)
			continue
		}
		var tax *jdh.Taxon
		if len(nod.Taxon) > 0 {
			tax = taxon(c, localDB, nod.Taxon)
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t", nod.Id)
			if tax != nil {
				fmt.Fprintf(os.Stdout, "%s [id:%s]", tax.Name, tax.Id)
			}
			fmt.Fprintf(os.Stdout, "\tlen: %d\tage: %d\n", nod.Len, nod.Age)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\t", nod.Id)
		if tax != nil {
			fmt.Fprintf(os.Stdout, "%s", tax.Name)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}
}
