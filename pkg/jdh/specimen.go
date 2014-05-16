// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package jdh

import (
	"strings"
	"time"

	"github.com/js-arias/jdh/pkg/geography"
)

// Specimen holds the information of a collection specimen.
type Specimen struct {
	// identifier of the specimen.
	Id string

	// id of the taxon assigned to the specimen.
	Taxon string

	// basis of the recorded specimen.
	Basis BasisOfRecord

	// A reference of the specimen.
	Reference string

	// Dataset is the id of the dataset that holds the specimen
	// information.
	Dataset string

	// catalog code of the specimen, it is expected to be in the format
	// <instituion acronym>:<collection acronym>:<catalog id>.
	Catalog string

	// the person who determined the identity the specimen.
	Determiner string

	// the person who collected the specimen.
	Collector string

	// the time and date of the collection event.
	Date time.Time

	// general geographic location data of the collection event.
	Geography geography.Location

	// locality of the collection event.
	Locality string

	// georeference of the specimen, if any.
	Georef geography.Georeference

	// extern identifiers of the specimen.
	Extern []string

	// free text comment about the specimen.
	Comment string
}

// BasisOfRecord is the id of a kind of specimen.
type BasisOfRecord uint

// Valid BasisOfRecord values.
const (
	UnknownBasis BasisOfRecord = iota
	Preserved                  // a preserved specimen
	Fossil                     // a fossilized specimen
	Observation                // a human observation
	Remote                     // a remote observation
)

// basis holds a list of record names accepted in jdh.
var basis = []string{
	"unkown",
	"preserved specimen",
	"fossil",
	"observation",
	"remote",
}

// GetBasisOfRecord returns the id of a record.
func GetBasisOfRecord(s string) BasisOfRecord {
	s = strings.ToLower(s)
	for i, b := range basis {
		if b == s {
			return BasisOfRecord(i)
		}
	}
	return UnknownBasis
}

// String returns the bais of record string of a given BasisOfRecord id.
func (b BasisOfRecord) String() string {
	i := int(b)
	if i >= len(basis) {
		return basis[0]
	}
	return basis[i]
}

// Specimens is the table that store the specimen information.
const Specimens Table = "specimens"

// Key values used in specimens table.
const (
	// Basis of the specimen record. It must be a string expression of
	// a valid BasisOfRecord accepted in jdh.
	SpeBasis Key = "basis"

	// The catalog code of the specimen. Usually is expected to be in the
	// form <inst>:<coll>:<catalog>.
	SpeCatalog = "catalog"

	// The person who collected the specimen.
	SpeCollector = "collector"

	// The dataset that kepts the specimen data.
	SpeDataset = "dataset"

	// The time of the collection, using the ISO 8601 layout.
	SpeDate = "date"

	// The person who determined the identity the specimen.
	SpeDeterminer = "determiner"

	// Point locality of the specimen sample.
	SpeLocality = "locality"

	// Used in list operations to retrieve only records with or without
	// a georeference. Valid values are "true" and "false"
	SpeGeoref = "georef"

	// A reference to the specimen.
	SpeReference = "reference"

	// Taxon id of the taxon associated with a specimen. Used in delete
	// operation will delete all the specimens associated with a taxon id.
	// Used in list operation to retrieve all the specimens directly
	// identied with the indicated id.
	SpeTaxon = "taxon"

	// Used in list operations to retrieve all the specimens associated
	// with a taxon id, or any of its descendants.
	SpeTaxonParent = "parent"
)

// Key values used for geography of an specimen.
const (
	// Country of the specimen sample location, using ISO 3166-1 alpha-2 code.
	GeoCountry Key = "coutry"

	// County, Municipality (or equivalent) of the specimen sample
	// location.
	GeoCounty = "county"

	// Longitude and latitude of the location, in the form "lon,lat"
	GeoLonLat = "lonLat"

	// Source of the georeference, for example, gps data or a gazetter.
	GeoSource = "source"

	// State, Province (or equivalent) of the specimen sample location.
	GeoState = "state"

	// Uncertainty of the georeference in meters.
	GeoUncertainty = "uncertainty"

	// Validation of the georeference.
	GeoValidation = "validation"
)
