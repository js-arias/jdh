// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var txSet = &cmdapp.Command{
	Name: "tx.set",
	Synopsis: `[-i|--id value] [-p|--port value] [<name> [<parentname>]]
	[<key=value>...]`,
	Short: "sets a taxon value",
	Long: `
Description

Tx.set sets a particular value for a taxon in the database. Use of this
command to edit the taxon database, instead of manual edition.

If no taxon and key is defined, the key values will be read from the
standard input, it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the taxon to be set.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "authority=(Linnaeus, 1758)".
      Valid keys are:
      Valid keys are:
          authority      Authorship of the taxon.
          comment        A free text comment on the taxon.
          extern         Extern identifiers of the taxon, in the form
                         <service>:<key>, for example: "gbif:5216933". If the
                         key is empty then the service will be eliminated,
                         eg. "gbif:".
          name           Name of the taxon.
          parent         Id of the new parent.
          rank           The taxon rank, valid values are:
                             unranked
                             kingdom
                             class
                             order
                             family
                             genus
                             species

          synonym        Set the taxon as synonym. If no id of a new parent
                         is defined, the taxon will be synonymized with its
                         current parent.
          valid          Set the taxon as valid, ignores the value. The
                         taxon will be set as sister of its previous senior.

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
	txSet.Flag.StringVar(&idFlag, "id", "", "")
	txSet.Flag.StringVar(&idFlag, "i", "", "")
	txSet.Flag.StringVar(&portFlag, "port", "", "")
	txSet.Flag.StringVar(&portFlag, "p", "", "")
	txSet.Run = txSetRun
}

func txSetRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	var tax *jdh.Taxon
	if len(idFlag) > 0 {
		tax = taxon(c, localDB, idFlag)
		if len(tax.Id) == 0 {
			return
		}
	} else if (len(args) > 0) && (strings.Index(args[0], "=") < 0) {
		name := args[0]
		pName := ""
		args = args[1:]
		if (len(args) > 0) && (strings.Index(args[0], "=") < 0) {
			pName = args[0]
			args = args[1:]

		}
		tax = pickTaxName(c, localDB, name, pName)
		if len(tax.Id) == 0 {
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong taxon name or id"))
		c.Usage()
	}
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, tax.Id)
	if len(args) == 0 {
		in := bufio.NewReader(os.Stdin)
		for {
			tn, err := readLine(in)
			if err != nil {
				break
			}
			if len(tn) == 1 {
				continue
			}
			ln := strings.Join(tn, " ")
			if strings.Index(ln, "=") < 0 {
				continue
			}
			vals.Add(parseKeyValArg(ln))
		}
	} else {
		for _, a := range args {
			if strings.Index(a, "=") < 0 {
				continue
			}
			vals.Add(parseKeyValArg(a))
		}
	}
	localDB.Exec(jdh.Set, jdh.Taxonomy, vals)
	localDB.Exec(jdh.Commit, "", nil)
}
