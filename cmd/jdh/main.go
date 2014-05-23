// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"runtime"

	"github.com/js-arias/cmdapp"
)

var app = &cmdapp.App{
	Name:     "jdh",
	Synopsis: "[help] <command> [<args>...]",
	Short:    "Joseph Dalton Hooker",
	Long: `
Description

Jdh (named after botanist and biogeographer Joseph Dalton Hooker
<http://en.wikipedia.org/wiki/Joseph_Dalton_Hooker>), is an open source 
software for management of taxonomic and biogeographic data.

Use 'jdh help --all' for a list of available commands. To see help or 
information about a command type 'jdh help <command>'.

Author

J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
INSUE, Facultad de Ciencias Naturales e Instituto Miguel Lillo,
Universidad Nacional de Tucumán, Miguel Lillo 205, S.M. de Tucumán (4000),
Tucumán, Argentina.

Reporting bugs

Please report any bug to J.S. Arias at <jsalarias@csnat.unt.edu.ar>.
	`,
	Commands: []*cmdapp.Command{
		jdhInit,
		jdhClose,
		dsIn,
		dsInfo,
		dsLs,
		dsSet,
		raDel,
		raInfo,
		raLs,
		raMk,
		raSet,
		spDel,
		spIn,
		spInfo,
		spGref,
		spLs,
		spPop,
		spSet,
		txDel,
		txForce,
		txIn,
		txInfo,
		txLs,
		txSet,
		txSync,
		txTaxo,
		gStart,
	},
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app.Run()
}
