// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package jdh

// Node is a node of a phylogenetic tree.
type Node struct {
	// identifier of the node.
	Id string

	// tree that contains the node.
	Tree string

	// taxon assigned to the node (if any).
	Taxon string

	// parent node.
	Parent string

	// Length of the node.
	Len uint

	// Age of the node, counted down from the present (present == 0).
	Age uint

	// free text comment about the node.
	Comment string
}

// Phylogeny is a phylogenetic tree.
type Phylogeny struct {
	// identifier of the tree.
	Id string

	// name of the tree.
	Name string

	// id of the root node of the tree.
	Root string

	// extern identifier of the tree.
	Extern []string

	// free text comment about the tree.
	Comment string
}

// Tables used for phylogenetic tree data.
const (
	// Used to retrieve, set information about a tree
	Trees Table = "trees"

	// Used to retrieve, set information about a node.
	Nodes = "nodes"
)

// Key values used for a phylogenetic tree node.
const (
	// The age of the node, an integer value.
	NodAge Key = "age"

	// Used in list will retrieve the list of descendants of the node.
	NodChildren = "children"

	// Used in set operation to collapse a node.
	NodCollapse = "collapse"

	// The length of the  branch that connects the node with its ancestor.
	// An integer value.
	NodLength = "length"

	// Id of the parent node of a node. In list operation will return the
	// list of ancestors of a node.
	NodParent = "parent"

	// Taxon id of the taxon associated with the node. Used in delete
	// will removes the association (but not the node). Used in list will
	// retrieve the node associated with the indicated taxon.
	NodTaxon = "taxon"

	// In list operations it retrieves all the nodes of a tree.
	NodTree = "nodtree"

	// In set operation, will move a node to be sister of the indicated
	// taxon.
	NodSister = "sister"
)

// Key values used for a phylogenetic tree.
const (
	// Used to set the name of the tree.
	TreName Key = "name"

	// Used in delete operation to remove all the references to a taxon
	// in the trees.
	TreTaxon = "phytaxon"
)
