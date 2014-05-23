// Authomatically generated doc.go file for use with godoc.

/*
Joseph Dalton Hooker.

Synopsis

    jdh [help] <command> [<args>...]

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

Initializes the jdh server

Synopsis

    jdh init [-d|--dir path] [-p|--port value]

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

Closes the server

Synopsis

    jdh close [-c|--commit] [-p|--port value]

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

Imports dataset data

Synopsis

    jdh ds.in [-f|--format value] [-p|--port value] [-v|--verbose]
	[<file>...]

Description

Ds.in reads dataset data form the indicated files, or standard input (if
no file is defined), and adds them to the jdh database.

Default input format is txt. If the format is txt, it is assummed that each
line corresponds to a dataset (lines starting with '#' or ';' will be ignored).

Options

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -v
    --verbose
      If set, the name and id of each added taxon will be print in the
      standard output.

    <file>
      One or more files to be proccessed by tx.in. If no file is given
      then the information is expected to be from the standard input.

Prints dataset information

Synopsis

    jdh ds.info -i|--id value [-e|--extdb name] [-k|--key value]
	[-m|--machine] [-p|--port value]

Description

Ds.info prints general information of a dataset in the database.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    datasets from gbif.

    -i value
    --id value
      Search for the indicated dataset id. It is a required option.

    -k value
    --key value
      If set, only a particular value of the taxon will be printed.
      Valid keys are:
          citation       Preferred citation of the dataset.
          comment        A free text comment on the dataset.
          license        License of use of the data.
          extern         Extern identifiers of the dataset, in the form
                         <service>:<key>.
          title          Title of the dataset.
          url            Url of the dataset.

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

Prints a list of datasets

Synopsis

    jdh ds.ls [-c|--citation] [-e|--extdb name] [-l|--license]
	[-m|--machine] [-p|--port value] [-u|--url] [-v|--verbose]

Description

Ds.ls prints a list of datasets. With no option, ds.ls prints all the datasets
in the database.

Options

    -c
    --citation
      If set, citation information will be printed.

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    datasets from gbif.

    -l
    --license
      If set, license information will be printed.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -u
    --url
      If set the url of the dataset will be printed.

    -v
    --verbose
      If defined, then a large list will be printed. This option is ignored
      if -m or --machine option is defined.

Sets a dataset value

Synopsis

    jdh ds.set -i|--id value [-p|--port value] [<key=value>...]

Description

Ds.set sets a particular value for a dataset in the database. Use this
command to edit the dataset database, instead of manual edition.

If no key is defined, the key values will be read from the standard input,
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the dataset to be set. It is a required option.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <key=value>
      Indicates the key, following by an equal and the new value, if the
      new value is empty, it is interpreted as a deletion of the
      current value. For flexibility it is recommended to use quotations,
      e.g. "title=Global Biodiversity Information Facility".
      Valid keys are:
          citation       Preferred citation of the dataset.
          comment        A free text comment on the dataset.
          license        License of use of the data.
          extern         Extern identifiers of the dataset, in the form
                         <service>:<key>.
          title          Title of the dataset.
          url            Url of the dataset.

Deletes rasterized distributions

Synopsis

    jdh ra.del [-i|--id value] [-p|--port value] [-t|--taxon value]
	[<name> [<parentname>]]

Description

Ra.del removes a rasterized distribution from the database, or, if option -t
or --taxon is defined, or taxon name is given, all the rasterized
distributions associated with the indicated taxon. This option deletes the
rasters, neither specimens or taxons, use command sp.del or tx.del to perform
that operations.

Operations

    -i value
    --id value
      Search for the indicated raster id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      Search for the indicated taxon id.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Prints information about a rasterized distribution

Synopsis

    jdh ra.info -i|--id value [-k|--key value] [-m|--machine]
	[-p|--port value]

Description

Ra.info prints general information of a rasterized distribution in the
database.

Options

    -i value
    --id value
      Search for the indicated rasterized distribution id.
      This option is required.

    -k value
    --key value
      If set, only a particular value of the specimen will be printed.
      Valid keys are:
          column         Number of columns in the raster.
          comment        A free text comment on the raster.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>.
          pixel          Sets a pixel value in the raster, in the form
                         "X,Y,Val", in which Val is an int.
          reference      A bibliographic reference to the raster.
          source         Source of the raster.
          taxon          Id of the taxon assigned to the raster.

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

Prints a list of rasterized distributions

Synopsis

    jdh ra.ls [-c|--children] [-m|--machine] [-p|--port value]
	[-t|--taxon value] [-v|--verbose] [<name> [<parentname>]]

Description

Ra.ls prints a list of rasterized distributions associated with a taxon.

Options

    -c
    --children
      If set, the rastes associated with the indicated taxon, as well
      as the ones from its descendants, will be printed.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      Search for the indicated taxon id.

    -v
    --verbose
      If defined, then a large list (including ids) will be printed. This
      option is ignored if -m or --machine option is defined.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Creates raster distributions from specimen data

Synopsis

    jdh ra.mk [-e|--extdb name] [-p|--port value] [-r|--rank name]
	[-s|--size value] [-t|--taxon value] [<name> [<parentname>]]

Description

Ra.mk uses specimen data in the database (or an extern database, defined with
-e, --extdb option) to creates a precense-absence rasterized distributions.
Only valid taxons will be rasterized (although their sinonyms will be used to
create the raster).

When the options -t, --taxon or a name are used, the effect of the command
will only affect the indicated taxon and its descendants.

When the -r, --rank option is used, only the taxons at or below the indicated
rank will be rasterized.

Options

    -e name
    --extdb name
      Sets the a extern database to extract distribution data.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set, only taxons below the indicated rank will be populated.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -s value
    --size value
      Sets the size of a pixel side (pixels, or cells are assumed as squared),
      in terms of arc degrees. By default, the value is 1. It must be a value
      between 0 and 360 (not inclusive). The value will be arranged to make
      the number of colums fit well in the 360 degrees.

    -t value
    --taxon value
      Search for the indicated taxon id.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Sets a value in a rasterized distribution

Synopsis

    jdh ra.set -i|--id value [-p|--port value] [<key=value>...]

Description

Ra.set sets a particular value for a rasterized d in the database. Use this
command to edit the rasterized distribution database, instead of manual
edition.

If no key is defined, the key values will be read from the standard input,
it is assumed that each line is in the form:
     'key=value'
Lines starting with '#' or ';' will be ignored.

Options

    -i value
    --id value
      Indicate the raster to be set. It is a required option.

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
          comment        A free text comment on the raster.
          extern         Extern identifiers of the specimen, in the form
                         <service>:<key>.
          pixel          Sets a pixel value in the raster, in the form
                         "X,Y,Val", in which Val is an int.
          reference      A bibliographic reference to the raster.
          source         Source of the raster.
          taxon          Id of the taxon assigned to the raster.

Deletes specimens

Synopsis

    jdh sp.del [-i|--id value] [-p|--port value] [-t|--taxon value]
	[<name> [<parentname>]]

Description

Sp.del removes an specimen from the database, or, if option -t or --taxon is
defined, or a taxon name is given, all the specimens associated with the
indicated taxon. This option just deletes the specimens, not the taxon, to
delete a taxon use the command tx.del.

Options

    -i value
    --id value
      Search for the indicated specimen id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      Search for the indicated taxon id.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Imports specimen data

Synopsis

    jdh sp.in [-a|--anc value] [-d|--dataset value] [-f|--format value]
	[-p|--port value] [-r|--rank name] [-s|--skip] [-v|--verbose]
	[<file>...]

Description

Sp.in reads specimen data from the indicated files, or standard input
(if no file is defined), and adds them to the jdh database.

Default input format is txt. If the format is txt, it is assummed that each
line corresponds to a specimen (lines starting with '#' or ';' will be ignored).

If the taxons in the input file are not in the database, they will be
added to the root, valid, and unranked. Use options -a, --anc, -r, --rank
to change this behavior.

Options

    -a value
    --anc value
      Sets the parent of the added taxons. The value must be a valid id.

    -d value
    --dataset value
      If defined, new specimens will be set to the indicated dataset. If
      the value is not a valid id, then this option will be ignored.

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format
          ndm        Ndm format

    -k
    --skip
      If set, taxons in the database will be ignored. Useful to add a
      dataset after correcting warnings.

    -r name
    --rank name
      Set the rank of the added taxon. If the taxon has a parent (the -a,
      --anc options) the parent must be concordant with the given rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -v
    --verbose
      If set, the name and id of each added taxon will be print in the
      standard output.

    <file>
      One or more files to be proccessed by tx.in. If no file is given
      then the information is expected to be from the standard input.

Prints general specimen information

Synopsis

    jdh sp.info -i|--id value [-e|--extdb name] [-k|--key value]
	[-m|--machine] [-p|--port value]

Description

Sp.info prints general information of an specimen in the database.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    specimens from gbif.

    -i value
    --id value
      Search for the indicated specimen id.
      This option is required.

    -k value
    --key value
      If set, only a particular value of the specimen will be printed.
      Valid keys are:
          basis          Basis of record.
          catalog        Catalog code of the specimen.
          collector      Collector of the specimen.
          comment        A free text comment on the specimen.
          country        Country in which the specimen was collected, using
                         ISO 3166-1 alpha-2.
          county         County (or a similar administration entity) in
                         which the specimen was collected.
          dataset        Dataset that contains the specimen information.
          date           Date of specimen collection, in ISO 8601 format,
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

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

Validate and add specimen georeferences

Synopsis

    jdh sp.gref [-a|--add] [-c|--correct] [-p|--port value]
	[-t|--taxon value] [-u|--uncert value] [-v|--verbose]
	[<name> [<parentname>]]

Description

Sp.gref uses a gazatteer service (geolocate web service
<http://www.museum.tulane.edu/geolocate/>) to validate or add a georeference
of the specimens in the database.

Specimens without a set country and locality will indicated as not validated,
but not corrected, or deleted.

By default, it just print the specimens that fail the validation. With -c,
--correct option, it will try to correct the georeference, if possible (check
for flips in latitude and longitude, for example).

With -a, --add option, non georeferences specimens will be searched, and if a
valid location is found (under a given bound, defined by -u, --uncert option),
the value will be used to set the point. If this option is not set, the point
will be indicated as not validates, but not corrected, or deleted.

Options

    -a
    --add
      Check non-georeferenced records, and, if they are found, they will be
      added to the database.

    -c
    --correct
      If set, it will try to correct invalid georeferences. It will try it
      by flipping lon, lat values, and changing lon, lat values sings.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -t value
    --taxon value
      If set, only specimens in selected taxon id will be searched.

    -u value
    --uncert value
      Set valid uncertainty (in meters), values below the given uncertainty
      will be scored as validated, or added, if -a, --add option is defined.
      Default value is 110000, which is roughly, about 1º at the equator. With
      0, the uncertainty values defined in each specimen will be used. Maximum
      value is 200000 (200 km).

    -v
    --verbose
      If set, details on the error (if available), will be printed.


    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Prints a list of specimens

Synopsis

    jdh sp.ls [-c|--children] [-e|--extdb name] [-g|--georef]
	[-m|--machine] [-n|--nonref] [-p|--port value] [-r|--country name]
	[-t|--taxon value] [-v|--verbose] [<name> [<parentname>]]

Description

Sp.ls prints a list of specimens associated with a taxon.

By default, all records will be printed, use -g, --georef, -n, or --nogeoref
to modify this behavior.

Options

    -c
    --children
      If set, the speciemens associated with the indicated taxon, as well
      as the ones from its descendants, will be printed.

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    specimens from gbif.

    -g
    --georef
      If set, only georeferenced records will be printed.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -n
    --nonref
      If defined, only records without a georeference will be printed. It
      will be ignored if -g, --georef is defined.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --country name
      If set, only specimens reported to the given country will be printed.

    -t value
    --taxon value
      Search for the indicated taxon id.

    -v
    --verbose
      If defined, then a large list (including ids) will be printed. This
      option is ignored if -m or --machine option is defined.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Add specimens from an extern database

Synopsis

    jdh sp.pop -e|--extdb name [-p|--port value] [-r|--rank name]
	[-t|--taxon value] [<name> [<parentname>]]

Description

Sp.pop uses an extern database to populate the local database with specimens.

When the options -t, --taxon or a name are used, the effect of the command
will only affect the indicated taxon and its descendants.

When the -r, --rank option is used, only the taxons at, or below the indicated
rank will be populated.

Options

    -e name
    --extdb name
      Sets the extern database.
      Valid values are:
          gbif    specimens from gbif.
      This parameter is required.

    -g
    --georef
      If set, only the specimens georeferenced in the extern database
      will be added.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set, only taxons below the indicated rank will be populated.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -t value
    --taxon value
      Search for the indicated taxon id.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -t or --taxon are
      defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -t or --taxon
      are defined.

Sets an specimen value

Synopsis

    jdh sp.set -i|--id value [-p|--port value] [<key=value>...]

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

Deletes a taxon

Synopsis

    jdh tx.del [-c|--collapse] [-i|--id value] [-p|--port value]
	[<name> [<parentname>]]

Description

Tx.del removes a taxon from the database, by default it deletes the taxon,
and all of its descendants. If the option -c, --collapse is defined, then
the taxon will be "collapsed": all of its descendants will be assigned to
the the taxon's parent, and then, it will be deleted.

Options

    -c
    --collapse
      Collapse the taxon: all the descendants (including synonyms) will be
      assigned to the ancestor of the taxon. If the taxon has no ancestor,
      then the valid descendants will be assigned to the root of the
      taxonomy and synonyms will be deleted with the taxon.

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Enforces a ranked taxonomy

Synopsis

    jdh tx.force [-i|--id value] [-p|--port value] [-r|--rank name]
	[<name> [<parentname>]]

Description

Tx.force enforces the database to be ranked, synonymizing all rankless taxa
with their most inmmediate ranked taxon.

Options

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set search only for taxons below the indicated rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Imports taxon data

Synopsis

    jdh tx.in [-a|--anc value] [-f|--format value] [-p|--port value]
	[-r|--rank name] [-s|--synonym] [-v|--verbose] [<file>...]

Description

Tx.in reads taxon data from the indicated files, or the standard input
(if no file is defined), and adds them to the jdh database.

Default input format is txt. If the format is txt, it is assummed that each
line corresponds to a taxon (lines starting with '#' or ';' will be ignored).

By default, taxons will be added to the root of the taxonomy, valid, and
unranked.

Options

    -a value
    --anc value
      Sets the parent of the added taxons. The value must be a valid id.

    -f value
    --format value
      Sets the format used in the source data. Valid values are:
          txt        Txt format

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      Set the rank of the added taxon. If the taxon has a parent (the -a,
      --anc options) the parent must be concordant with the given rank.
      Valid values are:
      	  unranked
          kingdom
          class
          order
          family
          genus
          species

    -s
    --synonym
      If set, the added taxons will be set as synonym of its parent. It
      requires that a valid parent will be defined (-n, --pname, -p or
      --parent options)

    -v
    --verbose
      If set, the name and id of each added taxon will be print in the
      standard output.

    <file>
      One or more files to be proccessed by tx.in. If no file is given
      then the information is expected to be from the standard input.

Prints general taxon information

Synopsis

    jdh tx.info [-e|--extdb name] [-i|--id value] [-k|--key value]
	[-m|--machine] [-p|--port value] [<name> [<parentname>]]

Description

Tx.info prints general information of a taxon in the database. For the
list of parents, descendants of synonyms of the taxon, use 'jdh tx.ls'.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).

    -i value
    --id value
      Search for the indicated taxon id.

    -k value
    --key value
      If set, only a particular value of the taxon will be printed.
      Valid keys are:
          authority      Authorship of the taxon.
          comment        A free text comment on the taxon.
          extern         Extern identifiers of the taxon, in the form
                         <service>:<key>, for example: gbif:5216933.
          name           Name of the taxon.
          parent         Id of the new parent.
          rank           The taxon rank.
          synonym        Prints the parent of the taxon, if it is a synonym.
                         If the taxon is valid, the valid string will be printed.
          valid          See synonym.

    -m
    --machine
      If set, the output will be machine readable. That is, just key=value pairs
      will be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Prints a list of taxons

Synopsis

    jdh tx.ls [-a|--ancs] [-e|--extdb name] [-i|--id value]
	[-m|--machine] [-p|--port value] [-r|--rank name] [-s|--synonym]
	[-v|--verbose] [<name> [<parentname>]]

Description

Tx.ls prints a list of taxons. With no option, tx.ls prints the taxons attached
to the root of taxonomy.

If a name or -i, --id option is defined, then the descendants of the indicated
taxon will be printed by default. This behaviour can be changed with other
options (e.g. -a, --ancs to show parents).

Options

    -a
    --ancs
      If set, the parents of the indicated taxon will be printed.

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).

    -i value
    --id value
      Search for the indicated taxon id.

    -m
    --machine
      If set, the output will be machine readable. That is, just ids will
      be printed.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If indicated, the only taxons at the given rank will be printed. Valid
      values are:
          unranked
          kingdom
          class
          order
          family
          genus
          species

    -s
    --synonym
      If set, the synonyms of the indicated taxon will be printed.

    -v
    --verbose
      If defined, then a large list (including ids and authors) will be
      printed. This option is ignored if -m or --machine option is defined.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Sets a taxon value

Synopsis

    jdh tx.set [-i|--id value] [-p|--port value] [<name> [<parentname>]]
	[<key=value>...]

Description

Tx.set sets a particular value for a taxon in the database. Use this
command to edit the taxon database, instead of manual edition.

If no key is defined, the key values will be read from the standard input,
it is assumed that each line is in the form:
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

Updates local database using an extern database

Synopsis

    jdh tx.sync -e|--extdb name [-d|--validate] [-i|--id value]
	[-l|--populate name] [-m|--match] [-p|--port value] [-r|--rank name]
	[-u|--update] [-v|--verbose] [<name> [<parentname>]]

Description

Tx.sync uses an extern database to update, validate or set the
local database.

When the options -i, --id or a name are used, the effect of the update
will only affect the indicated taxon and its descendants.

Except for the -m, --match option, all other options assume that the
extern ids for each taxon is already defined. The -m, --match option
will try to scan the extern database to match current taxon names with
names in the extern database.

Options

    -d
    --validate
      If set, and -u, --update option is defined, then the validity of the
      taxon will be set as in the extern database.

    -e name
    --extdb name
      Set the extern database.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).
      This parameter is required.

    -i value
    --id value
      Update only the indicated taxon (and its descendants).

    -l name
    --populate name
      If set, then the taxons below the indicated rank will be populated
      with all descendants (valid and invalid) from the extern database.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -m
    --match
      If set, it will search in the extern database for each name in the
      local database that has no assigned extern id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -r name
    --rank name
      If set, then it moves taxons to their respective rank. If the parent
      taxon is not in the database it will be added. The name value
      indicates the maximum rank to search, for example '--rank=class' only
      assign taxons up to class rank.
      Valid values are:
          kingdom
          class
          order
          family
          genus
          species

    -u
    --update
      If found, it will update the authorship, rank of a taxon, following
      the external database. If the option -d, --validate is defined, also
      sets the taxon status, and for synonyms, new parents will be added
      as needed.

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Prints taxonomy

Synopsis

    jdh tx.taxo [-e|--extdb name] [-f|--format name]	[-i|--id value]
	[-p|--port value] [-s|--simple]	[<name> [<parentname>]]

Description

Tx.taxo prints the taxonomy of the indicated taxon in the format of a
taxonomic catalog.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).

    -f
    --format
      Sets the output format, by default it will use txt format.
      Valid values are:
          txt     text format
          html    html format

    -i value
    --id value
      Search for the indicated taxon id.

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    <name>
      Search for the indicated name. If there are more than one taxon,
      then the list of possible candidates will be printed and the
      program will be terminated. Ignored if option -i or --id are defined.

    <parentname>
      If defined, the taxon search with <name> will be limited to
      descendants of the indicated name. Ignored if option -i or --id
      are defined.

Quick start guide

Synopsis

    jdh start

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

*/
package main
