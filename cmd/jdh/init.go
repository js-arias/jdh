// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/server"
)

var jdhInit = &cmdapp.Command{
	Name:     "init",
	Synopsis: `[-d|--dir path] [-p|--port value]`,
	Short:    "initializes the jdh server",
	IsCommon: true,
	Long: `
Description

Init startups the jdh database server. As jdh applications require a jdh
database, this command is usually the first one called before any other jdh
command.

By default, the server will be open in the current directory and at the
port :16917, this values can be changed by -d, --dir and -p, --port options
respectively.

Options

    -d path
    --dir path
      Sets the directory in which the database files will be located. By
      default, the current directory is used as the directory.
    
    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	jdhInit.Flag.StringVar(&dirFlag, "dir", "", "")
	jdhInit.Flag.StringVar(&dirFlag, "d", "", "")
	jdhInit.Flag.StringVar(&portFlag, "port", "", "")
	jdhInit.Flag.StringVar(&portFlag, "p", "", "")
	jdhInit.Run = initRun
}

func initRun(c *cmdapp.Command, args []string) {
	if err := server.Listen(portFlag, dirFlag); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", c.ErrStr(err))
		os.Exit(1)
	}
}
