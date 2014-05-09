// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

// connection flags
var (
	dirFlag  string // set db directory, -d|--dir
	portFlag string // set connection port, -p|--port
	commFlag bool   // commit flag, -c|--commit
)

// common flags
var (
	extDBFlag   string // set the extern database, -e|--extdb
	formatFlag  string // set a format value, -f|--format
	idFlag      string // set a db id, -i|--id
	keyFlag     string // key flag -k|--key
	machineFlag bool   // set machine output, -m|--machine
	matchFlag   bool   // set match option, -m|--match
	updateFlag  bool   // set update option, -u|--update
	verboseFlag bool   // set command verbosity, -v|--verbose
)

// flags used by taxon commands.
var (
	ancFlag     string // set a parent id, -a|--anc
	ancsFlag    bool   // ancs flag, -a|--ancs
	collFlag    bool   // set collapse option -c|--collapse
	popFlag     string // populate flag, -l|--populate
	rankFlag    string // set a rank, -r|--rank
	synonymFlag bool   // synonym flag, -s|--synonym
)
