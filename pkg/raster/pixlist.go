// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

package raster

import "image"

// A PixList raster is a raster in which pixels are stored as a simple list
// of the cells in which the pixel have a non 0 value.
type PixList struct {
	// List of the pixels, it is not guaranteed that the list is ordered.
	Pixel []Pixel

	// Rect is the raster bounds.
	Rect image.Rectangle
}

// NewPixList returns a new empty PixList.
func NewPixList() *PixList {
	return &PixList{}
}

// Bounds are the bounds of the PixList.
func (p *PixList) Bounds() image.Rectangle {
	return p.Rect
}

// At returns the value of the indicated point.
func (p *PixList) At(pt image.Point) int {
	if !pt.In(p.Rect) {
		return 0
	}
	for _, px := range p.Pixel {
		if px.Eq(pt) {
			return px.Value
		}
	}
	return 0
}

// Set sets the value of the indicated point.
func (p *PixList) Set(pt image.Point, val int) {
	if val == 0 {
		if len(p.Pixel) == 0 {
			return
		}
		if !pt.In(p.Rect) {
			return
		}
		isIn := false
		for i, px := range p.Pixel {
			if px.Eq(pt) {
				p.Pixel[i] = p.Pixel[len(p.Pixel)-1]
				p.Pixel[len(p.Pixel)-1] = Pixel{}
				p.Pixel = p.Pixel[:len(p.Pixel)-1]
				isIn = true
				break
			}
		}
		if !isIn {
			return
		}
		if len(p.Pixel) == 0 {
			p.Rect = image.Rectangle{}
			return
		}
		if len(p.Pixel) == 1 {
			p.Rect.Min.X, p.Rect.Min.Y = p.Pixel[0].X, p.Pixel[0].Y
			p.Rect.Max.X, p.Rect.Max.Y = p.Pixel[0].X, p.Pixel[0].Y
			return
		}
		if (pt.X == p.Rect.Min.X) || (pt.X == p.Rect.Max.X) || (pt.Y == p.Rect.Min.Y) || (pt.Y == p.Rect.Max.Y) {
			minX, maxX := p.Rect.Max.X, p.Rect.Min.X
			minY, maxY := p.Rect.Max.Y, p.Rect.Min.Y
			for _, px := range p.Pixel {
				if px.X > maxX {
					maxX = px.X + 1
				}
				if px.X < minX {
					minX = px.X
				}
				if px.Y > maxY {
					maxY = px.Y + 1
				}
				if px.Y < minY {
					minY = px.Y
				}
			}
			p.Rect = image.Rect(minX, minY, maxX, maxY)
		}
		return
	}
	if !pt.In(p.Rect) {
		p.Pixel = append(p.Pixel, Pixel{image.Point{X: pt.X, Y: pt.Y}, val})
		if len(p.Pixel) == 1 {
			p.Rect.Min.X, p.Rect.Min.Y = p.Pixel[0].X, p.Pixel[0].Y
			p.Rect.Max.X, p.Rect.Max.Y = p.Pixel[0].X, p.Pixel[0].Y
			return
		}
		minX, maxX := p.Rect.Max.X, p.Rect.Min.X
		minY, maxY := p.Rect.Max.Y, p.Rect.Min.Y
		for _, px := range p.Pixel {
			if px.X > maxX {
				maxX = px.X + 1
			}
			if px.X < minX {
				minX = px.X
			}
			if px.Y > maxY {
				maxY = px.Y + 1
			}
			if px.Y < minY {
				minY = px.Y
			}
		}
		p.Rect = image.Rect(minX, minY, maxX, maxY)
		return
	}
	for _, px := range p.Pixel {
		if px.Eq(pt) {
			px.Value = val
			return
		}
	}
	p.Pixel = append(p.Pixel, Pixel{image.Point{X: pt.X, Y: pt.Y}, val})
}
