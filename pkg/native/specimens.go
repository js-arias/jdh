// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package native

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/jdh/pkg/geography"
	"github.com/js-arias/jdh/pkg/jdh"
)

// Specimens holds the specimens of the database.
type specimens struct {
	db      *DB                  // parent database
	taxId   map[string]*speTaxon // a map of id:taxon
	taxLs   *list.List           // the list of taxons
	ids     map[string]*specimen // a map of id:specimen
	changed bool                 // if true, the database has changed
	next    int64                // next valid id
}

// SpeTaxon holds taxon information for the specimen database.
type speTaxon struct {
	id    string        // taxon's id
	specs *list.List    // list of specimens
	elem  *list.Element // element that contains the taxon
}

// Specimen holds specimen information.
type specimen struct {
	data *jdh.Specimen

	taxon *speTaxon     // taxon that contains the specimen
	elem  *list.Element // element that contains the specimen
}

// specimens file
const speFile = "specimens"

// OpenSpecimens open specimen data.
func openSpecimens(db *DB) *specimens {
	s := &specimens{
		db:    db,
		taxId: make(map[string]*speTaxon),
		taxLs: list.New(),
		ids:   make(map[string]*specimen),
		next:  1,
	}
	p := filepath.Join(db.path, speFile)
	f, err := os.Open(p)
	if err != nil {
		return s
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		spe := &jdh.Specimen{}
		if err := dec.Decode(spe); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-specimens: error: %v\n", err)
			break
		}
		s.setNext(spe.Id)
		if err := s.validate(spe); err != nil {
			log.Printf("db-specimens: error: %v\n", err)
			continue
		}
		s.addSpecimen(spe)
	}
	return s
}

// SetNext sets the value of the next id.
func (s *specimens) setNext(id string) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}
	if v >= s.next {
		s.next = v + 1
	}
}

// Validate validates that an specimen is valid in the database, and set
// some canonical values. It returns an error if the specimen is not valid.
func (s *specimens) validate(spe *jdh.Specimen) error {
	spe.Id = strings.TrimSpace(spe.Id)
	spe.Taxon = strings.TrimSpace(spe.Taxon)
	if (len(spe.Id) == 0) || (len(spe.Taxon) == 0) {
		return errors.New("specimen without identification")
	}
	if _, ok := s.ids[spe.Id]; ok {
		return fmt.Errorf("specimen id %s already in use", spe.Id)
	}
	spe.Catalog = strings.TrimSpace(spe.Catalog)
	if len(spe.Catalog) > 0 {
		if _, ok := s.ids[spe.Catalog]; ok {
			return fmt.Errorf("specimen catalog code %s already in use", spe.Catalog)
		}
	}
	if !s.db.t.isInDB(spe.Taxon) {
		return fmt.Errorf("taxon %s [associated with specimen %s] not in database", spe.Taxon, spe.Id)
	}
	if !spe.Geography.IsValid() {
		spe.Geography = geography.Location{}
	}
	if !spe.Georef.IsValid() {
		spe.Georef = geography.InvalidGeoref()
	} else if (spe.Georef.Point.Lon == 0) || (spe.Georef.Point.Lon == 1) || (spe.Georef.Point.Lat == 0) || (spe.Georef.Point.Lat == 1) {
		spe.Georef = geography.InvalidGeoref()
	}

	if spe.Basis > jdh.Remote {
		spe.Basis = jdh.UnknownBasis
	}
	spe.Reference = strings.TrimSpace(spe.Reference)
	spe.Determiner = strings.Join(strings.Fields(spe.Determiner), " ")
	spe.Collector = strings.Join(strings.Fields(spe.Collector), " ")
	spe.Dataset = strings.TrimSpace(spe.Dataset)
	if len(spe.Dataset) > 0 {
		if !s.db.d.isInDB(spe.Dataset) {
			spe.Dataset = ""
		}
	}
	ext := spe.Extern
	spe.Extern = nil
	for _, e := range ext {
		serv, id, err := jdh.ParseExtern(e)
		if err != nil {
			continue
		}
		if len(id) == 0 {
			continue
		}
		add := true
		for _, ex := range spe.Extern {
			if strings.HasPrefix(ex, serv) {
				add = false
				break
			}
		}
		if !add {
			continue
		}
		if _, ok := s.ids[e]; !ok {
			spe.Extern = append(spe.Extern, e)
		}
	}
	return nil
}

// AddSpecimen adds a new specimen to the database.
func (s *specimens) addSpecimen(spe *jdh.Specimen) {
	sp := &specimen{
		data: spe,
	}
	tax, ok := s.taxId[spe.Taxon]
	if !ok {
		tax = &speTaxon{
			id:    spe.Taxon,
			specs: list.New(),
		}
		tax.elem = s.taxLs.PushBack(tax)
		s.taxId[tax.id] = tax
	}
	sp.taxon = tax
	sp.elem = tax.specs.PushBack(sp)
	s.ids[spe.Id] = sp
	for _, e := range spe.Extern {
		s.ids[e] = sp
	}
}

// Add adds an specimen to the database.
func (s *specimens) add(spe *jdh.Specimen) (string, error) {
	id := strconv.FormatInt(s.next, 10)
	spe.Id = id
	if err := s.validate(spe); err != nil {
		return "", err
	}
	s.addSpecimen(spe)
	s.next++
	s.changed = true
	return id, nil
}

// Commit saves the occurrences into hard disk.
func (s *specimens) commit(e chan error) {
	if !s.changed {
		e <- nil
		return
	}
	p := filepath.Join(s.db.path, speFile)
	f, err := os.Create(p)
	if err != nil {
		e <- err
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for et := s.taxLs.Front(); et != nil; et = et.Next() {
		tax := et.Value.(*speTaxon)
		for e := tax.specs.Front(); e != nil; e = e.Next() {
			sp := e.Value.(*specimen)
			enc.Encode(sp.data)
		}
	}
	s.changed = false
	e <- nil
}

// Delete deletes an specimen or a group of specimens from the database.
func (s *specimens) delete(vals []jdh.KeyValue) error {
	noVal := true
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			if len(kv.Value[0]) == 0 {
				return errors.New("specimen without identification")
			}
			sp, ok := s.ids[kv.Value[0]]
			if !ok {
				return nil
			}
			tax := s.delSpecimen(sp)
			if tax.specs.Len() == 0 {
				tax.specs = nil
				s.taxLs.Remove(tax.elem)
				tax.elem = nil
				delete(s.taxId, tax.id)
			}
			s.changed = true
			noVal = false
			break
		}
		if kv.Key == jdh.SpeTaxon {
			if len(kv.Value[0]) == 0 {
				return errors.New("taxon without identification")
			}
			s.delTaxon(kv.Value[0])
			noVal = false
			break
		}
	}
	if noVal {
		return errors.New("specimen-taxon without identification")
	}
	return nil
}

// DelTaxon removes all the specimens associated with a particular taxon.
func (s *specimens) delTaxon(id string) {
	tax, ok := s.taxId[id]
	if !ok {
		return
	}
	for e := tax.specs.Front(); e != nil; e = tax.specs.Front() {
		sp := e.Value.(*specimen)
		s.delSpecimen(sp)
	}
	tax.specs = nil
	s.taxLs.Remove(tax.elem)
	tax.elem = nil
	delete(s.taxId, tax.id)
	s.changed = true
}

// DelSpecimen removes a particular specimen from the database. It returns the
// taxon that contained the specimen.
func (s *specimens) delSpecimen(sp *specimen) *speTaxon {
	for _, e := range sp.data.Extern {
		delete(s.ids, e)
	}
	if len(sp.data.Catalog) > 0 {
		delete(s.ids, sp.data.Catalog)
	}
	delete(s.ids, sp.data.Id)
	tax := sp.taxon
	sp.taxon = nil
	sp.data = nil
	tax.specs.Remove(sp.elem)
	sp.elem = nil
	s.changed = true
	return tax
}

// Get returns an specimen with a given id.
func (s *specimens) get(id string) (*jdh.Specimen, error) {
	if len(id) == 0 {
		return nil, errors.New("specimen without identification")
	}
	sp, ok := s.ids[id]
	if !ok {
		return nil, nil
	}
	return sp.data, nil
}

// List returns a list of specimens.
func (s *specimens) list(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	noVal := true
	// creates the list
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.SpeTaxon {
			if len(kv.Value[0]) == 0 {
				return nil, errors.New("taxon without identification")
			}
			tax, ok := s.taxId[kv.Value[0]]
			if !ok {
				return l, nil
			}
			for e := tax.specs.Front(); e != nil; e = e.Next() {
				sp := e.Value.(*specimen)
				l.PushBack(sp.data)
			}
			noVal = false
			break
		}
		if kv.Key == jdh.SpeTaxonParent {
			if len(kv.Value[0]) == 0 {
				return nil, errors.New("taxon without identification")
			}
			pId := kv.Value[0]
			if !s.db.t.isInDB(pId) {
				return l, nil
			}
			for m := s.taxLs.Front(); m != nil; m = m.Next() {
				tax := m.Value.(*speTaxon)
				if tax.id != pId {
					if !s.db.t.isDesc(tax.id, pId) {
						continue
					}
				}
				for e := tax.specs.Front(); e != nil; e = e.Next() {
					sp := e.Value.(*specimen)
					l.PushBack(sp.data)
				}
			}
			noVal = false
			break
		}
	}
	if noVal {
		return nil, errors.New("taxon without identification")
	}

	//filters the list.
	for _, kv := range vals {
		if l.Len() == 0 {
			break
		}
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.GeoCountry:
			for e := l.Front(); e != nil; {
				nx := e.Next()
				spe := e.Value.(*jdh.Specimen)
				remove := true
				for _, v := range kv.Value {
					c := geography.GetCountry(v)
					if len(c) == 0 {
						continue
					}
					if spe.Geography.Country == c {
						remove = false
						break
					}
				}
				if remove {
					l.Remove(e)
				}
				e = nx
			}
		case jdh.SpeGeoref:
			if (kv.Value[0] != "true") && (kv.Value[0] != "false") {
				continue
			}
			ok := false
			if kv.Value[0] == "true" {
				ok = true
			}
			for e := l.Front(); e != nil; {
				nx := e.Next()
				spe := e.Value.(*jdh.Specimen)
				if spe.Georef.IsValid() != ok {
					l.Remove(e)
				}
				e = nx
			}
		}
	}
	return l, nil
}

// Set sets a value of an specimen in the database.
func (s *specimens) set(vals []jdh.KeyValue) error {
	id := ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			id = kv.Value[0]
			break
		}
	}
	if len(id) == 0 {
		return errors.New("specimen without identification")
	}
	sp, ok := s.ids[id]
	if !ok {
		return nil
	}
	spe := sp.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.SpeBasis:
			v := jdh.UnknownBasis
			if len(kv.Value) > 0 {
				v = jdh.GetBasisOfRecord(strings.TrimSpace(kv.Value[0]))
			}
			if spe.Basis == v {
				continue
			}
			spe.Basis = v
		case jdh.SpeCatalog:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if spe.Catalog == v {
				continue
			}
			if len(v) > 0 {
				if _, ok := s.ids[v]; ok {
					return fmt.Errorf("specimen catalog code %s already in use", kv.Value)
				}
			}
			if len(spe.Catalog) > 0 {
				delete(s.ids, spe.Catalog)
			}
			spe.Catalog = v
			if len(v) > 0 {
				s.ids[v] = sp
			}
		case jdh.SpeCollector:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Collector == v {
				continue
			}
			spe.Collector = v
		case jdh.SpeDataset:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if spe.Dataset == v {
				continue
			}
			if len(v) > 0 {
				if !s.db.d.isInDB(v) {
					continue
				}
			}
			spe.Dataset = v
		case jdh.SpeDate:
			if len(kv.Value) > 0 {
				v := strings.TrimSpace(kv.Value[0])
				t, err := time.Parse(jdh.Iso8601, v)
				if err != nil {
					return err
				}
				if spe.Date.Equal(t) {
					continue
				}
				spe.Date = t
				break
			}
			if spe.Date.IsZero() {
				continue
			}
			spe.Date = time.Time{}
		case jdh.SpeDeterminer:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Determiner == v {
				continue
			}
			spe.Determiner = v
		case jdh.SpeLocality:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Locality == v {
				continue
			}
			spe.Locality = v
		case jdh.SpeReference:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if spe.Reference == v {
				continue
			}
			spe.Reference = v
		case jdh.SpeTaxon:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if len(v) == 0 {
				continue
			}
			if spe.Taxon == v {
				continue
			}
			tax, ok := s.taxId[v]
			if !ok {
				if !s.db.t.isInDB(v) {
					continue
				}
				tax = &speTaxon{
					id:    v,
					specs: list.New(),
				}
				tax.elem = s.taxLs.PushBack(tax)
				s.taxId[tax.id] = tax
			}
			oldtax := sp.taxon
			oldtax.specs.Remove(sp.elem)
			sp.elem = tax.specs.PushBack(sp)
			sp.taxon = tax
			sp.data.Taxon = tax.id
			if oldtax.specs.Len() == 0 {
				oldtax.specs = nil
				s.taxLs.Remove(oldtax.elem)
				oldtax.elem = nil
				delete(s.taxId, oldtax.id)
			}
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if spe.Comment == v {
				continue
			}
			spe.Comment = v
		case jdh.KeyExtern:
			ok := false
			for _, v := range kv.Value {
				v = strings.TrimSpace(v)
				if len(v) == 0 {
					continue
				}
				serv, ext, err := jdh.ParseExtern(v)
				if err != nil {
					return err
				}
				if len(ext) == 0 {
					if !s.delExtern(sp, serv) {
						continue
					}
					ok = true
					continue
				}
				if s.addExtern(sp, v) != nil {
					continue
				}
				ok = true
			}
			if !ok {
				continue
			}
		case jdh.GeoCountry:
			v := geography.Country("")
			if len(kv.Value) > 0 {
				v = geography.GetCountry(strings.Join(strings.Fields(kv.Value[0]), " "))
			}
			if spe.Geography.Country == v {
				continue
			}
			spe.Geography.Country = v
		case jdh.GeoCounty:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Geography.County == v {
				continue
			}
			spe.Geography.County = v
		case jdh.GeoLonLat:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if len(v) == 0 {
				if !spe.Georef.IsValid() {
					continue
				}
				spe.Georef = geography.InvalidGeoref()
				break
			}
			coor := strings.Split(v, ",")
			if len(coor) != 2 {
				return errors.New("invalid geographic coordinate values")
			}
			lon, err := strconv.ParseFloat(coor[0], 64)
			if err != nil {
				return err
			}
			lat, err := strconv.ParseFloat(coor[1], 64)
			if err != nil {
				return err
			}
			if (lon == 0) || (lon == 1) || (lat == 0) || (lat == 1) {
				return errors.New("invalid geographic coordinate values")
			}
			if (!geography.IsLon(lon)) || (!geography.IsLat(lat)) {
				return errors.New("invalid geographic coordinate values")
			}
			spe.Georef.Point = geography.Point{Lon: lon, Lat: lat}
		case jdh.GeoSource:
			if !spe.Georef.IsValid() {
				continue
			}
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Georef.Source == v {
				continue
			}
			spe.Georef.Source = v
		case jdh.GeoState:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Geography.State == v {
				continue
			}
			spe.Geography.State = v
		case jdh.GeoUncertainty:
			if !spe.Georef.IsValid() {
				continue
			}
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			un64, err := strconv.ParseUint(v, 10, 0)
			if err != nil {
				return err
			}
			un := uint(un64)
			if un == spe.Georef.Uncertainty {
				continue
			}
			spe.Georef.Uncertainty = un
		case jdh.GeoValidation:
			if !spe.Georef.IsValid() {
				continue
			}
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if spe.Georef.Validation == v {
				continue
			}
			spe.Georef.Validation = v
		default:
			continue
		}
		s.changed = true
	}
	return nil
}

// AddExtern adds an extern id to an specimen.
func (s *specimens) addExtern(sp *specimen, extern string) error {
	serv, id, err := jdh.ParseExtern(extern)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return nil
	}
	if or, ok := s.ids[extern]; ok {
		return fmt.Errorf("extern id %s of %s alredy in use by %s", extern, sp.data.Id, or.data.Id)
	}
	// the service is already assigned, then overwrite
	for i, e := range sp.data.Extern {
		if strings.HasPrefix(e, serv) {
			delete(s.ids, e)
			sp.data.Extern[i] = extern
			s.ids[extern] = sp
			return nil
		}
	}
	sp.data.Extern = append(sp.data.Extern, extern)
	s.ids[extern] = sp
	return nil
}

// DelExtern deletes an extern id of an specimen.
func (s *specimens) delExtern(sp *specimen, service string) bool {
	for i, e := range sp.data.Extern {
		if strings.HasPrefix(e, service) {
			delete(s.ids, e)
			copy(sp.data.Extern[i:], sp.data.Extern[i+1:])
			sp.data.Extern[len(sp.data.Extern)-1] = ""
			sp.data.Extern = sp.data.Extern[:len(sp.data.Extern)-1]
			return true
		}
	}
	return false
}
