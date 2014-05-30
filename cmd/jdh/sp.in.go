// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/geography"
	"github.com/js-arias/jdh/pkg/jdh"
)

var spIn = &cmdapp.Command{
	Name: "sp.in",
	Synopsis: `[-a|--anc value] [-d|--dataset value] [-f|--format value]
	[-p|--port value] [-r|--rank name] [-s|--skip] [-v|--verbose]
	[<file>...]`,
	Short: "imports specimen data",
	Long: `
Description

Sp.in reads specimen data from the indicated files, or standard input
(if no file is defined), and adds them to the jdh database.

Default input format is txt. If the format is txt, it is assummed that each 
line corresponds to a specimen (lines starting with '#' or ';' will be ignored).

If the taxons in the input file are not in the database, they will be
added to the root, valid, and unranked. Use options -a, --anc, -r, --rank
to change this behavior.

Options

    -a value
    --anc value
      Sets the parent of the added taxons. The value must be a valid id.
    
    -d value
    --dataset value
      If defined, new specimens will be set to the indicated dataset. If 
      the value is not a valid id, then this option will be ignored.

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format
          ndm        Ndm format

    -k
    --skip
      If set, taxons in the database will be ignored. Useful to add a
      dataset after correcting warnings.

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

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -v
    --verbose
      If set, the name and id of each added taxon will be print in the
      standard output.
    
    <file>
      One or more files to be proccessed by sp.in. If no file is given
      then the information is expected to be from the standard input.
	`,
}

func init() {
	spIn.Flag.StringVar(&ancFlag, "anc", "", "")
	spIn.Flag.StringVar(&ancFlag, "a", "", "")
	spIn.Flag.StringVar(&dsetFlag, "dataset", "", "")
	spIn.Flag.StringVar(&dsetFlag, "d", "", "")
	spIn.Flag.StringVar(&formatFlag, "format", "", "")
	spIn.Flag.StringVar(&formatFlag, "f", "", "")
	spIn.Flag.StringVar(&portFlag, "port", "", "")
	spIn.Flag.StringVar(&portFlag, "p", "", "")
	spIn.Flag.StringVar(&rankFlag, "rank", "", "")
	spIn.Flag.StringVar(&rankFlag, "r", "", "")
	spIn.Flag.BoolVar(&skipFlag, "skip", false, "")
	spIn.Flag.BoolVar(&skipFlag, "s", false, "")
	spIn.Flag.BoolVar(&verboseFlag, "verbose", false, "")
	spIn.Flag.BoolVar(&verboseFlag, "v", false, "")
	spIn.Run = spInRun
}

func spInRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	pId := ""
	if len(ancFlag) > 0 {
		p := taxon(c, localDB, ancFlag)
		if !p.IsValid {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("taxon "+p.Name+" a synonym, can ot be a parent"))
			os.Exit(1)
		}
		pId = p.Id
	}
	var set *jdh.Dataset
	if len(dsetFlag) > 0 {
		set = dataset(c, localDB, dsetFlag)
	} else {
		set = &jdh.Dataset{}
	}
	rank := jdh.Unranked
	if len(rankFlag) > 0 {
		rank = jdh.GetRank(rankFlag)
	}
	format := "txt"
	if len(formatFlag) > 0 {
		format = formatFlag
	}
	if len(args) > 0 {
		switch format {
		case "txt":
			for _, fname := range args {
				spInTxt(c, fname, pId, rank, set)
			}
		case "ndm":
			for _, fname := range args {
				spInNdm(c, fname, pId, rank, set)
			}
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	} else {
		switch format {
		case "txt":
			spInTxt(c, "", pId, rank, set)
		case "ndm":
			spInNdm(c, "", pId, rank, set)
		default:
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("format "+format+" unknown"))
			os.Exit(1)
		}
	}
	localDB.Exec(jdh.Commit, "", nil)
}

func spInTxt(c *cmdapp.Command, fname, parent string, rank jdh.Rank, set *jdh.Dataset) {
	var in *bufio.Reader
	if len(fname) > 0 {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			return
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}
	for {
		ln, err := readLine(in)
		if err != nil {
			break
		}
		if len(ln) < 3 {
			continue
		}
		id := ln[0]
		lon, err := strconv.ParseFloat(ln[1], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			continue
		}
		lat, err := strconv.ParseFloat(ln[2], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			continue
		}
		cat := ""
		if len(ln) > 3 {
			cat = ln[3]
			if sp := specimen(c, localDB, cat); len(sp.Id) > 0 {
				continue
			}
		}
		if ((!geography.IsLon(lon)) || (!geography.IsLat(lat))) && (len(cat) == 0) {
			continue
		}
		sId := set.Id
		if len(ln) > 4 {
			sId = ln[4]
			if ds := dataset(c, localDB, sId); len(ds.Id) > 0 {
				sId = ds.Id
			} else {
				sId = ""
			}
		}
		spe := &jdh.Specimen{
			Taxon:   id,
			Catalog: cat,
			Dataset: sId,
		}
		if geography.IsLon(lon) && geography.IsLat(lat) {
			spe.Georef.Point = geography.Point{Lon: lon, Lat: lat}
		} else {
			if len(cat) == 0 {
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(fmt.Sprintf("invalid coordinates %.5f %.5f", lon, lat)))
				continue
			}
			spe.Georef = geography.InvalidGeoref()
		}
		_, err = localDB.Exec(jdh.Add, jdh.Specimens, spe)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			continue
		}
	}
}

func spInNdm(c *cmdapp.Command, fname, parent string, rank jdh.Rank, set *jdh.Dataset) {
	var in *bufio.Reader
	if len(fname) > 0 {
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			return
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}
	var tok string
	var err error
	//read header
	var transpose, nocommas bool
	xneg, yneg := float64(1), float64(1)
	for {
		tok, err = readString(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if tok == "xydata" {
			break
		}
		switch tok {
		case "ynegative":
			yneg = -1
		case "xnegative":
			xneg = -1
		case "transpose", "longlat":
			transpose = true
		case "nocommas":
			nocommas = true
		case "latlong":
			transpose = false
		}
	}
	tok, err = readString(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	for {
		if err == io.EOF {
			break
		}
		if (tok == ";") || (tok == "groups") || (tok == "map") {
			break
		}
		if tok != "sp" {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expecting 'sp' found: "+tok))
			os.Exit(1)
		}
		txNum := ""
		txNum, err = readString(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		var r rune
		r, err = peekNext(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		var txname string
		if r == '[' {
			// feed the first rune
			in.ReadRune()
			nm, err := readBlock(in, ']')
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
			fld := strings.Fields(nm)
			if len(fld) > 1 {
				// skips the optional filling
				lst := []rune(fld[len(fld)-1])
				if unicode.IsNumber(lst[0]) {
					fld = fld[:len(fld)-1]
				}
			}
			txname = strings.Join(fld, " ")
		}
		// skips specimens for anonymous taxons
		if len(txname) == 0 {
			if verboseFlag {
				fmt.Fprintf(os.Stdout, "WARNING:\t%s\t\tUnnamed taxon\n", txNum)
			}
			tok, err = skipNdmTaxon(c, in)
			continue
		}
		mult, id := spInSearchNmdTaxon(c, txname, parent, txNum, rank)
		if mult {
			tok, err = skipNdmTaxon(c, in)
			continue
		}
		if len(id) == 0 {
			pId := parent
			// we know that the first name of a species is the genus.
			if rank == jdh.Species {
				tn := strings.Fields(txname)
				if len(tn) > 1 {
					if p := taxInDB(c, localDB, tn[0], parent, jdh.Genus, true); p != nil {
						pId = p.Id
					} else {
						par := &jdh.Taxon{
							Name:    tn[0],
							IsValid: true,
							Parent:  parent,
							Rank:    jdh.Genus,
						}
						pId, err = localDB.Exec(jdh.Add, jdh.Taxonomy, par)
						if err != nil {
							fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
							tok, err = skipNdmTaxon(c, in)
							continue
						}
						if verboseFlag {
							fmt.Fprintf(os.Stdout, "%s %s\n", pId, par.Name)
						}
					}
				}
			}
			tax := &jdh.Taxon{
				Name:    txname,
				IsValid: true,
				Parent:  pId,
				Rank:    rank,
			}
			id, err = localDB.Exec(jdh.Add, jdh.Taxonomy, tax)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				tok, err = skipNdmTaxon(c, in)
				continue
			}
			if verboseFlag {
				fmt.Fprintf(os.Stdout, "%s %s\n", id, tax.Name)
			}
		} else if skipFlag {
			tok, err = skipNdmTaxon(c, in)
			continue
		}
		// read the specimens
		speOk := 0
		for {
			tok, err = readString(in)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
				os.Exit(1)
			}
			if (tok == "sp") || (tok == ";") || (tok == "groups") || (tok == "map") {
				break
			}
			var v1, v2 float64
			if nocommas {
				val1 := tok
				tok, err = readString(in)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					os.Exit(1)
				}
				val2 := tok
				v1, err = strconv.ParseFloat(val1, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
				v2, err = strconv.ParseFloat(val2, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
			} else {
				v := strings.Split(tok, ",")
				if len(v) < 2 {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expecting more than one comma separtated values"))
					continue
				}
				v1, err = strconv.ParseFloat(v[0], 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
				v2, err = strconv.ParseFloat(v[1], 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
			}
			lon, lat := float64(360), float64(360)
			if transpose {
				lon = v1 * xneg
				lat = v2 * yneg
			} else {
				lon = v2 * xneg
				lat = v1 * yneg
			}
			if geography.IsLon(lon) && geography.IsLat(lat) {
				spe := &jdh.Specimen{
					Taxon:   id,
					Dataset: set.Id,
				}
				spe.Georef.Point = geography.Point{Lon: lon, Lat: lat}
				_, err = localDB.Exec(jdh.Add, jdh.Specimens, spe)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
					continue
				}
				speOk++
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(fmt.Sprintf("invalid coordinates %.5f %.5f", v1, v2)))
				continue
			}
		}
		if verboseFlag {
			fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%d\tspecs added\n", txNum, txname, id, speOk)
		}
	}
}

func skipNdmTaxon(c *cmdapp.Command, in *bufio.Reader) (string, error) {
	for {
		tok, err := readString(in)
		if err != nil {
			if err == io.EOF {
				return "", err
			}
			fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
			os.Exit(1)
		}
		if (tok == "sp") || (tok == ";") || (tok == "groups") || (tok == "map") {
			return tok, nil
		}
	}
}

func spInSearchNmdTaxon(c *cmdapp.Command, name, parent, txNum string, rank jdh.Rank) (bool, string) {
	args := new(jdh.Values)
	args.Add(jdh.TaxName, name)
	if len(parent) != 0 {
		args.Add(jdh.TaxParent, parent)
	}
	if rank != jdh.Unranked {
		args.Add(jdh.TaxRank, rank.String())
	}
	l, err := localDB.List(jdh.Taxonomy, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
	defer l.Close()
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
			if verboseFlag {
				fmt.Fprintf(os.Stdout, "WARNING:\t%s\t%s\tAmbiguos name\n", txNum)
				fmt.Fprintf(os.Stdout, "%s\t%s\n", tax.Id, tax.Name)
			}
			mult = true

		}
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "%s\t%s\n", ot.Id, ot.Name)
		}
	}
	if mult || (tax == nil) {
		return mult, ""
	}
	return false, tax.Id
}
