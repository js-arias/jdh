// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package geolocate implements the geolocate gazetteer service.
package geolocate

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/js-arias/jdh/pkg/geography"
)

const gazette = "geolocate"

func init() {
	geography.RegisterGazetter(gazette, open)
}

type locAnswer struct {
	Features []feature
}

type feature struct {
	Geometry   point
	Properties property
}

type point struct {
	Coordinates []float64
}

type property struct {
	UncertaintyRadiusMeters uint
	Debug                   string
}

// service implements the geolocate web service.
type service struct {
	isClosed bool
	request  chan string
	answer   chan interface{}
}

// Open opens the geolocate service.
func open(param string) (geography.Gazetter, error) {
	s := &service{
		request: make(chan string),
		answer:  make(chan interface{}),
	}
	go s.req()
	return s, nil
}

// Name returns the name of the location service.
func (s *service) Name() string {
	return "GEOLocate Web Service"
}

// Close closes the service.
func (s *service) Close() {
	if s.isClosed {
		return
	}
	close(s.request)
}

// Locate returns the point and uncertainty associated with a location as
// interpreted by the geolocation service. Uncertainty indicates the maximum
// uncertainty (in meters) accepted for the point (with 0 any uncertainty
// will be accepted).
func (s *service) Locate(l *geography.Location, locality string, uncertainty uint) (geography.Georeference, error) {
	ls, err := s.List(l, locality, uncertainty)
	if err != nil {
		return geography.Georeference{Point: geography.InvalidPoint()}, err
	}
	if len(ls) == 1 {
		return ls[0], nil
	}
	if len(ls) == 0 {
		return geography.Georeference{Point: geography.InvalidPoint()}, geography.ErrNotInDB
	}
	u := uncertainty
	if u == 0 {
		u = 200000
	}
	// set a mid point
	var sLon, sLat float64
	unc := uint(0)
	for _, p := range ls {
		sLon += p.Point.Lon
		sLat += p.Point.Lat
		if unc < p.Uncertainty {
			unc = p.Uncertainty
		}
	}
	den := float64(len(ls))
	lon, lat := sLon/den, sLat/den
	if !(geography.IsLon(lon) && geography.IsLat(lat)) {
		return geography.Georeference{Point: geography.InvalidPoint()}, geography.ErrAmbiguous
	}
	max := uint(0)
	for _, p := range ls {
		d := p.Point.Distance(lon, lat)
		if d > max {
			max = d
			if (d + unc) > u {
				break
			}
		}
	}
	p := ls[0]
	p.Uncertainty = max + unc
	if (p.Uncertainty > u) || (!p.IsValid()) {
		return geography.Georeference{Point: geography.InvalidPoint()}, geography.ErrAmbiguous
	}
	p.Point = geography.Point{lon, lat}
	return p, nil
}

// List returns a list of points that fullfill a given location.
func (s *service) List(l *geography.Location, locality string, uncertainty uint) ([]geography.Georeference, error) {
	if s.isClosed {
		return nil, geography.ErrClosed
	}
	if !l.IsValid() {
		return nil, geography.ErrNoLoc
	}
	var ls []geography.Georeference
	var err error
	if len(locality) > 0 {
		ls, err = s.list(l, locality, uncertainty)
		if err != nil {
			if err != geography.ErrNotInDB {
				return nil, err
			}
		}
		if len(ls) > 0 {
			return ls, nil
		}
	}
	if len(l.County) > 0 {
		return s.list(l, l.County, uncertainty)
	}
	if err != nil {
		return nil, err
	}
	if len(locality) == 0 {
		return nil, geography.ErrNoLoc
	}
	return ls, nil
}

func (s *service) list(l *geography.Location, locality string, uncertainty uint) ([]geography.Georeference, error) {
	req := wsHead + prepare(l, locality)
	s.request <- req
	a := <-s.answer
	switch answer := a.(type) {
	case error:
		return nil, answer
	case *http.Response:
		defer answer.Body.Close()
		d := json.NewDecoder(answer.Body)
		an := &locAnswer{}
		if err := d.Decode(an); err != nil {
			return nil, err
		}
		var ls []geography.Georeference
		for _, f := range an.Features {
			if uncertainty == 0 {
				ls = append(ls, geography.Georeference{
					Point: geography.Point{
						Lon: f.Geometry.Coordinates[0],
						Lat: f.Geometry.Coordinates[1],
					},
					Uncertainty: f.Properties.UncertaintyRadiusMeters,
					Source:      "online gazetter: " + s.Name(),
					Validation:  s.Name(),
				})
				continue
			}
			if f.Properties.UncertaintyRadiusMeters < uncertainty {
				ls = append(ls, geography.Georeference{
					Point: geography.Point{
						Lon: f.Geometry.Coordinates[0],
						Lat: f.Geometry.Coordinates[1],
					},
					Uncertainty: f.Properties.UncertaintyRadiusMeters,
					Source:      "online gazetter: " + s.Name(),
					Validation:  s.Name(),
				})
			}
		}
		return ls, nil
	}
	return nil, geography.ErrNotInDB
}

func prepare(l *geography.Location, locality string) string {
	vals := url.Values{}
	vals.Add("country", l.Country.Name())
	vals.Add("locality", locality)
	if len(l.State) > 0 {
		vals.Add("state", l.State)
	}
	if len(l.County) > 0 {
		vals.Add("county", l.County)
	}
	vals.Add("enableH20", "false")
	vals.Add("hwyX", "false")
	vals.Add("fmt", "geojson")
	return vals.Encode()
}

const wsHead = "http://www.museum.tulane.edu/webservices/geolocatesvcv2/glcwrap.aspx?"

// Req process requests.
func (s *service) req() {
	for r := range s.request {
		answer, err := http.Get(r)
		if err != nil {
			s.answer <- err
			continue
		}
		s.answer <- answer

		// this is set to not overload the geolocate server...
		time.Sleep(100 * time.Millisecond)
	}
}
