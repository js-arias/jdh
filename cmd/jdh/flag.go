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

// flags used by dataset commands.
var (
	citFlag bool // set citation option, -c|--citation
	licFlag bool // set license option, -l|--license
	urlFlag bool // set url option, -u|--url
)

// flags used by taxon commands.
var (
	ancFlag     string // set a parent id, -a|--anc
	ancsFlag    bool   // ancs flag, -a|--ancs
	colpFlag    bool   // set collapse option -c|--collapse
	popFlag     string // populate flag, -l|--populate
	rankFlag    string // set a rank, -r|--rank
	synonymFlag bool   // synonym flag, -s|--synonym
	validFlag   bool   // validate flag, -d|--validate
)

// flags used by specimen commands
var (
	addFlag     bool   // add georeferences, -a|--add
	childFlag   bool   // children flag, -c|--children
	corrFlag    bool   // correct a georeference, -c|--correct
	delFlag     bool   // delete georeference, -d|--delete
	dsetFlag    string // set dataset, -d|--dataset
	geoRefFlag  bool   // georef flag -g|--georef
	countryFlag string // set country, -r|--coutry
	noRefFlag   bool   // no georef flag -n|--nogeoref
	skipFlag    bool   // skip flag, -s|--skip
	taxonFlag   string // set taxon, -t|--taxon
	uncertFlag  int    // set uncertainty, -u|--uncert
)
