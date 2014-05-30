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

var trInfo = &cmdapp.Command{
	Name: "tr.info",
	Synopsis: `[-i|--id value] [-n|--node value] [-k|--key value]
	[-m|--machine] [-p|--port value]`,
	Short: "prints tree or node information",
	Long: `
Description

Tr.info prints general information of a phylogenetic tree (using the -i, --id
optin) or a node (using the -n, --node option) in a tree present in the
database.

Options

    -i value
    --id value
      Search for the indicated tree id.
    
    -n value
    --node value
      Search for the indicated node id.
    
    -k value
    --key value
      If set, only a particular value of the taxon will be printed.
      Valid keys when a a tree is searched (-i, --id option) are:
          comment        A free text comment on the tree.
          extern         Extern identifiers of the tree, in the form
                         <service>:<key>.
          name           Name of the tree.
      
      Valid keys when a a node is searched (-n, --node option) are:
          age            Age of the node.
          comment        A free text comment on the node.
          lenght         The length of the branch leading to the node.
          taxon          The taxon asociated with the node.
          
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
	trInfo.Flag.StringVar(&idFlag, "id", "", "")
	trInfo.Flag.StringVar(&idFlag, "i", "", "")
	trInfo.Flag.StringVar(&keyFlag, "key", "", "")
	trInfo.Flag.StringVar(&keyFlag, "k", "", "")
	trInfo.Flag.BoolVar(&machineFlag, "machine", false, "")
	trInfo.Flag.BoolVar(&machineFlag, "m", false, "")
	trInfo.Flag.StringVar(&nodeFlag, "node", "", "")
	trInfo.Flag.StringVar(&nodeFlag, "n", "", "")
	trInfo.Flag.StringVar(&portFlag, "port", "", "")
	trInfo.Flag.StringVar(&portFlag, "p", "", "")
	trInfo.Run = trInfoRun
}

func trInfoRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if len(nodeFlag) > 0 {
		trInfoNode(c)
		return
	}
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong tree or node id"))
		c.Usage()
	}
	phy := phylogeny(c, localDB, idFlag)
	if len(phy.Id) == 0 {
		return
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TreName, phy.Name)
			for _, e := range phy.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, phy.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.TreName:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.TreName, phy.Name)
		case jdh.KeyExtern:
			for _, e := range phy.Extern {
				fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyExtern, e)
			}
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, phy.Comment)
		}
		return
	}
	if len(keyFlag) == 0 {
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", phy.Id)
		if len(phy.Name) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Name:", phy.Name)
		}
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Root:", phy.Root)
		if len(phy.Extern) > 0 {
			fmt.Fprintf(os.Stdout, "Extern ids:\n")
			for _, e := range phy.Extern {
				fmt.Fprintf(os.Stdout, "\t%s\n", e)
			}
		}
		if len(phy.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", phy.Comment)
		}
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.TreName:
		fmt.Fprintf(os.Stdout, "%s\n", phy.Name)
	case jdh.KeyExtern:
		for _, e := range phy.Extern {
			fmt.Fprintf(os.Stdout, "%s\n", e)
		}
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", phy.Comment)
	}
}

func trInfoNode(c *cmdapp.Command) {
	nod := phyloNode(c, localDB, nodeFlag)
	if len(nod.Id) == 0 {
		return
	}
	if machineFlag {
		if len(keyFlag) == 0 {
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodTaxon, nod.Taxon)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodLength, nod.Len)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodAge, nod.Age)
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, nod.Comment)
			return
		}
		switch jdh.Key(keyFlag) {
		case jdh.NodTaxon:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodTaxon, nod.Taxon)
		case jdh.NodLength:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodLength, nod.Len)
		case jdh.NodAge:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.NodAge, nod.Age)
		case jdh.KeyComment:
			fmt.Fprintf(os.Stdout, "%s=%s\n", jdh.KeyComment, nod.Comment)
		}
		return
	}
	if len(keyFlag) == 0 {
		var tax *jdh.Taxon
		if len(nod.Taxon) > 0 {
			tax = taxon(c, localDB, nod.Taxon)
		}
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Id:", nod.Id)
		fmt.Fprintf(os.Stdout, "%-16s %s\n", "Tree:", nod.Tree)
		if tax != nil {
			fmt.Fprintf(os.Stdout, "%-16s %s %s [id: %s]\n", "Taxon:", tax.Name, tax.Authority, tax.Id)
		}
		if len(nod.Parent) > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %s\n", "Parent:", nod.Parent)
		} else {
			fmt.Fprintf(os.Stdout, "Root node\n")
		}
		if nod.Len > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %d\n", "Length:", nod.Len)
		}
		if nod.Age > 0 {
			fmt.Fprintf(os.Stdout, "%-16s %d\n", "Age:", nod.Age)
		}
		if len(nod.Comment) > 0 {
			fmt.Fprintf(os.Stdout, "Comments:\n%s\n", nod.Comment)
		}
		vals := new(jdh.Values)
		vals.Add(jdh.NodChildren, nod.Id)
		l, err := localDB.List(jdh.Nodes, vals)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		first := true
		for {
			desc := &jdh.Node{}
			if err := l.Scan(desc); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
			if first {
				fmt.Fprintf(os.Stdout, "Children nodes:\n")
				first = false
			}
			fmt.Fprintf(os.Stdout, "\t%s", desc.Id)
			if len(desc.Taxon) > 0 {
				tax := taxon(c, localDB, desc.Taxon)
				fmt.Fprintf(os.Stdout, "\t[%s [id: %s]]", tax.Name, tax.Id)
			}
			fmt.Fprintf(os.Stdout, "\n")
		}
		return
	}
	switch jdh.Key(keyFlag) {
	case jdh.NodTaxon:
		fmt.Fprintf(os.Stdout, "%s\n", nod.Taxon)
	case jdh.NodLength:
		fmt.Fprintf(os.Stdout, "%s\n", nod.Len)
	case jdh.NodAge:
		fmt.Fprintf(os.Stdout, "%s\n", nod.Age)
	case jdh.KeyComment:
		fmt.Fprintf(os.Stdout, "%s\n", nod.Comment)
	}
}
