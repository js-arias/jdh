JDH - Joseph Dalton Hooker
==========================

Jdh (named after botanist and biogeographer [Joseph Dalton Hooker](
http://en.wikipedia.org/wiki/Joseph_Dalton_Hooker), is an open source 
software for management of taxonomic and biogeographic data.

It is implemented in Go.

Motivation
----------

The main objective of JDH is to provide a simple and flexible platform to 
manipulate taxonomic and biogeographic data, in a way that will be possible to 
interact with biodiversity databases (e.g. [GBIF](http://www.gbif.org/)), and 
allowing an easy use of that data in biogeographic studies.

Code organization
-----------------

The code is separated in two parts, a libary (pkg directory) with several Go 
packages aimed to store and manipulate biogeographic data. The second part is 
the source for jdh binary executables (cmd directory), whereas the code is 
fully functional, the main objective is to be an example of the usage of the 
jdh packages.

To see main code documentation use godoc.

Architecture
------------

JDH is a server-client model. A server is used to store the database that is 
typically accessed by a net connection. The program include a default server 
(implemented in pkg/native and pkg/server, default port is :16917), that is 
assumed by each command, but it is possible to use other servers, this is 
implemented through database drivers (implemented in pkg/driver), so it is 
possible to use the jdh commands to consult remote databases (such as GBIF) in 
the same way as the local database.

A jdh dataset is a based on text, so it can be easily maintained using 
[git](http://git-scm.com/).

Quick startup guide
-------------------

The command structure of jdh is as a set of hosted commands (as in 
[go command](http://golang.org/cmd/go/) and 
[git command](http://git-scm.com/docs). Usually the command starts with an 
acronym that indicate the kind of expected data.

To startup the server, navigate to the database directory. It can be any 
directory, if database files are present in the directory, they will be used 
by the server. Then, it is recommended to create a new directory for each 
dataset you want to use. Once you are in the database directory type 'jdh init':

    $ jdh init

As the server must be open through the whole season of jdh, you can open it in 
an independent command line season, or in the background, redirecting the 
loging file. In Linux:

    $ jdh init > log.txt &

And in Windows:

    > start /b jdh init > log.txt

Once the server is open, by default any jdh command will be using the data in 
the database file as the data for the command.

Once you finish a session, you must close the server. To do this type 
'jdh close':

    $ jdh close

And the server will be closed.

In standard jdh commands, modifications in the database will be immediately 
committed (saved to hard drive). If you are using some customised 
application, this might be not the case, so you can call jdh close with the 
option '--commit' to make sure the database will be saved to hard drive.

    $ jdh close --commit


Funding
-------

This project is made thanks to the financial support from a doctoral grant 
awarded by CONICET (to J. Salvador Arias), and the GBIF Student Award (2012, 
to J. Salvador Arias). Material facilities are provided by INSUE at the 
Facultad de Ciencias Naturales e Instituto Miguel Lillo, Universidad Nacional 
de Tucumán, Tucumán (Argentina).

Authorship and license
----------------------

Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
All rights reserved.
Distributed under BSD2 license that can be found in the LICENSE file.

