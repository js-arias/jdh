// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
)

var jdhClose = &cmdapp.Command{
	Name:     "close",
	Synopsis: `[-c|--commit] [-p|--port value]`,
	Short:    "closes the server",
	IsCommon: true,
	Long: `
Description

Close sends a shutdown request to the server. It is up to the server to
honor this request.

Options

    -c
    --commit
      If set, the database will be saved into harddisk before closing.
      
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	jdhClose.Flag.BoolVar(&commFlag, "commit", false, "")
	jdhClose.Flag.BoolVar(&commFlag, "c", false, "")
	jdhClose.Flag.StringVar(&portFlag, "port", "", "")
	jdhClose.Flag.StringVar(&portFlag, "p", "", "")
	jdhClose.Run = closeRun
}

func closeRun(c *cmdapp.Command, args []string) {
	openLocal(c)
	if commFlag {
		localDB.Exec(jdh.Commit, "", nil)
	}
	localDB.Close()
}
