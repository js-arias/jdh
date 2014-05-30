// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var trIn = &cmdapp.Command{
	Name: "tr.in",
	Synopsis: `[-a|--anc value] [-f|--format value] [-p|--port value]
	[-r|--rank value] [-v|--verbose] [<file>...]`,
	Short: "imports tree data",
	Long: `
Description

Tr.in reads tree data from the indicated files, or standard input (if no file
is defined), and adds them to the jdh database.

Default input format are tnt tree files (i.e. tread command), assuming that
the trees are saved with the names and underlines will be replaced with spaces.

Options

    -a value
    --anc value
      Sets the parent of the terminals of the tree. The value must be a valid
      id.

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          tnt        Tnt tree format (tread)
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      Set the rank of the added taxon. If the taxon has a parent (the -a,
      --anc options) the parent must be concordant with the given rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

    -v
    --verbose
      If set, the id of each added tree will be print in the standard output.
    
    <file>
      One or more files to be proccessed by tr.in. If no file is given
      then the information is expected to be from the standard input.
	`,
}

func init() {
	trIn.Flag.StringVar(&ancFlag, "anc", "", "")
	trIn.Flag.StringVar(&ancFlag, "a", "", "")
	trIn.Flag.StringVar(&formatFlag, "format", "", "")
	trIn.Flag.StringVar(&formatFlag, "f", "", "")
	trIn.Flag.StringVar(&portFlag, "port", "", "")
	trIn.Flag.StringVar(&portFlag, "p", "", "")
	trIn.Flag.StringVar(&rankFlag, "rank", "", "")
	trIn.Flag.StringVar(&rankFlag, "r", "", "")
	trIn.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	trIn.Flag.BoolVar(&verboseFlag, "v", false, "")
	trIn.Run = trInRun
}

func trInRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	format := "tnt"
	if len(formatFlag) > 0 {
		format = formatFlag
	}
	pId := ""
	if len(ancFlag) > 0 {
		p := taxon(c, localDB, ancFlag)
		if !p.IsValid {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("taxon "+p.Name+" a synonym, can ot be a parent"))
			os.Exit(1)
		}
		pId = p.Id
	}
	if len(args) > 0 {
		switch format {
		case "tnt":
			for _, fname := range args {
				trInTnt(c, fname, pId)
			}
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	} else {
		switch format {
		case "tnt":
			trInTnt(c, "", pId)
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func trInTnt(c *cmdapp.Command, fname, pId string) {
	var in *bufio.Reader
	if len(fname) > 0 {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}
	if err := trInSkipTntHead(in); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	taxLs := make(map[string]string)
	for {
		id, err := localDB.Exec(jdh.Add, jdh.Trees, &jdh.Phylogeny{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if _, err = trInReadTreeNode(in, id, "", pId, taxLs); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if !trInTntNext(in) {
			break
		}
	}
}

// TrInReadTreeNode reads a tree node in parenthetical notation.
func trInReadTreeNode(in *bufio.Reader, tree, anc, pId string, taxLs map[string]string) (string, error) {
	nod := &jdh.Node{
		Tree:   tree,
		Parent: anc,
	}
	id, err := localDB.Exec(jdh.Add, jdh.Nodes, nod)
	if err != nil {
		return "", err
	}
	nod.Id = id
	last := ""
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return "", err
		}
		if unicode.IsSpace(r) {
			continue
		}
		if r == '(' {
			last, err = trInReadTreeNode(in, tree, nod.Id, pId, taxLs)
			if err != nil {
				return "", err
			}
			continue
		}
		if r == ')' {
			break
		}
		if r == ',' {
			last = ""
			continue
		}
		if r == ':' {
			if len(last) == 0 {
				return "", errors.New("unexpected ':' token")
			}
			v, err := trInNumber(in)
			if err != nil {
				return "", err
			}
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, last)
			vals.Add(jdh.NodLength, strconv.FormatUint(uint64(v), 10))
			localDB.Exec(jdh.Set, jdh.Nodes, vals)
			last = ""
			continue
		}
		// a terminal
		in.UnreadRune()
		nm, err := trInString(in)
		if err != nil {
			return "", err
		}
		tax, ok := taxLs[nm]
		if !ok {
			tax, err = trInPickTax(nm, pId)
			if err != nil {
				return "", err
			}
			taxLs[nm] = tax
		}
		term := &jdh.Node{
			Tree:   tree,
			Parent: nod.Id,
			Taxon:  tax,
		}
		last, err = localDB.Exec(jdh.Add, jdh.Nodes, term)
		if err != nil {
			return "", err
		}
	}
	return nod.Id, nil
}

// TrInPickTax returns the id of a given taxon name. If there are more taxons
// fullfilling the name, then it will print a list of the potential
// names and finish the program.
func trInPickTax(name, parent string) (string, error) {
	vals := new(jdh.Values)
	vals.Add(jdh.TaxName, name)
	if len(parent) != 0 {
		vals.Add(jdh.TaxParent, parent)
	}
	rank := jdh.Unranked
	if len(rankFlag) > 0 {
		rank = jdh.GetRank(rankFlag)
		if rank != jdh.Unranked {
			vals.Add(jdh.TaxRank, rankFlag)
		}
	}
	l, err := localDB.List(jdh.Taxonomy, vals)
	if err != nil {
		return "", err
	}
	var tax *jdh.Taxon
	mult := ""
	for {
		ot := &jdh.Taxon{}
		if err := l.Scan(ot); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if tax == nil {
			tax = ot
			continue
		}
		if len(mult) == 0 {
			mult = fmt.Sprintf("ambiguos taxon name\n")
			mult += fmt.Sprintf("%s\t%s\n", tax.Id, tax.Name)
		}
		mult += fmt.Sprintf("%s\t%s\n", ot.Id, ot.Name)
	}
	if len(mult) > 0 {
		return "", errors.New(mult)
	}
	if tax == nil {
		tax = &jdh.Taxon{
			Name:    name,
			IsValid: true,
			Parent:  parent,
			Rank:    rank,
		}
		return localDB.Exec(jdh.Add, jdh.Taxonomy, tax)
	}
	return tax.Id, nil
}

// TrInNumber reads a number from a tree number.
func trInNumber(in *bufio.Reader) (uint, error) {
	var s []rune
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return 0, err
		}
		if unicode.IsSpace(r) || (r == ',') {
			break
		}
		if (r == '(') || (r == ')') {
			in.UnreadRune()
			break
		}
		s = append(s, r)
	}
	v, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return 0, err
	}
	return uint(v * 1000000), nil
}

// TrInString reads a string from a tree string.
func trInString(in *bufio.Reader) (string, error) {
	var nm []rune
	r, _, _ := in.ReadRune()
	if r == '\'' {
		return readBlock(in, '\'')
	}
	in.UnreadRune()
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return "", err
		}
		if unicode.IsSpace(r) || (r == ',') {
			break
		}
		if (r == '(') || (r == ')') || (r == ':') {
			in.UnreadRune()
			break
		}
		nm = append(nm, r)
	}
	return strings.Join(strings.Split(string(nm), "_"), " "), nil
}

// TrInSkipTntHead skips the tnt tread header.
func trInSkipTntHead(in *bufio.Reader) error {
	if err := skipSpaces(in); err != nil {
		return err
	}
	r, _, _ := in.ReadRune()
	if r != 't' {
		return errors.New("expecting 'tread' command")
	}
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return err
		}
		if r == '\'' {
			if _, err := readBlock(in, '\''); err != nil {
				return err
			}
			continue
		}
		if r == '(' {
			return nil
		}
	}
}

// TrInTntNext returns true if there are more trees in the file.
func trInTntNext(in *bufio.Reader) bool {
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return false
		}
		if r == ';' {
			return false
		}
		if r == '*' {
			break
		}
	}
	// there are more trees, feed characters until found and open
	// parenthesis.
	if _, err := readBlock(in, '('); err != nil {
		return false
	}
	return true
}
