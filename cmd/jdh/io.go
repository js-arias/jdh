// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"strings"
	"unicode"
)

// ReadLine reads a line from an input buffer.
func readLine(in *bufio.Reader) ([]string, error) {
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if len(line) == 0 {
				return nil, err
			}
		}
		args := strings.Fields(line)
		// skip empty lines and comments
		if (len(args) == 0) || (args[0][0] == '#') || (args[0][0] == ';') {
			continue
		}
		return args, nil
	}
}

// ReadString reads a string from an input buffer.
func readString(in *bufio.Reader) (string, error) {
	if err := skipSpaces(in); err != nil {
		return "", err
	}
	tok := make([]rune, 0)
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return "", err
		}
		if unicode.IsSpace(r) {
			return string(tok), nil
		}
		tok = append(tok, r)
	}
}

// SkipSpaces skip spaces from an input buffer.
func skipSpaces(in *bufio.Reader) error {
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return err
		}
		if !unicode.IsSpace(r) {
			in.UnreadRune()
			return nil
		}
	}
}

// PeeKNext peek next valid (non space) rune from an input buffer.
func peekNext(in *bufio.Reader) (rune, error) {
	if err := skipSpaces(in); err != nil {
		return rune(0), err
	}
	r, _, err := in.ReadRune()
	if err != nil {
		return rune(0), err
	}
	in.UnreadRune()
	return r, nil
}

// ReadBlock reads a delimited block from the input buffer.
func readBlock(in *bufio.Reader, end rune) (string, error) {
	tok := make([]rune, 0)
	for {
		r, _, err := in.ReadRune()
		if err != nil {
			return "", err
		}
		if r == end {
			return string(tok), nil
		}
		tok = append(tok, r)
	}
}
