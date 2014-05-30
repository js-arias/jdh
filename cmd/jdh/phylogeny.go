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

// PhyloNode gets a node.
func phyloNode(c *cmdapp.Command, db jdh.DB, id string) *jdh.Node {
	sc, err := db.Get(jdh.Nodes, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	nod := &jdh.Node{}
	if err := sc.Scan(nod); err != nil {
		if err == io.EOF {
			return &jdh.Node{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return nod
}

// Phylogeny gets a phylogeny.
func phylogeny(c *cmdapp.Command, db jdh.DB, id string) *jdh.Phylogeny {
	sc, err := db.Get(jdh.Trees, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	phy := &jdh.Phylogeny{}
	if err := sc.Scan(phy); err != nil {
		if err == io.EOF {
			return &jdh.Phylogeny{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return phy
}
