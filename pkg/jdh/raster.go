// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package jdh

import (
	"strings"

	"github.com/js-arias/jdh/pkg/raster"
)

// Raster holds the information about a rasterized taxon distribution.
type Raster struct {
	// identifier of the raster.
	Id string

	// id of the taxon used in the raster.
	Taxon string

	// source of the raster.
	Source RasterSource

	// a reference of the raster.
	Reference string

	// number of colums the raster will have if it occupies all earth
	// (i.e. a 360 degrees span). The number of rows is just this number
	// divided by two.
	Cols uint

	// stored raster.
	Raster *raster.PixList

	// extern identifiers of the raster.
	Extern []string

	// free text comment about the raster.
	Comment string
}

// RasterSource indicates how the raster was adquired.
type RasterSource int

// Valid RasterSource values.
const (
	UnknownRaster RasterSource = iota

	// a raster from explicit georeferenced points. Its values are expected
	// to be booleans (0 or 1).
	ExplicitPoints

	// a raster from expert's opinion about taxon's distribution. Its
	// values are expected to be booleans (0 or 1).
	ExpertOpinion

	// a raster from a predictive distribution algorithm. Its values
	// are expected to be between 0 and 1000.
	MachineModel
)

// rasSource holds a list of the raster type names accepted in jdh.
var rasSource = []string{
	"unknown",
	"explicit points",
	"expert opinion",
	"machine model",
}

// GetRasterSource returns the id of a raster source.
func GetRasterSource(s string) RasterSource {
	s = strings.ToLower(s)
	for i, sr := range rasSource {
		if sr == s {
			return RasterSource(i)
		}
	}
	return UnknownRaster
}

// String returns the raster source for a guiven RasterSource id.
func (sr RasterSource) String() string {
	i := int(sr)
	if i > len(rasSource) {
		return rasSource[0]
	}
	return rasSource[i]
}

// RasDistros is the table that store the rasterized distributions.
const RasDistros Table = "rasdistros"

// Key values used in the rasterized distributions table.
const (
	// Number of columns of the raster. This value can not be setted.
	RDisCols Key = "cols"

	// Source of the raster using a string value of RasterSource.
	RDisSource = "source"

	// Taxon id of the taxon associated with a raster. Used in delete
	// operation will delete all the rasters associated with a taxon id.
	// Used in list operation to retrieve all the rasters directly
	// identified with the indicated id.
	RDisTaxon = "taxon"

	// Used in list operations to retrieve all the rasters associated
	// with a taxon id, or any of its descendants.
	RDisTaxonParent = "parent"
)

// Key values used for the raster of a rasterized distribution.
const (
	// Sets a pixel value, in the form "X,Y,Val"
	RasPixel Key = "pixel"
)
