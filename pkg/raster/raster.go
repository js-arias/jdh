// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

// Package raster implements a different taxon distribution raster kinds and
// their operations.
//
// The rasters are expected to represent taxon's distributions.
//
// Where comparing rasters, it is assumed that they are in the same scale, so
// control on the raster scale must be done outside the package.
package raster

import "image"

// Raster defines a rectangular area that represent a taxon distribution.
type Raster interface {
	// Bounds returns the boundaries of the raster in which valid pixels
	// are found, any At operation outside the Bounds will always returns
	// a zero value.
	Bounds() image.Rectangle

	// At returns the value of the indicated point.
	At(image.Point) int

	// Set sets a value of the indicated point. If the point is outside
	// the bounds, and the value is not zero, or point is zero and located
	// at the bounds, then the bounds will be redefined.
	Set(image.Point, int)
}

// Pixel is a point with a defined integer value.
type Pixel struct {
	image.Point

	// The pixel value.
	Value int
}
