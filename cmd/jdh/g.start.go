// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"github.com/js-arias/cmdapp"
)

var gStart = &cmdapp.Command{
	Name:     "start",
	Short:    "quick start guide",
	IsCommon: true,
	Long: `
Jdh (named after botanist and biogeographer Joseph Dalton Hooker 
<http://en.wikipedia.org/wiki/Joseph_Dalton_Hooker">), is an open source 
software for management, processing, and analysis of taxonomic and 
biogeographic data.

Startup of JDH

Jdh is a client-server program. Unless you are using an extern database, 
most of the jdh commands, require a local database server must be available 
to access the data.

To startup the server, navigate to the database directory. It can be any 
directory, if database files are present in the directory, they will be used 
by the server. Then, it is recommended to create a new directory for each 
dataset you want to use. Once you are in the database directory type 
'jdh init':

	$ jdh init

As the server must be open through the whole season of jdh, you can open it in 
an independent command line season, or in the background, redirecting the 
loging file. In Linux:

	$ jdh init > log.txt &

And in Windows:

	> start /b jdh init > log.txt

Once the server is open, by default any jdh command will be using the data in 
the database file as the data for the command.

Finish session

Once you finish a session, you must close the server. To do this type 
'jdh close':

	$ jdh close

The server will be closed.

In standard jdh commands, modifications in the database will be immediately 
committed (saved to hard drive). If you are using some third party 
application, this might be not the case, so you can call jdh close with the 
option '--commit' to make sure the database will be saved to hard drive.

	$ jdh close --commit

Other suggestions to maintain the database

The jdh native database is a based on text, so it can be maintained using git 
<http://git-scm.com/>. This powerful tool will allows to track the history of 
the database, and will help with datasharing in the case of multi-author 
projects.
	`,
}
