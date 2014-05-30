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

var trDel = &cmdapp.Command{
	Name: "tr.del",
	Synopsis: `[-c|--collapse] [-i|--id value] [-n|--node value]
	[-p|--port value]`,
	Short: "deletes a tree or a node",
	Long: `
Description

Tr.del removes a tree (with the -i, --id option) or a node (with -n, --node
option) from the database. When deleting a node, it will delete it, and all
of its descendants, optionally, if the option -c, --collapse is defined, then
the node will be deleted, but their descendats will be assigned to its parent.

Options

    -c
    --collapse
      Collapse the node: all the descendants will be assigned to the ancestor
      of the node. If the node to be collapsed is the root node, then, the
      collapse will be ignored.

    -i value
    --id value
      Search for the indicated tree id.
    
    -n value
    --node value
      Search for the indicated node id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	trDel.Flag.BoolVar(&colpFlag, "collapse", false, "")
	trDel.Flag.BoolVar(&colpFlag, "c", false, "")
	trDel.Flag.StringVar(&idFlag, "id", "", "")
	trDel.Flag.StringVar(&idFlag, "i", "", "")
	trDel.Flag.StringVar(&nodeFlag, "node", "", "")
	trDel.Flag.StringVar(&nodeFlag, "n", "", "")
	trDel.Flag.StringVar(&portFlag, "port", "", "")
	trDel.Flag.StringVar(&portFlag, "p", "", "")
	trDel.Run = trDelRun
}

func trDelRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if len(nodeFlag) > 0 {
		trDelNode(c)
		return
	}
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong tree or node id"))
		c.Usage()
	}
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, idFlag)
	if _, err := localDB.Exec(jdh.Delete, jdh.Trees, vals); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func trDelNode(c *cmdapp.Command) {
	vals := new(jdh.Values)
	if colpFlag {
		vals.Add(jdh.NodCollapse, nodeFlag)
	} else {
		vals.Add(jdh.KeyId, nodeFlag)
	}
	if _, err := localDB.Exec(jdh.Delete, jdh.Nodes, vals); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	localDB.Exec(jdh.Commit, "", nil)
}
