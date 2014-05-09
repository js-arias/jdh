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

// Taxon gets a taxon.
func taxon(c *cmdapp.Command, db jdh.DB, id string) *jdh.Taxon {
	sc, err := db.Get(jdh.Taxonomy, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	tax := &jdh.Taxon{}
	if err := sc.Scan(tax); err != nil {
		if err == io.EOF {
			return &jdh.Taxon{}
		}
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return tax
}

// TaxInDB returns true if a taxon is in the database.
func taxInDB(c *cmdapp.Command, db jdh.DB, name, parent string, rank jdh.Rank, valid bool) *jdh.Taxon {
	args := new(jdh.Values)
	args.Add(jdh.TaxName, name)
	if len(parent) != 0 {
		args.Add(jdh.TaxParent, parent)
	}
	if rank != jdh.Unranked {
		args.Add(jdh.TaxRank, rank.String())
	}
	l, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer l.Close()
	for {
		tax := &jdh.Taxon{}
		if err := l.Scan(tax); err != nil {
			if err == io.EOF {
				return nil
			}
		}
		if len(tax.Id) > 0 {
			if tax.IsValid == valid {
				return tax
			}
		}
	}
}

// PickTaxName search for a unique taxon name. If there are more taxons
// fullfilling the name, then it will print a list of the potential
// names and finish the program.
func pickTaxName(c *cmdapp.Command, db jdh.DB, name, parent string) *jdh.Taxon {
	args := new(jdh.Values)
	args.Add(jdh.TaxName, name)
	if len(parent) != 0 {
		args.Add(jdh.TaxParentName, parent)
	}
	l, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	var tax *jdh.Taxon
	mult := false
	for {
		ot := &jdh.Taxon{}
		if err := l.Scan(ot); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if tax == nil {
			tax = ot
			continue
		}
		if !mult {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("ambiguos taxon name"))
			fmt.Fprintf(os.Stderr, "%s\t%s\n", tax.Id, tax.Name)
			mult = true
		}
		fmt.Fprintf(os.Stderr, "%s\t%s\n", ot.Id, ot.Name)
	}
	if mult {
		os.Exit(0)
	}
	if tax == nil {
		return &jdh.Taxon{}
	}
	return tax
}

// GetTaxDesc return the list of descendants of a taxon.
func getTaxDesc(c *cmdapp.Command, db jdh.DB, id string, valid bool) jdh.ListScanner {
	args := new(jdh.Values)
	if valid {
		args.Add(jdh.TaxChildren, id)
	} else {
		args.Add(jdh.TaxSynonyms, id)
	}
	l, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return l
}

// IsInParentName search a name if the parent taxons of a taxon.
func isInParentName(c *cmdapp.Command, db jdh.DB, id, parent string) bool {
	args := new(jdh.Values)
	args.Add(jdh.TaxParents, id)
	pl, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer pl.Close()
	for {
		p := &jdh.Taxon{}
		if err := pl.Scan(p); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if p.Name == parent {
			return true
		}
	}
	return false
}

// IsInParentList search an id in the list of parents of the taxon.
func isInParentList(c *cmdapp.Command, db jdh.DB, id string, pIds []string) bool {
	args := new(jdh.Values)
	args.Add(jdh.TaxParents, id)
	pl, err := db.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer pl.Close()
	for {
		p := &jdh.Taxon{}
		if err := pl.Scan(p); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		for _, pid := range pIds {
			if p.Id == pid {
				return true
			}
		}
	}
	return false
}
