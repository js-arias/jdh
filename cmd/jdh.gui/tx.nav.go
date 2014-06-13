// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"strings"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/sparta"
	"github.com/js-arias/sparta/widget"
)

var txNav = &cmdapp.Command{
	Name:     "tx.nav",
	Synopsis: `[-e|--extdb name] [-p|--port value]`,
	Short:    "displays taxonomy",
	Long: `
Description

Tx.nav displays the taxonomy stored in the database.

Options

    -e name
    --extdb name
      Set the extern database. By default, the local database is used.
      Valid values are:
          gbif    taxonomy from gbif.
          inat    taxonomy from inaturalist.
          ncbi    taxonomy from ncbi (genbank).

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	txNav.Flag.StringVar(&extDBFlag, "extdb", "", "")
	txNav.Flag.StringVar(&extDBFlag, "e", "", "")
	txNav.Flag.StringVar(&portFlag, "port", "", "")
	txNav.Flag.StringVar(&portFlag, "p", "", "")
	txNav.Run = txNavRun
}

func txNavRun(c *cmdapp.Command, args []string) {
	cmd = c
	var db jdh.DB
	if len(extDBFlag) != 0 {
		openExt(c, extDBFlag, "")
		db = extDB
	} else {
		openLocal(c)
		db = localDB
	}

	title := fmt.Sprintf("%s: please wait", c.Name)
	m := widget.NewMainWindow("main", title)
	geo := m.Property(sparta.Geometry).(image.Rectangle)
	widget.NewButton(m, "upTax", "up", image.Rect(5, 5, 5+(sparta.WidthUnit*10), 5+(3*sparta.HeightUnit/2)))
	tx := widget.NewCanvas(m, "info", image.Rect(210, 10+(3*sparta.HeightUnit/2), geo.Dx()-10, geo.Dy()-10))
	tx.SetProperty(sparta.Border, true)
	tx.Capture(sparta.Expose, txNavInfoExpose)
	wnd["info"] = tx
	l := widget.NewList(m, "taxonList", image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, geo.Dy()-10))
	wnd["taxonList"] = l

	m.Capture(sparta.Configure, txNavConf)
	m.Capture(sparta.Command, txNavComm)
	sparta.Block(nil)
	go txNavInitList(m, l, db, nil, 0)

	sparta.Run()
}

func txNavConf(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.ConfigureEvent)
	l := wnd["taxonList"]
	l.SetProperty(sparta.Geometry, image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, ev.Rect.Dy()-10))
	tx := wnd["info"]
	tx.SetProperty(sparta.Geometry, image.Rect(210, 10+(3*sparta.HeightUnit/2), ev.Rect.Dx()-10, ev.Rect.Dy()-10))
	return false
}

func txNavComm(m sparta.Widget, e interface{}) bool {
	d := m.Property(sparta.Data)
	if d == nil {
		return true
	}
	data := d.(*txList)
	ev := e.(sparta.CommandEvent)
	switch ev.Source.Property(sparta.Name).(string) {
	case "taxonList":
		if ev.Value < 0 {
			i := -ev.Value - 1
			if i >= len(data.desc) {
				break
			}
			title := fmt.Sprintf("%s: please wait", cmd.Name)
			m.SetProperty(sparta.Caption, title)
			ev.Source.SetProperty(widget.ListList, nil)
			tx := wnd["info"]
			tx.SetProperty(sparta.Data, nil)
			tx.Update()
			sparta.Block(nil)
			go txNavInitList(m, ev.Source, data.db, data, i)
			break
		}
		if data.IsSel(ev.Value) {
			data.sels = nil
		} else {
			data.sels = []int{ev.Value}
		}
		tx := wnd["info"]
		tx.SetProperty(sparta.Data, nil)
		tx.Update()
		ev.Source.Update()
		sparta.Block(nil)
		go func() {
			txNavInfo(tx, data)
			sparta.Unblock(nil)
		}()
	case "upTax":
		if data.tax.Id == "0" {
			break
		}
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		m.SetProperty(sparta.Caption, title)
		l := wnd["taxonList"]
		l.SetProperty(widget.ListList, nil)
		tx := wnd["info"]
		tx.SetProperty(sparta.Data, nil)
		sparta.Block(nil)
		go txNavAncList(m, l, data.db, data.tax)
	}
	return true
}

func txNavInfoExpose(tx sparta.Widget, e interface{}) bool {
	d := tx.Property(sparta.Data)
	if d == nil {
		return false
	}
	data := d.(*txTaxAnc)
	c := tx.(*widget.Canvas)
	txt := widget.Text{}
	txt.Pos.X = 2
	txt.Pos.Y = 2
	txt.Text = "Id: " + data.tax.Id
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	txt.Text = data.tax.Name
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	txt.Text = data.tax.Authority
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	txt.Text = data.tax.Rank.String()
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	if data.tax.IsValid {
		txt.Text = "Valid"
		c.Draw(txt)
		if data.anc != nil {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "Parent: " + data.anc.Name
			c.Draw(txt)

		}
	} else {
		txt.Text = "Synonym of " + data.anc.Name
		c.Draw(txt)

	}
	if len(data.tax.Extern) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Extern ids:"
		c.Draw(txt)
		for _, e := range data.tax.Extern {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "    " + e
			c.Draw(txt)
		}
	}
	if len(data.tax.Comment) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Comments:"
		c.Draw(txt)
		cmt := strings.Split(data.tax.Comment, "\n")
		for _, e := range cmt {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "  " + e
			c.Draw(txt)
		}
	}
	return false
}

func txNavInitList(m, l sparta.Widget, db jdh.DB, data *txList, i int) {
	if data == nil {
		data = newTxList(nil, db, true)
	} else {
		d := newTxList(data.desc[i], data.db, true)
		if len(d.desc) == 0 {
			data.sels = []int{i}
		} else {
			data = d
		}
	}
	title := fmt.Sprintf("%s: %s [id: %s]", cmd.Name, data.tax.Name, data.tax.Id)
	m.SetProperty(sparta.Caption, title)
	m.SetProperty(sparta.Data, data)
	l.SetProperty(widget.ListList, data)
	tx := wnd["info"]
	txNavInfo(tx, data)
	sparta.Unblock(nil)
}

func txNavAncList(m, l sparta.Widget, db jdh.DB, tax *jdh.Taxon) {
	var p *jdh.Taxon
	if len(tax.Parent) > 0 {
		p = taxon(cmd, db, tax.Parent)
		if len(p.Id) == 0 {
			p = nil
		}
	}
	data := newTxList(p, db, true)
	title := fmt.Sprintf("%s: %s [id: %s]", cmd.Name, data.tax.Name, data.tax.Id)
	m.SetProperty(sparta.Caption, title)
	m.SetProperty(sparta.Data, data)
	for i, d := range data.desc {
		if d.Id == tax.Id {
			data.sels = []int{i}
			break
		}
	}
	l.SetProperty(widget.ListList, data)
	tx := wnd["info"]
	txNavInfo(tx, data)
	sparta.Unblock(nil)
}

func txNavInfo(tx sparta.Widget, data *txList) {
	if len(data.sels) == 0 {
		tx.SetProperty(sparta.Data, nil)
	} else {
		pair := &txTaxAnc{
			tax: data.desc[data.sels[0]],
			anc: data.tax,
		}
		if data.tax.Id == "0" {
			pair.anc = nil
		}
		tx.SetProperty(sparta.Data, pair)
	}
	tx.Update()
}
