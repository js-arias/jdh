// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package native

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

// Distros holds the distribution rasters of the taxons in the database.
type distros struct {
	db      *DB                  // parent database
	taxId   map[string]*rasTaxon // a map of id:taxon
	taxLs   *list.List           // the list of taxons
	ids     map[string]*raster   // a map of id:raster
	changed bool
	next    int64
}

// RasTaxon holds taxon information for the raster database.
type rasTaxon struct {
	id   string        // raster's id
	rsLs *list.List    // list of rasters
	elem *list.Element // element that contains the taxon
}

// Raster holds raster information.
type raster struct {
	data *jdh.Raster

	taxon *rasTaxon     // taxon of the raster
	elem  *list.Element // element that contains the raster
}

// distributions file
const distroFile = "distros"

// OpenDistros open raster distribution data.
func openDistros(db *DB) *distros {
	d := &distros{
		db:    db,
		taxId: make(map[string]*rasTaxon),
		taxLs: list.New(),
		ids:   make(map[string]*raster),
		next:  1,
	}
	p := filepath.Join(db.path, distroFile)
	f, err := os.Open(p)
	if err != nil {
		return d
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		ras := &jdh.Raster{}
		if err := dec.Decode(ras); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-distros: error: %v\n", err)
			break
		}
		d.setNext(ras.Id)
		if err := d.validate(ras); err != nil {
			log.Printf("db-distros: error: %v\n", err)
			continue
		}
		d.addRaster(ras)
	}
	return d
}

// SetNext sets the value of the next id.
func (d *distros) setNext(id string) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}
	if v >= d.next {
		d.next = v + 1
	}
}

// Validate validates that a raster is valid in the database, and set some
// canonical values. It returns an error if the raster is not valid.
func (d *distros) validate(ras *jdh.Raster) error {
	ras.Id = strings.TrimSpace(ras.Id)
	ras.Taxon = strings.TrimSpace(ras.Taxon)
	if (len(ras.Id) == 0) || (len(ras.Taxon) == 0) {
		return errors.New("raster without identification")
	}
	if _, ok := d.ids[ras.Id]; ok {
		return fmt.Errorf("raster id %s already in use", ras.Id)
	}
	if !d.db.t.isInDB(ras.Taxon) {
		return fmt.Errorf("taxon %s [associated with raster %s] not in dabtase", ras.Taxon, ras.Id)
	}
	if ras.Cols == 0 {
		return fmt.Errorf("raster %s with an invalid number of cols: 0", ras.Id)
	}
	if ras.Raster == nil {
		return fmt.Errorf("raster %s without rasterized data", ras.Id)
	}
	if ras.Source > jdh.MachineModel {
		ras.Source = jdh.UnknownRaster
	}
	ras.Reference = strings.TrimSpace(ras.Reference)
	ext := ras.Extern
	ras.Extern = nil
	for _, e := range ext {
		serv, id, err := jdh.ParseExtern(e)
		if err != nil {
			continue
		}
		if len(id) == 0 {
			continue
		}
		add := true
		for _, ex := range ras.Extern {
			if strings.HasPrefix(ex, serv) {
				add = false
				break
			}
		}
		if !add {
			continue
		}
		if _, ok := d.ids[e]; !ok {
			ras.Extern = append(ras.Extern, e)
		}
	}
	return nil
}

// AddRaster adds a new raster to the database.
func (d *distros) addRaster(ras *jdh.Raster) {
	rd := &raster{
		data: ras,
	}
	tax, ok := d.taxId[ras.Taxon]
	if !ok {
		tax = &rasTaxon{
			id:   ras.Taxon,
			rsLs: list.New(),
		}
		tax.elem = d.taxLs.PushBack(tax)
		d.taxId[tax.id] = tax
	}
	rd.taxon = tax
	rd.elem = tax.rsLs.PushBack(rd)
	d.ids[ras.Id] = rd
	for _, e := range ras.Extern {
		d.ids[e] = rd
	}
}

// Add adds a raster to the database.
func (d *distros) add(ras *jdh.Raster) (string, error) {
	id := strconv.FormatInt(d.next, 10)
	ras.Id = id
	if err := d.validate(ras); err != nil {
		return "", err
	}
	d.addRaster(ras)
	d.next++
	d.changed = true
	return id, nil
}

// Commit saves the rasters into hard disk.
func (d *distros) commit(e chan error) {
	if !d.changed {
		e <- nil
		return
	}
	p := filepath.Join(d.db.path, distroFile)
	f, err := os.Create(p)
	if err != nil {
		e <- err
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for et := d.taxLs.Front(); et != nil; et = et.Next() {
		tax := et.Value.(*rasTaxon)
		for e := tax.rsLs.Front(); e != nil; e = e.Next() {
			rd := e.Value.(*raster)
			enc.Encode(rd.data)
		}
	}
	d.changed = false
	e <- nil
}

// Delete deletes a raster or a set of rasters from the database.
func (d *distros) delete(vals []jdh.KeyValue) error {
	noVal := true
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			if len(kv.Value[0]) == 0 {
				return errors.New("raster without identification")
			}
			rd, ok := d.ids[kv.Value[0]]
			if !ok {
				return nil
			}
			tax := d.delRaster(rd)
			if tax.rsLs.Len() == 0 {
				tax.rsLs = nil
				d.taxLs.Remove(tax.elem)
				tax.elem = nil
				delete(d.taxId, tax.id)
			}
			d.changed = true
			noVal = false
			break
		}
		if kv.Key == jdh.RDisTaxon {
			if len(kv.Value[0]) == 0 {
				return errors.New("taxon without identification")
			}
			d.delTaxon(kv.Value[0])
			noVal = false
			break
		}
	}
	if noVal {
		return errors.New("raster-taxon without identification")
	}
	return nil
}

// DelTaxon removes all the rasters associated with a particular taxon.
func (d *distros) delTaxon(id string) {
	tax, ok := d.taxId[id]
	if !ok {
		return
	}
	for e := tax.rsLs.Front(); e != nil; e = tax.rsLs.Front() {
		rd := e.Value.(*raster)
		d.delRaster(rd)
	}
	tax.rsLs = nil
	d.taxLs.Remove(tax.elem)
	tax.elem = nil
	delete(d.taxId, tax.id)
	d.changed = true
}

// DelRaster removes a particular raster from the database. It returns the
// taxon that contained the raster.
func (d *distros) delRaster(rd *raster) *rasTaxon {
	for _, e := range rd.data.Extern {
		delete(d.ids, e)
	}
	delete(d.ids, rd.data.Id)
	tax := rd.taxon
	rd.taxon = nil
	rd.data = nil
	tax.rsLs.Remove(rd.elem)
	rd.elem = nil
	d.changed = true
	return tax
}

// Get returns a raster with a given id.
func (d *distros) get(id string) (*jdh.Raster, error) {
	if len(id) == 0 {
		return nil, errors.New("raster without identification")
	}
	rd, ok := d.ids[id]
	if !ok {
		return nil, nil
	}
	return rd.data, nil
}

// List returns a list of rasters.
func (d *distros) list(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	noVal := true
	// creates the list
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.RDisTaxon {
			if len(kv.Value[0]) == 0 {
				return nil, errors.New("taxon without identification")
			}
			tax, ok := d.taxId[kv.Value[0]]
			if !ok {
				return l, nil
			}
			for e := tax.rsLs.Front(); e != nil; e = e.Next() {
				rd := e.Value.(*raster)
				l.PushBack(rd.data)
			}
			noVal = false
			break
		}
		if kv.Key == jdh.RDisTaxonParent {
			if len(kv.Value[0]) == 0 {
				return nil, errors.New("taxon without identification")
			}
			pId := kv.Value[0]
			if !d.db.t.isInDB(pId) {
				return l, nil
			}
			for m := d.taxLs.Front(); m != nil; m = m.Next() {
				tax := m.Value.(*rasTaxon)
				if tax.id != pId {
					if !d.db.t.isDesc(tax.id, pId) {
						continue
					}
				}
				for e := tax.rsLs.Front(); e != nil; e = e.Next() {
					rd := e.Value.(*raster)
					l.PushBack(rd.data)
				}
			}
			noVal = false
			break
		}
	}
	if noVal {
		return nil, errors.New("taxon without identification")
	}

	// filters the list
	for _, kv := range vals {
		if l.Len() == 0 {
			break
		}
		if len(kv.Value) == 0 {
			continue
		}
		switch kv.Key {
		case jdh.RDisCols:
			val, err := strconv.ParseUint(kv.Value[0], 10, 0)
			if (err != nil) || (val == 0) {
				break
			}
			v := uint(val)
			for e := l.Front(); e != nil; {
				nx := e.Next()
				ras := e.Value.(*jdh.Raster)
				if ras.Cols != v {
					l.Remove(e)
				}
				e = nx
			}
		case jdh.RDisSource:
			val := strings.ToLower(strings.TrimSpace(kv.Value[0]))
			v := jdh.GetRasterSource(val)
			if v.String() != val {
				break
			}
			for e := l.Front(); e != nil; {
				nx := e.Next()
				ras := e.Value.(*jdh.Raster)
				if ras.Source != v {
					l.Remove(e)
				}
				e = nx
			}
		}
	}
	return l, nil
}

// Set sets a value of a raster in the database.
func (d *distros) set(vals []jdh.KeyValue) error {
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
		return errors.New("raster without identification")
	}
	rd, ok := d.ids[id]
	if !ok {
		return nil
	}
	ras := rd.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.RasPixel:
			if len(kv.Value) == 0 {
				continue
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				continue
			}
			coor := strings.Split(v, ",")
			if len(coor) != 3 {
				return errors.New("invalid raster pixel reference: " + v)
			}
			x64, err := strconv.ParseInt(coor[0], 10, 0)
			if err != nil {
				return err
			}
			y64, err := strconv.ParseInt(coor[1], 10, 0)
			if err != nil {
				return err
			}
			p64, err := strconv.ParseInt(coor[2], 10, 0)
			if err != nil {
				return err
			}
			pt, px := image.Pt(int(x64), int(y64)), int(p64)
			if ras.Raster.At(pt) == px {
				continue
			}
			ras.Raster.Set(pt, px)
		case jdh.RDisSource:
			v := jdh.UnknownRaster
			if len(kv.Value) > 0 {
				v = jdh.GetRasterSource(strings.TrimSpace(kv.Value[0]))
			}
			if ras.Source == v {
				continue
			}
		case jdh.RDisTaxon:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if len(v) == 0 {
				continue
			}
			if ras.Taxon == v {
				continue
			}
			tax, ok := d.taxId[v]
			if !ok {
				if !d.db.t.isInDB(v) {
					continue
				}
				tax = &rasTaxon{
					id:   v,
					rsLs: list.New(),
				}
				tax.elem = d.taxLs.PushBack(tax)
				d.taxId[tax.id] = tax
			}
			oldtax := rd.taxon
			oldtax.rsLs.Remove(rd.elem)
			rd.elem = tax.rsLs.PushBack(rd)
			rd.taxon = tax
			rd.data.Taxon = tax.id
			if oldtax.rsLs.Len() == 0 {
				oldtax.rsLs = nil
				d.taxLs.Remove(oldtax.elem)
				oldtax.elem = nil
				delete(d.taxId, oldtax.id)
			}
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if ras.Comment == v {
				continue
			}
			ras.Comment = v
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
					if !d.delExtern(rd, serv) {
						continue
					}
					ok = true
					continue
				}
				if d.addExtern(rd, v) != nil {
					continue
				}
				ok = true
			}
			if !ok {
				continue
			}
		case jdh.KeyReference:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if ras.Reference == v {
				continue
			}
			ras.Reference = v
		default:
			continue
		}
		d.changed = true
	}
	return nil
}

// AddExtern adds an extern id to an specimen.
func (d *distros) addExtern(rd *raster, extern string) error {
	serv, id, err := jdh.ParseExtern(extern)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return nil
	}
	if or, ok := d.ids[extern]; ok {
		return fmt.Errorf("extern id %s of %s alredy in use by %s", extern, rd.data.Id, or.data.Id)
	}
	// the service is already assigned, then overwrite
	for i, e := range rd.data.Extern {
		if strings.HasPrefix(e, serv) {
			delete(d.ids, e)
			rd.data.Extern[i] = extern
			d.ids[extern] = rd
			return nil
		}
	}
	rd.data.Extern = append(rd.data.Extern, extern)
	d.ids[extern] = rd
	return nil
}

// DelExtern deletes an extern id of an specimen.
func (d *distros) delExtern(rd *raster, service string) bool {
	for i, e := range rd.data.Extern {
		if strings.HasPrefix(e, service) {
			delete(d.ids, e)
			copy(rd.data.Extern[i:], rd.data.Extern[i+1:])
			rd.data.Extern[len(rd.data.Extern)-1] = ""
			rd.data.Extern = rd.data.Extern[:len(rd.data.Extern)-1]
			return true
		}
	}
	return false
}
