// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package gbif

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/js-arias/jdh/pkg/jdh"
)

type dsAnswer struct {
	Offset, Limit int64
	EndOfRecords  bool
	Results       []*dataset
}

type dataset struct {
	Key      string
	Title    string
	Citation citation
	Rights   string
	Homepage string
}

type citation struct {
	Text string
}

// returns a copy of dataset
func (ds *dataset) copy() *jdh.Dataset {
	if len(ds.Key) == 0 {
		return &jdh.Dataset{}
	}
	u, _ := url.Parse(ds.Homepage)
	set := &jdh.Dataset{
		Id:       strings.TrimSpace(ds.Key),
		Title:    strings.Join(strings.Fields(ds.Title), " "),
		Citation: strings.Join(strings.Fields(ds.Citation.Text), " "),
		License:  strings.TrimSpace(ds.Rights),
		Url:      u.String(),
	}
	return set
}

// listSet search for a list of datasets.
func (db *DB) listSet(kvs []jdh.KeyValue) (jdh.ListScanner, error) {
	l := &listScanner{
		c:   make(chan interface{}, 20),
		end: make(chan struct{}),
	}
	go func() {
		vals := url.Values{}
		title := ""
		license := ""
		for _, kv := range kvs {
			if len(kv.Value) == 0 {
				continue
			}
			switch kv.Key {
			case jdh.DataTitle:
				title = strings.Join(strings.Fields(kv.Value[0]), " ")
			case jdh.DataLicense:
				license = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
		}
		for off := int64(0); ; {
			request := wsHead + "dataset"
			if off > 0 {
				vals.Set("offset", strconv.FormatInt(off, 10))
				request += "?" + vals.Encode()
			}
			an := new(dsAnswer)
			if err := db.listRequest(request, an); err != nil {
				l.setErr(err)
				return
			}
			for _, ds := range an.Results {
				if (len(title) > 0) && (ds.Title != title) {
					continue
				}
				if (len(license) > 0) && (ds.Rights != license) {
					continue
				}
				select {
				case l.c <- ds.copy():
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

// GetSet returns a jdh scanner with a dataset.
func (db *DB) getSet(id string) (jdh.Scanner, error) {
	if len(id) == 0 {
		return nil, errors.New("dataset without identification")
	}
	ds, err := db.getDataset(id)
	if err != nil {
		return nil, err
	}
	return &getScanner{val: ds.copy()}, nil
}

func (db *DB) getDataset(id string) (*dataset, error) {
	ds := &dataset{}
	request := wsHead + "dataset/" + id
	if err := db.simpleRequest(request, ds); err != nil {
		return nil, err
	}
	return ds, nil
}
