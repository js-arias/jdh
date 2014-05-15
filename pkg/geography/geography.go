// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

// Package geography implements some geographic utilities.
package geography

import (
	"errors"
	"fmt"
	"math"
)

// Location is an specific geographic location. A location can be
// georeferenced.
type Location struct {
	Country  Country
	State    string
	County   string
	Locality string
	GeoRef   Georeference
}

// IsValid returns true if the location is a valid one (i.e. with both
// Country and Locality fields filled. It must be noted that a location
// can return invalid, but it still have a valid georeference, so the
// validity is only a term to be used in the context of geolocation
// assignation or validation using a Geolocater interface.
func (l *Location) IsValid() bool {
	if (len(l.Country) == 0) || (len(l.Locality) == 0) {
		return false
	}
	return true
}

// Georeference is the information of a particular georeference, including
// the validation status of the georeference.
type Georeference struct {
	Point       Point  // geodetic longitude and latitude
	Uncertainty uint   // georeference uncertainty in meters
	Source      string // source of georeference
	Validation  string // source of the georeference validation
}

// Maximum and minimum values for geographic coordinates
const (
	MinLon = -180
	MaxLon = 180
	MinLat = -90
	MaxLat = 90
)

// Point is a geographic point with geodetic coordinates
type Point struct {
	Lon, Lat float64
}

// InvalidPoint is an imposible geographic point
func InvalidPoint() Point {
	return Point{360, 360}
}

// IsValid returns true if the geographic point has valid coordinates.
func (p *Point) IsValid() bool {
	return IsLon(p.Lon) && IsLat(p.Lat)
}

// IsLon returns true if the value is a valid longitude.
func IsLon(val float64) bool {
	if (val < MaxLon) && (val >= MinLon) {
		return true
	}
	return false
}

// IsLat return true if the value is a valid latitude.
func IsLat(val float64) bool {
	if (val < MaxLat) && (val > MinLat) {
		return true
	}
	return false
}

// WGS84 mean radius (in meters).
const EarthRadius = 6371009

// Distance calculates the great circle distance between two points in meters
// using the WGS84 ellipsoid.
func (p *Point) Distance(lon, lat float64) uint {
	l1, l2 := toRad(p.Lat), toRad(lat)
	dLon := toRad(p.Lon) - toRad(lon)
	dLat := l1 - l2
	s1 := math.Sin(dLat / 2)
	s1 *= s1
	f1 := math.Sin(dLon / 2)
	s2 := f1 * f1 * math.Cos(l1) * math.Cos(l2)
	v := math.Sqrt(s1 + s2)
	return uint(2 * EarthRadius * math.Asin(v))
}

func toRad(angle float64) float64 {
	return angle * math.Pi / 180
}

// Errors from the Gazetter service
var (
	ErrAmbiguous = errors.New("Ambiguous location")
	ErrNotInDB   = errors.New("Location not in database")
	ErrNoLoc     = errors.New("Expecting valid county, locality")
	ErrInvalid   = errors.New("Invalid point")
	ErrClosed    = errors.New("Closed gazetter")
)

// Gazetter is a georeferencing service.
type Gazetter interface {
	// Name of the service.
	Name() string

	// Close the service.
	Close()

	// Locate returns the point and uncertainty associated with a
	// location as interpreted by the geolocation service. Uncertainty
	// indicates the maximum uncertainty (in meters) accepted for the
	// point (with 0 any uncertainty will be accepted).
	Locate(loc *Location, uncertainty uint) (Georeference, error)

	// List returns a list of points that fullfill a given location.
	List(loc *Location, uncertainty uint) ([]Georeference, error)
}

// gazette holds the information of a geolocation service.
type gazette struct {
	name string
	open func(par string) (Gazetter, error)
}

// services holds the list geolocation services.
var services []gazette

// RegisterGazetter registers a geolocation service by its name.
func RegisterGazetter(name string, open func(par string) (Gazetter, error)) {
	if open == nil {
		panic("geography: open function is nil")
	}
	if len(name) == 0 {
		panic("geography: empty service name")
	}
	for _, s := range services {
		if s.name == name {
			panic("geography: RegisterGeolocater called twice for driver " + name)
		}
	}
	services = append(services, gazette{name, open})
}

// OpenGazetter opens a geolocation service by its service name.
func OpenGazetter(name, param string) (Gazetter, error) {
	for _, s := range services {
		if s.name == name {
			return s.open(param)
		}
	}
	return nil, fmt.Errorf("location servie %s unregistred", name)
}
