// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"container/list"
	"fmt"
	"io"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var trForce = &cmdapp.Command{
	Name:     "tr.force",
	Synopsis: `[-i|--id value] [-p|--port value] [-r|--report]`,
	Short:    "enforces valid taxons as tree terminals",
	Long: `
Description

Tr.force enforces the terminals of the trees of the database to be all valid
taxons. If the taxon is not presented in the tree, then the invalid terminal
will be replaced with the valid taxon, if the valid taxon is already in the
tree, then, the invalid taxon will be deleted.

If the option -r, --report is set, it only show what terminals should be
removed without performing any operation.

Options

    -i value
    --id value
      Search for the indicated tree id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r
    --report
      If set, then only the modifications will be printed, without doing any
      operation.
	`,
}

func init() {
	trForce.Flag.StringVar(&idFlag, "id", "", "")
	trForce.Flag.StringVar(&idFlag, "i", "", "")
	trForce.Flag.StringVar(&portFlag, "port", "", "")
	trForce.Flag.StringVar(&portFlag, "p", "", "")
	trForce.Flag.BoolVar(&repFlag, "report", false, "")
	trForce.Flag.BoolVar(&repFlag, "r", false, "")
	trForce.Run = trForceRun
}

func trForceRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if len(idFlag) > 0 {
		phy := phylogeny(c, localDB, idFlag)
		if len(phy.Id) == 0 {
			return
		}
		trForceProc(c, phy)
		localDB.Exec(jdh.Commit, "", nil)
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
		if len(phy.Id) == 0 {
			continue
		}
		trForceProc(c, phy)
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func trForceProc(c *cmdapp.Command, phy *jdh.Phylogeny) {
	txLs := list.New()
	vals := new(jdh.Values)
	vals.Add(jdh.TreTaxon, phy.Id)
	l, err := localDB.List(jdh.Trees, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		var tId jdh.IdElement
		if err := l.Scan(&tId); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if len(tId.Id) == 0 {
			continue
		}
		txLs.PushBack(tId.Id)
	}
	if txLs.Len() == 0 {
		return
	}
	root := phyloNode(c, localDB, phy.Root)
	trForceNode(c, root, txLs)
}

func trForceNode(c *cmdapp.Command, nod *jdh.Node, txLs *list.List) {
	vals := new(jdh.Values)
	vals.Add(jdh.NodChildren, nod.Id)
	l, err := localDB.List(jdh.Nodes, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	childs := 0
	for {
		desc := &jdh.Node{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		trForceNode(c, desc, txLs)
		childs++
	}
	if childs > 1 {
		return
	}
	if len(nod.Taxon) == 0 {
		return
	}
	tax := taxon(c, localDB, nod.Taxon)
	if tax.IsValid {
		return
	}
	par := taxon(c, localDB, tax.Parent)
	todel := false
	for e := txLs.Front(); e != nil; e = e.Next() {
		tId := e.Value.(string)
		if tId == par.Id {
			todel = true
			break
		}
	}
	if todel {
		fmt.Fprintf(os.Stdout, "%s: deleted\n", tax.Name)
		if !repFlag {
			vals.Reset()
			vals.Add(jdh.KeyId, nod.Id)
			localDB.Exec(jdh.Delete, jdh.Nodes, vals)
		}
		return
	}
	fmt.Fprintf(os.Stdout, "%s: changed to: %s\n", tax.Name, par.Name)
	if !repFlag {
		vals.Reset()
		vals.Add(jdh.KeyId, nod.Id)
		vals.Add(jdh.NodTaxon, par.Id)
		localDB.Exec(jdh.Set, jdh.Nodes, vals)
		txLs.PushBack(par.Id)
	}
}
