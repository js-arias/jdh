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

var spSet = &cmdapp.Command{
	Name:     "sp.set",
	Synopsis: `-i|--id value [-p|--port value] [<key=value>...]`,
	Short:    "sets an specimen value",
	Long: `
Description

Sp.set sets a particular value for an specimen in the database. Use this
command to edit the specimen database, instead of manual edition.

If no key is defined, the key values will be read from the standard input, 
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the specimen to be set. It is a required option.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "basis=preserved specimen"
      Valid keys are:
          basis          Basis of record.
          catalog        Catalog code of the specimen.
          collecion      Collection in which the specimen is vouchered.
          collector      Collector of the specimen.
          comment        A free text comment on the specimen.
          country        Country in which the specimen was collected, using
                         ISO 3166-1 alpha-2.
          county         County (or a similar administration entity) in
                         which the specimen was collected.
          date           date of specimen collection, in ISO 8601 format,
                         for example: 2006-01-02T15:04:05+07:00.
          determiner     Person who identify the specimen.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>, for example: gbif:866197949.
          locality       Locality in which the specimen was collected.
          lonlat         Longitude and latitude of the collection point.
          reference      A bibliographic reference to the specimen.
          source         Source of the georeference assignation.
          state          State or province in which the specimen was
                         collected.
          taxon          Id of the taxon assigned to the specimen.
          uncertainty    Uncertainty, in meters, of the georeference
                         assignation.
          validation     Source of the georeference validation.
	`,
}

func init() {
	spSet.Flag.StringVar(&idFlag, "id", "", "")
	spSet.Flag.StringVar(&idFlag, "i", "", "")
	spSet.Flag.StringVar(&portFlag, "port", "", "")
	spSet.Flag.StringVar(&portFlag, "p", "", "")
	spSet.Run = spSetRun
}

func spSetRun(c *cmdapp.Command, args []string) {
	if len(idFlag) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr("expectiong specimen id"))
		c.Usage()
	}
	openLocal(c)
	spe := specimen(c, localDB, idFlag)
	if len(spe.Id) == 0 {
		return
	}
	vals := valsFromArgs(spe.Id, args)
	localDB.Exec(jdh.Set, jdh.Specimens, vals)
	localDB.Exec(jdh.Commit, "", nil)
}
