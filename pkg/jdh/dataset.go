// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package jdh

// Dataset is a museum collection, a published dataset, or any other
// source of specimen records.
type Dataset struct {
	// identifier of the dataset.
	Id string

	// title of the dataset.
	Title string

	// simple citation of the dataset.
	Citation string

	// license or rights of use of the dataset.
	License string

	// url of the dataset, if any
	Url string

	// extern identifiers of the dataset.
	Extern []string

	// free text comment about the dataset.
	Comment string
}

// Datasets is the table that store dataset information.
const Datasets Table = "datasets"

// Key values used in Datasets table.
const (
	// Preferred citation of the dataset, as a free string. An empty
	// value will delete the content of the citation field.
	DataCitation Key = "citation"

	// License or rights of use of the data included in the dataset. An
	// empty value will delete the content of the license field.
	DataLicense = "license"

	// Title of the dataset.
	DataTitle = "title"

	// Url of the dataset. An empty value will delete the url.
	DataUrl = "url"
)
