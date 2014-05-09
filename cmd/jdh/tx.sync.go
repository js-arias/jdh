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

var txSync = &cmdapp.Command{
	Name: "tx.sync",
	Synopsis: `-e|--extdb name [-i|--id value] [-l|--populate name]
	[-m|--match] [-p|--port value] [-r|--rank name] [-u|--update]
	[-v|--verbose] [<name> [<parentname>]]`,
	Short:    "updates local database using an extern database",
	IsCommon: true,
	Long: `
Description

Tx.sync uses an extern database to update, validate or set the
local database.

When the options -i, --id or a name are used, the effect of the update
will only affect the indicated taxon and its descendants.

Except for the -m, --match option, all other options assume that the
extern ids for each taxon is already defined. The -m, --match option
will try to scan the extern database to match current taxon names with
names in the extern database.

Options

    -e name
    --extdb name
      Set the extern database.
      Valid values are:
          gbif    taxonomy from gbif.
          ncbi    taxonomy from ncbi (genbank).
      This parameter is required.
    
    -i value
    --id value
      Update only the indicated taxon (and its descendants).
      
    -l name
    --populate name
      If set, then the taxons below the indicated rank will be populated
      with all descendants (valid and invalid) from the extern database.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -m
    --match
      If set, it will search in the extern database for each name in the
      local database that has no assigned extern id.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set, then it moves taxons to their respective rank. If the parent
      taxon is not in the database it will be added. The name value
      indicates the maximum rank to search, for example '--rank=class' only
      assign taxons up to class rank.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species
  
    -u
    --update
      If found, it will update the authorship, rank, taxon status and source
      of a taxon, following the external database. For synonyms, new parents
      will be added as needed. 
      
    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id 
      are defined.      
	`,
}

func init() {
	txSync.Flag.StringVar(&extDBFlag, "extdb", "", "")
	txSync.Flag.StringVar(&extDBFlag, "e", "", "")
	txSync.Flag.StringVar(&idFlag, "id", "", "")
	txSync.Flag.StringVar(&idFlag, "i", "", "")
	txSync.Flag.BoolVar(&matchFlag, "match", false, "")
	txSync.Flag.BoolVar(&matchFlag, "m", false, "")
	txSync.Flag.StringVar(&popFlag, "populate", "", "")
	txSync.Flag.StringVar(&popFlag, "l", "", "")
	txSync.Flag.StringVar(&portFlag, "port", "", "")
	txSync.Flag.StringVar(&portFlag, "p", "", "")
	txSync.Flag.StringVar(&rankFlag, "rank", "", "")
	txSync.Flag.StringVar(&rankFlag, "r", "", "")
	txSync.Flag.BoolVar(&updateFlag, "update", false, "")
	txSync.Flag.BoolVar(&updateFlag, "u", false, "")
	txSync.Run = txSyncRun
}

func txSyncRun(c *cmdapp.Command, args []string) {
	if len(extDBFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong '--extdb' option"))
		c.Usage()
	}
	openLocal(c)
	openExt(c, extDBFlag, "")
	var tax *jdh.Taxon
	if len(idFlag) > 0 {
		tax = taxon(c, localDB, idFlag)
		if len(tax.Id) == 0 {
			return
		}
	} else if len(args) > 0 {
		if len(args) > 2 {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("too many arguments"))
			os.Exit(1)
		}
		pName := ""
		if len(args) > 1 {
			pName = args[1]
		}
		tax = pickTaxName(c, localDB, args[0], pName)
		if len(tax.Id) == 0 {
			return
		}
	} else {
		tax = &jdh.Taxon{}
	}
	noFlag := true
	if matchFlag {
		txSyncMatch(c, tax)
		localDB.Exec(jdh.Commit, "", nil)
		noFlag = false
	}
	if updateFlag {
		txSyncUpdate(c, tax)
		localDB.Exec(jdh.Commit, "", nil)
		noFlag = false
	}
	if len(rankFlag) > 0 {
		rank := jdh.GetRank(rankFlag)
		if rank == jdh.Unranked {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("invalid rank"))
			os.Exit(1)
		}
		txSyncRank(c, tax, rank)
		localDB.Exec(jdh.Commit, "", nil)
		noFlag = false
	}
	if len(popFlag) > 0 {
		rank := jdh.GetRank(popFlag)
		if rank == jdh.Unranked {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("invalid rank"))
			os.Exit(1)
		}
		prev := tax.Rank
		if prev == jdh.Unranked {
			prev = jdh.Kingdom
		}
		txSyncPop(c, tax, prev, rank)
		localDB.Exec(jdh.Commit, "", nil)
		noFlag = false
	}
	if noFlag {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("undefined action"))
		c.Usage()
	}
}

// TxSyncMatch implements the match option of tx.sync.
func txSyncMatch(c *cmdapp.Command, tax *jdh.Taxon) {
	defer func() {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncMatchNav(c, l)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncMatchNav(c, l)
	}()

	if len(tax.Id) == 0 {
		return
	}
	eid := searchExtern(extDBFlag, tax.Extern)
	if len(eid) > 0 {
		return
	}
	var pEids []string
	args := new(jdh.Values)
	args.Add(jdh.TaxParents, tax.Id)
	pl, err := localDB.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		p := &jdh.Taxon{}
		if err := pl.Scan(p); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		pe := searchExtern(extDBFlag, p.Extern)
		if len(pEids) == 0 {
			continue
		}
		pEids = append(pEids, pe)
	}
	args.Reset()
	args.Add(jdh.TaxName, tax.Name)
	l, err := extDB.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	var extTax *jdh.Taxon
	for {
		et := &jdh.Taxon{}
		if err := l.Scan(et); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if len(pEids) > 0 {
			if !isInParentList(c, extDB, et.Id, pEids) {
				continue
			}
		}
		if (len(tax.Authority) > 0) && (len(et.Authority) > 0) {
			if tax.Authority != et.Authority {
				continue
			}
		}
		if extTax == nil {
			extTax = et
			continue
		}
		// The name is ambiguous
		fmt.Fprintf(os.Stderr, "taxon name %s is ambiguous in %s\n", tax.Name, extDBFlag)
		l.Close()
		return
	}
	if extTax == nil {
		return
	}
	args.Reset()
	args.Add(jdh.KeyId, tax.Id)
	args.Add(jdh.KeyExtern, extDBFlag+":"+extTax.Id)
	localDB.Exec(jdh.Set, jdh.Taxonomy, args)
}

func txSyncMatchNav(c *cmdapp.Command, l jdh.ListScanner) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txSyncMatch(c, desc)
	}
}

// TxSyncUpdate implements the update option of tx.sync.
func txSyncUpdate(c *cmdapp.Command, tax *jdh.Taxon) {
	// first process the descendants (as it is possible that a taxon
	// will be set as synonym, and then thier descendant set will change)
	l := getTaxDesc(c, localDB, tax.Id, true)
	txSyncUpdateNav(c, l)
	l = getTaxDesc(c, localDB, tax.Id, false)
	txSyncUpdateNav(c, l)

	if len(tax.Id) == 0 {
		return
	}
	eid := searchExtern(extDBFlag, tax.Extern)
	if len(eid) == 0 {
		return
	}
	ext := taxon(c, extDB, eid)
	if len(ext.Id) == 0 {
		fmt.Fprintf(os.Stderr, "unable to retrieve %s:%s\n", tax.Extern, eid)
		return
	}
	args := new(jdh.Values)
	if len(ext.Authority) > 0 {
		args.Add(jdh.KeyId, tax.Id)
		args.Add(jdh.TaxAuthority, ext.Authority)
		localDB.Exec(jdh.Set, jdh.Taxonomy, args)
	}
	if tax.IsValid != ext.IsValid {
		args.Reset()
		args.Add(jdh.KeyId, tax.Id)
		if ext.IsValid {
			args.Add(jdh.TaxValid, "")
			localDB.Exec(jdh.Set, jdh.Taxonomy, args)
		} else {
			p := taxon(c, localDB, extDBFlag+":"+ext.Parent)
			if len(p.Id) == 0 {
				exPar := taxon(c, extDB, ext.Parent)
				p = txSyncAddToTaxonomy(c, exPar)
			}
			args.Add(jdh.TaxSynonym, p.Id)
			localDB.Exec(jdh.Set, jdh.Taxonomy, args)
		}
	}
	if ext.Rank != jdh.Unranked {
		args.Reset()
		args.Add(jdh.KeyId, tax.Id)
		args.Add(jdh.TaxRank, ext.Rank.String())
		localDB.Exec(jdh.Set, jdh.Taxonomy, args)
	}
	if len(ext.Comment) > 0 {
		if len(tax.Comment) > 0 {
			tax.Comment += "\n"
		}
		tax.Comment += ext.Comment
		args.Reset()
		args.Add(jdh.KeyId, tax.Id)
		args.Add(jdh.KeyComment, tax.Comment)
		localDB.Exec(jdh.Set, jdh.Taxonomy, args)
	}
}

func txSyncUpdateNav(c *cmdapp.Command, l jdh.ListScanner) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txSyncUpdate(c, desc)
	}
}

func txSyncAddToTaxonomy(c *cmdapp.Command, src *jdh.Taxon) *jdh.Taxon {
	dest := &jdh.Taxon{}
	*dest = *src
	dest.Id = ""
	dest.Parent = getValidParent(c, src.Id)
	dest.Extern = []string{extDBFlag + ":" + src.Id}
	var err error
	dest.Id, err = localDB.Exec(jdh.Add, jdh.Taxonomy, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	return dest
}

func getValidParent(c *cmdapp.Command, extId string) string {
	args := new(jdh.Values)
	args.Add(jdh.TaxParents, extId)
	l, err := extDB.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer l.Close()
	for {
		et := &jdh.Taxon{}
		if err := l.Scan(et); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		p := taxon(c, localDB, extDBFlag+":"+et.Id)
		if len(p.Id) > 0 {
			if p.IsValid {
				return p.Id
			} else {
				return p.Parent
			}
		}
	}
	return ""
}

// txSyncRank implements rank option of tx.sync.
func txSyncRank(c *cmdapp.Command, tax *jdh.Taxon, rank jdh.Rank) {
	for i := jdh.Species; i >= rank; i-- {
		txSyncRankSet(c, tax, i)
	}
}

func txSyncRankSet(c *cmdapp.Command, tax *jdh.Taxon, rank jdh.Rank) {
	if len(tax.Id) == 0 {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncRankNav(c, l, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncRankNav(c, l, rank)
		return
	}
	if (tax.Rank < rank) && (tax.Rank != jdh.Unranked) {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncRankNav(c, l, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncRankNav(c, l, rank)
		return
	} else if tax.Rank == rank {
		return
	}
	eid := searchExtern(extDBFlag, tax.Extern)
	if len(eid) == 0 {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncRankNav(c, l, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncRankNav(c, l, rank)
		return
	}
	pExt := txSyncRankedParent(c, eid, rank)
	if pExt == nil {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncRankNav(c, l, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncRankNav(c, l, rank)
		return
	}
	// check if the potential parent is already in the database.
	p := taxon(c, localDB, extDBFlag+":"+pExt.Id)
	if len(p.Id) == 0 {
		p = txSyncAddToTaxonomy(c, pExt)
	}
	args := new(jdh.Values)
	args.Add(jdh.KeyId, tax.Id)
	args.Add(jdh.TaxParent, p.Id)
	localDB.Exec(jdh.Set, jdh.Taxonomy, args)
}

func txSyncRankNav(c *cmdapp.Command, l jdh.ListScanner, rank jdh.Rank) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txSyncRankSet(c, desc, rank)
	}
}

// search for a parent of the given rank.
func txSyncRankedParent(c *cmdapp.Command, id string, rank jdh.Rank) *jdh.Taxon {
	args := new(jdh.Values)
	args.Add(jdh.TaxParents, id)
	l, err := extDB.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer l.Close()
	for {
		p := &jdh.Taxon{}
		if err := l.Scan(p); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if p.Rank == rank {
			return p
		}
	}
	return nil
}

// txSyncPop implements populate option of tx.sync.
func txSyncPop(c *cmdapp.Command, tax *jdh.Taxon, prevRank, rank jdh.Rank) {
	r := tax.Rank
	if r == jdh.Unranked {
		r = prevRank
	}
	defer func() {
		l := getTaxDesc(c, localDB, tax.Id, true)
		txSyncPopNav(c, l, r, rank)
		l = getTaxDesc(c, localDB, tax.Id, false)
		txSyncPopNav(c, l, r, rank)
	}()
	if len(tax.Id) == 0 {
		return
	}
	if r < rank {
		return
	}
	eid := searchExtern(extDBFlag, tax.Extern)
	if len(eid) == 0 {
		return
	}
	// populate the taxon with new descendants
	l := getTaxDesc(c, extDB, eid, true)
	txSyncPopDesc(c, l, tax)
	l = getTaxDesc(c, extDB, eid, false)
	txSyncPopDesc(c, l, tax)
}

func txSyncPopDesc(c *cmdapp.Command, l jdh.ListScanner, p *jdh.Taxon) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		// check if the taxon is in the database
		if d := taxon(c, localDB, extDBFlag+":"+desc.Id); len(d.Id) != 0 {
			continue
		}
		// adds the new taxon
		tax := &jdh.Taxon{}
		*tax = *desc
		tax.Id = ""
		tax.Parent = p.Id
		tax.Extern = []string{extDBFlag + ":" + desc.Id}
		if _, err := localDB.Exec(jdh.Add, jdh.Taxonomy, tax); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
	}
}

func txSyncPopNav(c *cmdapp.Command, l jdh.ListScanner, prevRank, rank jdh.Rank) {
	for {
		desc := &jdh.Taxon{}
		if err := l.Scan(desc); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		txSyncPop(c, desc, prevRank, rank)
	}
}
