// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package jdh

import (
	"strings"
)

// Taxon holds the taxonomic information of a taxon.
type Taxon struct {
	// identifier of the taxon.
	Id string

	// taxon's canonical name.
	Name string

	// taxon's authority citation.
	Authority string

	// taxon's rank id.
	Rank Rank

	// true if the taxon is valid/accepted.
	IsValid bool

	// id of the parent taxon.
	Parent string

	// extern identifiers of the taxon.
	Extern []string

	// free text comment about the taxon.
	Comment string
}

// Rank is a rank value. Ranks are arranged in a way that an inclusive rank
// in the taxonomy is always smaller than more exclusive ranks. Then is
// possible to use this form:
//
//     if rank < jdh.Genus {
//         // do something
//     }
type Rank uint

// Valid taxonomic ranks.
const (
	Unranked Rank = iota
	Kingdom
	Phylum
	Class
	Order
	Family
	Genus
	Species
)

// Ranks holds a list of the ranks accepted in jdh.
var ranks = []string{
	"unranked",
	"kingdom",
	"phylum",
	"class",
	"order",
	"family",
	"genus",
	"species",
}

// GetRank returns a rank id from a string.
func GetRank(s string) Rank {
	s = strings.ToLower(s)
	for i, r := range ranks {
		if r == s {
			return Rank(i)
		}
	}
	return Unranked
}

// String returns the rank string of a given rank id.
func (r Rank) String() string {
	i := int(r)
	if i >= len(ranks) {
		return ranks[0]
	}
	return ranks[i]
}

// Taxonomy is the table that store taxon information.
const Taxonomy Table = "taxonomy"

// Key values used in Taxonomy table.
const (
	// Authority asociated with the taxon.
	TaxAuthority Key = "authority"

	// Used in list operation to retrieve taxon's valid children.
	// The value is the id of the taxon.
	TaxChildren = "children"

	// Name of the taxon. In set operation, empty values are not
	// accepted. In list operation, if the name ends with an
	// asterisk ("*"), the name will be interpreted as a prefix.
	TaxName = "name"

	// Taxon's parent in set operation. In list opertation, is used
	// as filter to select only taxons descendants of the indicated
	// taxon.
	TaxParent = "parent"

	// Used in list operation as a filter to select only taxons that
	// have a parent with a given parent name.
	TaxParentName = "parentName"

	// Used in list operation to retrieve taxon's parents. The value
	// is the id of the taxon.
	TaxParents = "parents"

	// Rank of the taxon. It must be the string expression of a
	// rank accepted in jdh.
	TaxRank = "rank"

	// Set a taxon as a synonym of the taxon id indicated in value.
	// Only used during set opertation. If the value is empty, it will
	// be assumed that the parent of the taxon its is new senior
	// synonym.
	TaxSynonym = "synonym"

	// Used in list operation to retrieve taxon's synonyms. The value
	// is the id of the taxon.
	TaxSynonyms = "synonyms"

	// Validity of the taxon's name. Only using during set operation
	// an will always set a taxon as valid (the value field will be
	// ignored). The taxon will be set as valid, and sister of its
	// previous senior.
	TaxValid = "valid"
)
