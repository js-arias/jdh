// Authomatically generated doc.go file for use with godoc.

/*
Joseph Dalton Hooker GUI.

Synopsis

    jdh.gui [help] <command> [<args>...]

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

Displays a tree

Synopsis

    jdh.gui tr.view [-p|--port value] [-s|--set]

Description

Tr.view displays a tree. By default the tree will be fitted to the window.

If there are more than one tree in the database, you can use space, enter keys
to move to the next tree, and backspace to move to previous tree.

Options

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"

    -s
    --set
      If set, the tree can be edited with the mouse.

*/
package main
