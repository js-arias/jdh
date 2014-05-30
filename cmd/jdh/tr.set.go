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

var trSet = &cmdapp.Command{
	Name: "tr.set",
	Synopsis: `[-i|--id value] [-n|--node value] [-p|--port value]
	[<key=value>...]`,
	Short: "set a tree or node value",
	Long: `
Description

Tr.set sets a particular value for a tree (with the -i, --id option) or a node
(with the -n, --node option) in the database. Use this command to edit the tree
database, instead of manual edition.

If no key is defined, the key values will be read from the standard input, 
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

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

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "name=Strict consensus tree".
      Valid keys when a a tree is searched (-i, --id option) are:
          comment        A free text comment on the tree.
          extern         Extern identifiers of the tree, in the form
                         <service>:<key>.
          name           Name of the tree.
      
      Valid keys when a a node is searched (-n, --node option) are:
          age            Age of the node.
          comment        A free text comment on the node.
          lenght         The length of the branch leading to the node.
          sister         move a node to be a sister of the indicated node.
          taxon          The taxon asociated with the node.
	`,
}

func init() {
	trSet.Flag.StringVar(&idFlag, "id", "", "")
	trSet.Flag.StringVar(&idFlag, "i", "", "")
	trSet.Flag.StringVar(&nodeFlag, "node", "", "")
	trSet.Flag.StringVar(&nodeFlag, "n", "", "")
	trSet.Flag.StringVar(&portFlag, "port", "", "")
	trSet.Flag.StringVar(&portFlag, "p", "", "")
	trSet.Run = trSetRun
}

func trSetRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if len(nodeFlag) > 0 {
		trSetNode(c, args)
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
	vals := valsFromArgs(phy.Id, args)
	localDB.Exec(jdh.Set, jdh.Trees, vals)
	localDB.Exec(jdh.Commit, "", nil)
}

func trSetNode(c *cmdapp.Command, args []string) {
	nod := phyloNode(c, localDB, nodeFlag)
	if len(nod.Id) == 0 {
		return
	}
	vals := valsFromArgs(nod.Id, args)
	localDB.Exec(jdh.Set, jdh.Trees, vals)
	localDB.Exec(jdh.Commit, "", nil)
}
