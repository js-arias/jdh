// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package gbif

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/jdh/pkg/geography"
	"github.com/js-arias/jdh/pkg/jdh"
)

var basis = []string{
	"UNKNOWN",
	"PRESERVED_SPECIMEN",
	"FOSSIL_SPECIMEN",
	"OBSERVATION",
}

func getBasis(s string) jdh.BasisOfRecord {
	for i, b := range basis {
		if b == s {
			return jdh.BasisOfRecord(i)
		}
	}
	return jdh.UnknownBasis
}

type occAnswer struct {
	Offset, Limit int64
	EndOfRecords  bool
	Results       []*occurrence
}

type occurrence struct {
	Key                 int64  // id
	TaxonKey            int64  // taxon
	BasisOfRecord       string // basis
	DatasetKey          string // collection
	InstitutionCode     string
	CollectionCode      string
	CatalogNumber       string
	IdentifierName      string // indetifiedby
	CollectorName       string // collector
	OccurrenceDate      string // date
	Country             string // country
	StateProvince       string // state
	County              string // county
	Locality            string // locality
	DecimalLongitude    float64
	DecimalLatitude     float64
	GeoreferenceSources string
	// comments
	FieldNotes        string
	OccurrenceRemarks string
}

func (o *occurrence) copy() *jdh.Specimen {
	cat := strings.TrimSpace(o.InstitutionCode)
	cat += ":" + strings.TrimSpace(o.CollectionCode)
	cat += ":" + strings.TrimSpace(o.CatalogNumber)
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", o.OccurrenceDate)
	spe := &jdh.Specimen{
		Id:         strconv.FormatInt(o.Key, 10),
		Taxon:      strconv.FormatInt(o.TaxonKey, 10),
		Basis:      getBasis(o.BasisOfRecord),
		Dataset:    strings.TrimSpace(o.DatasetKey),
		Catalog:    cat,
		Determiner: strings.Join(strings.Fields(o.IdentifierName), " "),
		Collector:  strings.Join(strings.Fields(o.CollectorName), " "),
		Date:       t,
		Location: geography.Location{
			Country:  geography.GetCountry(o.Country),
			State:    strings.Join(strings.Fields(o.StateProvince), " "),
			County:   strings.Join(strings.Fields(o.County), " "),
			Locality: strings.Join(strings.Fields(o.Locality), " "),
		},
		Comment: strings.TrimSpace(o.FieldNotes + "\n" + o.OccurrenceRemarks),
	}
	lon, lat := float64(360), float64(360)
	if o.DecimalLongitude != 0 {
		lon = o.DecimalLongitude
	}
	if o.DecimalLatitude != 0 {
		lat = o.DecimalLatitude
	}
	if geography.IsLon(lon) && geography.IsLat(lat) {
		spe.Location.GeoRef.Point = geography.Point{Lon: lon, Lat: lat}
		spe.Location.GeoRef.Source = strings.Join(strings.Fields(o.GeoreferenceSources), " ")
	} else {
		spe.Location.GeoRef.Point = geography.InvalidPoint()
	}
	return spe
}

func (db *DB) occurrences(kvs []jdh.KeyValue) (jdh.ListScanner, error) {
	l := &listScanner{
		c:   make(chan interface{}, 20),
		end: make(chan struct{}),
	}
	id := ""
	var tax int64
	for _, kv := range kvs {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.SpeTaxon {
			id = strings.TrimSpace(kv.Value[0])
			var err error
			tax, err = strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, err
			}
			break
		}
		if kv.Key == jdh.SpeTaxonParent {
			id = strings.TrimSpace(kv.Value[0])
			break
		}
	}
	if len(id) == 0 {
		return nil, errors.New("taxon " + id + " without [ny] identification")
	}
	go func() {
		vals := url.Values{}
		vals.Add("taxonKey", id)
		vals.Add("basisOfRecord", "PRESERVED_SPECIMEN")
		vals.Add("basisOfRecord", "FOSSIL_SPECIMEN")
		for _, kv := range kvs {
			if len(kv.Value) == 0 {
				continue
			}
			switch kv.Key {
			case jdh.LocCountry:
				for _, v := range kv.Value {
					vals.Add("country", string(geography.GetCountry(v)))
				}
			case jdh.LocGeoRef:
				if kv.Value[0] == "true" {
					vals.Set("has_coordinate", "true")
				} else if kv.Value[0] == "false" {
					vals.Set("has_coordinate", "false")
				}
			}
		}
		for off := int64(0); ; {
			if off > 0 {
				vals.Set("offset", strconv.FormatInt(off, 10))
			}
			request := wsHead + "occurrence/search?" + vals.Encode()
			an := new(occAnswer)
			if err := db.listRequest(request, an); err != nil {
				l.setErr(err)
				return
			}
			for _, oc := range an.Results {
				if tax > 0 {
					if oc.TaxonKey != tax {
						continue
					}
				}
				select {
				case l.c <- oc.copy():
				case <-l.end:
					return
				}
			}
			if an.EndOfRecords {
				break
			}
			off += an.Limit
		}
		select {
		case l.c <- nil:
		case <-l.end:
		}
	}()
	return l, nil
}

// specimen returns a jdh scanner with an specimen.
func (db *DB) specimen(id string) (jdh.Scanner, error) {
	if len(id) == 0 {
		return nil, errors.New("specimen without identification")
	}
	o, err := db.getSpecimen(id)
	if err != nil {
		return nil, err
	}
	return &getScanner{val: o.copy()}, nil
}

func (db *DB) getSpecimen(id string) (*occurrence, error) {
	o := &occurrence{}
	request := wsHead + "occurrence/" + id
	if err := db.simpleRequest(request, o); err != nil {
		return nil, err
	}
	return o, nil
}
