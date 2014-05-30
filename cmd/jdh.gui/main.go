// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"runtime"

	"github.com/js-arias/cmdapp"
)

var app = &cmdapp.App{
	Name:     "jdh.gui",
	Synopsis: "[help] <command> [<args>...]",
	Short:    "Joseph Dalton Hooker GUI",
	Long: `
Description

Jdh.gui is a gui application based on jdh. To run it requires a running jdh
server, or an extern database.

Use 'jdh.gui help --all' for a list of all available commands. To see help or
information about a commend type 'jdh.gui help <command>'.

Author

J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
INSUE, Facultad de Ciencias Naturales e Instituto Miguel Lillo,
Universidad Nacional de Tucumán, Miguel Lillo 205, S.M. de Tucumán (4000),
Tucumán, Argentina.

Reporting bugs

Please report any bug to J.S. Arias at <jsalarias@csnat.unt.edu.ar>.
	`,
	Commands: []*cmdapp.Command{
		trView,
	},
}

var cmd *cmdapp.Command

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app.Run()
}
