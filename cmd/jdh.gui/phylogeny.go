// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"io"
	"os"
	"sort"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/sparta"
	"github.com/js-arias/sparta/widget"
)

// PhyloNode gets a node.
func phyloNode(c *cmdapp.Command, id string) *jdh.Node {
	sc, err := localDB.Get(jdh.Nodes, id)
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

type trList struct {
	phyLs []*jdh.Phylogeny
	pos   int
}

type trNode struct {
	// info
	id       string
	name     widget.Text
	taxon    string
	parent   *trNode
	children []*trNode

	// configuration
	terms uint
	level uint
	nest  uint

	// virtual position
	minX   float32
	maxX   float32
	startY float32
	endY   float32
	vY     float32

	// current position
	pos      image.Point
	ancLine  []image.Point
	descLine []image.Point
}

func setNode(nod *jdh.Node, anc *trNode, data *trData) *trNode {
	trn := &trNode{
		id:     nod.Id,
		taxon:  nod.Taxon,
		parent: anc,
	}
	data.node = append(data.node, trn)
	if anc != nil {
		trn.nest = anc.nest + 1
		trn.minX = float32(trn.nest)
	}
	if len(nod.Taxon) > 0 {
		tax := taxon(cmd, localDB, nod.Taxon)
		trn.name.Text = tax.Name
		if !tax.IsValid {
			val := taxon(cmd, localDB, tax.Parent)
			nm := fmt.Sprintf("[%s syn. of %s]", tax.Name, val.Name)
			trn.name.Text = nm
		}
	}
	vals := new(jdh.Values)
	vals.Add(jdh.NodChildren, nod.Id)
	l, err := localDB.List(jdh.Nodes, vals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", cmd.ErrStr(err))
		os.Exit(1)
	}
	for {
		desc := &jdh.Node{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", cmd.ErrStr(err))
			os.Exit(1)
		}
		d := setNode(desc, trn, data)
		trn.terms += d.terms
		if trn.level <= d.level {
			trn.level = d.level + 1
		}
		trn.children = append(trn.children, d)
	}
	if len(trn.children) == 0 {
		data.terms++
		trn.terms = 1
	} else {
		sort.Sort(bySize(trn.children))
	}
	return trn
}

func (trn *trNode) setNode(data *trData, nt int) int {
	root := data.node[0]
	trn.maxX = float32(root.level - trn.level)
	if len(trn.children) == 0 {
		t := float32(nt)
		trn.startY, trn.endY, trn.vY = t, t, t
		return nt + 1
	}
	maxY := float32(0)
	minY := float32(data.terms * 2)
	for _, d := range trn.children {
		nt = d.setNode(data, nt)
		if d.vY > maxY {
			maxY = d.vY
		}
		if d.vY < minY {
			minY = d.vY
		}
	}
	trn.startY = minY
	trn.endY = maxY
	trn.vY = ((maxY - minY) / 2) + minY
	return nt
}

func (trn *trNode) isValidSis(on *trNode) bool {
	if trn == on {
		return false
	}
	for p := trn.parent; p != nil; p = p.parent {
		if p == on {
			return false
		}
	}
	return true
}

type bySize []*trNode

func (b bySize) Len() int {
	return len(b)
}

func (b bySize) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b bySize) Less(i, j int) bool {
	return b[i].terms < b[j].terms
}

type trData struct {
	terms int
	x, y  float32
	pos   image.Point
	node  []*trNode
	aln   bool
	sel   *trNode
}

func (data *trData) putOnScreen() {
	for _, n := range data.node {
		var x int
		if data.aln {
			x = int(n.minX*data.x) + 5 + data.pos.X
		} else {
			x = int(n.maxX*data.x) + 5 + data.pos.X
		}
		y := int(n.vY*data.y) + 10 + data.pos.Y
		n.pos = image.Pt(x, y)
		p := n.parent
		if p != nil {
			var ancX int
			if data.aln {
				ancX = int(p.minX*data.x) + 5 + data.pos.X
			} else {
				ancX = int(p.maxX*data.x) + 5 + data.pos.X
			}
			n.ancLine = []image.Point{n.pos, image.Pt(ancX, y)}
		} else {
			n.ancLine = []image.Point{n.pos, image.Pt(x-5, y)}
		}
		if n.level > 0 {
			topY := int(n.startY*data.y) + 10 + data.pos.Y
			downY := int(n.endY*data.y) + 10 + data.pos.Y
			n.descLine = []image.Point{image.Pt(x, topY), image.Pt(x, downY)}
		}
		if len(n.name.Text) > 0 {
			x += sparta.WidthUnit
			y := int(n.vY*data.y) + 10 + data.pos.Y - (sparta.HeightUnit / 2)
			n.name.Pos = image.Pt(x, y)
		}
	}
}

func setTree(phy *jdh.Phylogeny, rect image.Rectangle) *trData {
	data := &trData{}
	nod := phyloNode(cmd, phy.Root)
	setNode(nod, nil, data)
	root := data.node[0]
	root.setNode(data, 0)
	data.y = float32(sparta.HeightUnit)
	data.x = float32(sparta.WidthUnit * 2)
	return data
}
