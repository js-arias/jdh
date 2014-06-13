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
	_ "github.com/js-arias/sparta/init"
	"github.com/js-arias/sparta/widget"
)

var spNav = &cmdapp.Command{
	Name:     "sp.nav",
	Synopsis: `[-p|--port value]`,
	Short:    "edits taxonomy",
	Long: `
Description

Sp.nav displays the specimens stored in the database.

Options

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	spNav.Flag.StringVar(&portFlag, "port", "", "")
	spNav.Flag.StringVar(&portFlag, "p", "", "")
	spNav.Run = spNavRun
}

func spNavRun(c *cmdapp.Command, args []string) {
	cmd = c
	var db jdh.DB
	openLocal(c)
	db = localDB

	title := fmt.Sprintf("%s: please wait", c.Name)
	m := widget.NewMainWindow("main", title)
	geo := m.Property(sparta.Geometry).(image.Rectangle)
	widget.NewButton(m, "upTax", "up", image.Rect(5, 5, 5+(sparta.WidthUnit*10), 5+(3*sparta.HeightUnit/2)))

	dy := geo.Dy() - 10 - (10 + (3 * sparta.HeightUnit / 2))

	l := widget.NewList(m, "taxonList", image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, (dy/2)+10+(3*sparta.HeightUnit/2)))
	wnd["taxonList"] = l

	s := widget.NewList(m, "speList", image.Rect(210, 10+(3*sparta.HeightUnit/2), 410, (dy/2)+10+(3*sparta.HeightUnit/2)))
	wnd["speList"] = s

	tx := widget.NewCanvas(m, "info", image.Rect(5, (dy/2)+20+(3*sparta.HeightUnit/2), geo.Dx()-10, geo.Dy()-10))
	tx.SetProperty(sparta.Border, true)
	tx.Capture(sparta.Expose, spNavInfoExpose)
	wnd["info"] = tx

	m.Capture(sparta.Configure, spNavConf)
	m.Capture(sparta.Command, spNavComm)
	sparta.Block(nil)

	go func() {
		spNavInitTaxList(m, l, db, nil, 0)
		sparta.Unblock(nil)
	}()

	sparta.Run()
}

func spNavConf(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.ConfigureEvent)
	dy := ev.Rect.Dy() - 10 - (10 + (3 * sparta.HeightUnit / 2))
	l := wnd["taxonList"]
	l.SetProperty(sparta.Geometry, image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, (dy/2)+10+(3*sparta.HeightUnit/2)))
	l = wnd["speList"]
	l.SetProperty(sparta.Geometry, image.Rect(210, 10+(3*sparta.HeightUnit/2), 410, (dy/2)+10+(3*sparta.HeightUnit/2)))
	tx := wnd["info"]
	tx.SetProperty(sparta.Geometry, image.Rect(5, (dy/2)+20+(3*sparta.HeightUnit/2), ev.Rect.Dx()-10, ev.Rect.Dy()-10))
	return false
}

func spNavComm(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.CommandEvent)
	switch ev.Source.Property(sparta.Name).(string) {
	case "speList":
		d := ev.Source.Property(widget.ListList)
		if d == nil {
			break
		}
		data := d.(*spList)
		i := ev.Value
		if ev.Value < 0 {
			i = -ev.Value - 1
		}
		tx := wnd["info"]
		if data.IsSel(i) {
			data.sel = -1
			tx.SetProperty(sparta.Data, nil)
			tx.Update()
			ev.Source.Update()
			break
		}
		data.sel = i
		tx.SetProperty(sparta.Data, nil)
		tx.Update()
		ev.Source.Update()
		sparta.Block(nil)
		go func() {
			spNavInfo(tx, data.db, data.spe[i], data.tax)
			sparta.Unblock(nil)
		}()
	case "taxonList":
		d := ev.Source.Property(widget.ListList)
		if d == nil {
			break
		}
		data := d.(*txList)
		s := wnd["speList"]
		tx := wnd["info"]
		if ev.Value < 0 {
			i := -ev.Value - 1
			if i >= len(data.desc) {
				break
			}
			title := fmt.Sprintf("%s: please wait", cmd.Name)
			m.SetProperty(sparta.Caption, title)
			ev.Source.SetProperty(widget.ListList, nil)
			s.SetProperty(widget.ListList, nil)
			tx.SetProperty(sparta.Data, nil)
			tx.Update()
			sparta.Block(nil)
			go func() {
				spNavInitTaxList(m, ev.Source, data.db, data, i)
				sparta.Unblock(nil)
			}()
			break
		}
		if data.IsSel(ev.Value) {
			data.sels = nil
			tx.SetProperty(sparta.Data, nil)
			tx.Update()
			s.SetProperty(widget.ListList, nil)
		} else {
			data.sels = []int{ev.Value}
			tx.SetProperty(sparta.Data, nil)
			tx.Update()
			sparta.Block(nil)
			go func() {
				spNavInitSpeList(m, s)
				sparta.Unblock(nil)
			}()
		}
		ev.Source.Update()
	case "upTax":
		l := wnd["taxonList"]
		s := wnd["speList"]
		tx := wnd["info"]
		d := l.Property(widget.ListList)
		if d == nil {
			break
		}
		data := d.(*txList)
		if data.tax.Id == "0" {
			break
		}
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		m.SetProperty(sparta.Caption, title)
		l.SetProperty(widget.ListList, nil)
		s.SetProperty(widget.ListList, nil)
		tx.SetProperty(sparta.Data, nil)
		tx.Update()
		sparta.Block(nil)
		go func() {
			spNavAncList(m, l, data.db, data.tax)
			spNavInitSpeList(m, s)
			sparta.Unblock(nil)
		}()
	}
	return true
}

func spNavInfoExpose(tx sparta.Widget, e interface{}) bool {
	d := tx.Property(sparta.Data)
	if d == nil {
		return false
	}
	data := d.(*spInfo)
	c := tx.(*widget.Canvas)
	txt := widget.Text{}
	txt.Pos.X = 2
	txt.Pos.Y = 2
	txt.Text = "Id: " + data.spe.Id
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	txt.Text = "Taxon: " + data.tax.Name
	c.Draw(txt)
	txt.Pos.Y += sparta.HeightUnit
	txt.Text = "Basis: " + data.spe.Basis.String()
	c.Draw(txt)
	if len(data.spe.Reference) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Ref: " + data.spe.Reference
		c.Draw(txt)
	}
	if data.set != nil {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Dataset: " + data.set.Title
		c.Draw(txt)
	}
	if len(data.spe.Catalog) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Catalog: " + data.spe.Catalog
		c.Draw(txt)
	}
	if len(data.spe.Determiner) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Determiner: " + data.spe.Determiner
		c.Draw(txt)
	}
	if len(data.spe.Collector) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Taxon: " + data.spe.Collector
		c.Draw(txt)
	}
	if !data.spe.Date.IsZero() {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Date:: " + data.spe.Date.Format(jdh.Iso8601)
		c.Draw(txt)
	}
	if len(data.spe.Geography.Country) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Country: " + data.spe.Geography.Country.Name()
		c.Draw(txt)
	}
	if len(data.spe.Geography.State) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "State: " + data.spe.Geography.State
		c.Draw(txt)
	}
	if len(data.spe.Locality) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Locality: " + data.spe.Locality
		c.Draw(txt)
	}
	if data.spe.Georef.IsValid() {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = fmt.Sprintf("LonLat: %.3f,%.3f", data.spe.Georef.Point.Lon, data.spe.Georef.Point.Lat)
		c.Draw(txt)
		if data.spe.Georef.Uncertainty != 0 {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = fmt.Sprintf("Uncert: %d", data.spe.Georef.Uncertainty)
			c.Draw(txt)
		}
		if len(data.spe.Georef.Source) > 0 {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "Source: " + data.spe.Georef.Source
			c.Draw(txt)
		}
		if len(data.spe.Georef.Validation) > 0 {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "Val: " + data.spe.Georef.Validation
			c.Draw(txt)
		}
	}
	if len(data.spe.Extern) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Extern ids:"
		c.Draw(txt)
		for _, e := range data.spe.Extern {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "    " + e
			c.Draw(txt)
		}
	}
	if len(data.spe.Comment) > 0 {
		txt.Pos.Y += sparta.HeightUnit
		txt.Text = "Comments:"
		c.Draw(txt)
		cmt := strings.Split(data.spe.Comment, "\n")
		for _, e := range cmt {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "  " + e
			c.Draw(txt)
		}
	}
	if data.set != nil {
		if len(data.set.Citation) > 0 {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "Citation: " + data.set.Citation
			c.Draw(txt)
		}
		if len(data.set.License) > 0 {
			txt.Pos.Y += sparta.HeightUnit
			txt.Text = "License: " + data.set.License
			c.Draw(txt)
		}
	}
	return false
}

func spNavInitTaxList(m, l sparta.Widget, db jdh.DB, data *txList, i int) {
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
}

func spNavAncList(m, l sparta.Widget, db jdh.DB, tax *jdh.Taxon) {
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
}

func spNavInitSpeList(m, s sparta.Widget) {
	d := m.Property(sparta.Data)
	l := wnd["speList"]
	if d == nil {
		l.SetProperty(widget.ListList, nil)
		return
	}
	data := d.(*txList)
	if len(data.sels) == 0 {
		l.SetProperty(widget.ListList, nil)
		return
	}
	tax := data.desc[data.sels[0]]
	ls := newSpList(tax, data.db)
	l.SetProperty(widget.ListList, ls)
}

func spNavInfo(tx sparta.Widget, db jdh.DB, spe *jdh.Specimen, tax *jdh.Taxon) {
	if spe == nil {
		tx.SetProperty(sparta.Data, nil)
		tx.Update()
		return
	}
	if spe.Taxon != tax.Id {
		tx.SetProperty(sparta.Data, nil)
		tx.Update()
		return
	}
	info := &spInfo{
		tax: tax,
		spe: spe,
	}
	if len(spe.Dataset) > 0 {
		info.set = dataset(cmd, db, spe.Dataset)
	}
	tx.SetProperty(sparta.Data, info)
	tx.Update()
}
