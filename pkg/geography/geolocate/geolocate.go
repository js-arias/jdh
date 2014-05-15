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

func (s *service) Name() string {
	return "GEOLocate Web Service"
}

func (s *service) Close() {
	if s.isClosed {
		return
	}
	close(s.request)
}

func (s *service) Locate(l *geography.Location, uncertainty uint) (geography.Georeference, error) {
	ls, err := s.List(l, uncertainty)
	if err != nil {
		return geography.Georeference{Point: geography.InvalidPoint()}, err
	}
	if len(ls) > 1 {
		return geography.Georeference{Point: geography.InvalidPoint()}, geography.ErrAmbiguous
	}
	return ls[0], nil
}

func (s *service) List(l *geography.Location, uncertainty uint) ([]geography.Georeference, error) {
	if s.isClosed {
		return nil, geography.ErrClosed
	}
	if !l.IsValid() {
		return nil, geography.ErrNoLoc
	}
	req := wsHead + prepare(l)
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
					Source:      "online gazetter",
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
					Source:      "online gazetter",
					Validation:  s.Name(),
				})
			}
		}
		return ls, nil
	}
	return nil, geography.ErrNotInDB
}

func prepare(l *geography.Location) string {
	vals := url.Values{}
	vals.Add("country", l.Country.Name())
	vals.Add("locality", l.Locality)
	if len(l.State) > 0 {
		vals.Add("state", l.State)
	}
	if len(l.County) > 0 {
		vals.Add("county", l.County)
	}
	vals.Add("enableH20", "false")
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
