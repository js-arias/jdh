// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/sparta"
	_ "github.com/js-arias/sparta/init"
	"github.com/js-arias/sparta/widget"
)

var txEd = &cmdapp.Command{
	Name:     "tx.ed",
	Synopsis: `[-p|--port value]`,
	Short:    "edits taxonomy",
	Long: `
Description

Tx.ed edits the taxonomic hierarchy of the taxonomy stored in the database.

Options

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
	`,
}

func init() {
	txEd.Flag.StringVar(&portFlag, "port", "", "")
	txEd.Flag.StringVar(&portFlag, "p", "", "")
	txEd.Run = txEdRun
}

func txEdRun(c *cmdapp.Command, args []string) {
	cmd = c
	openLocal(c)
	title := fmt.Sprintf("%s: please wait", c.Name)
	m := widget.NewMainWindow("main", title)
	geo := m.Property(sparta.Geometry).(image.Rectangle)
	widget.NewButton(m, "upTax", "up", image.Rect(5, 5, 5+(sparta.WidthUnit*10), 5+(3*sparta.HeightUnit/2)))
	widget.NewButton(m, "val", "validate", image.Rect(10+(sparta.WidthUnit*10), 5, 10+(sparta.WidthUnit*20), 5+(3*sparta.HeightUnit/2)))
	widget.NewButton(m, "upVal", "up", image.Rect(270, 5, 270+(sparta.WidthUnit*10), 5+(3*sparta.HeightUnit/2)))
	widget.NewButton(m, "syn", "syn ->", image.Rect(205, 10+(3*sparta.HeightUnit/2), 205+(sparta.WidthUnit*10), 10+(6*sparta.HeightUnit/2)))
	widget.NewButton(m, "move", "move ->", image.Rect(205, 15+(6*sparta.HeightUnit/2), 205+(sparta.WidthUnit*10), 15+(9*sparta.HeightUnit/2)))

	l := widget.NewList(m, "taxonList", image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, geo.Dy()-10))
	wnd["taxonList"] = l

	s := widget.NewList(m, "validList", image.Rect(210+(sparta.WidthUnit*10), 10+(3*sparta.HeightUnit/2), 405+(sparta.WidthUnit*10), geo.Dy()-10))
	wnd["validList"] = s

	m.Capture(sparta.Configure, txEdConf)
	m.Capture(sparta.Command, txEdComm)
	sparta.Block(nil)

	go func() {
		txEdInitList(m, l, nil, 0, true)
		txEdInitList(m, s, nil, 0, false)
		txEdSetCaption(m)
		sparta.Unblock(nil)
	}()

	sparta.Run()
}

func txEdConf(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.ConfigureEvent)
	l := wnd["taxonList"]
	l.SetProperty(sparta.Geometry, image.Rect(5, 10+(3*sparta.HeightUnit/2), 200, ev.Rect.Dy()-10))
	l = wnd["validList"]
	l.SetProperty(sparta.Geometry, image.Rect(210+(sparta.WidthUnit*10), 10+(3*sparta.HeightUnit/2), 405+(sparta.WidthUnit*10), ev.Rect.Dy()-10))

	return false
}

func txEdComm(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.CommandEvent)
	switch ev.Source.Property(sparta.Name).(string) {
	case "move":
		frWg := wnd["taxonList"]
		d := frWg.Property(widget.ListList)
		if d == nil {
			break
		}
		frLs := d.(*txList)
		if (len(frLs.sels) == 0) && (frLs.tax.Id == "0") {
			break
		}
		toWg := wnd["validList"]
		d = toWg.Property(widget.ListList)
		if d == nil {
			break
		}
		toLs := d.(*txList)
		var to *jdh.Taxon
		if len(toLs.sels) == 0 {
			to = toLs.tax
		} else {
			to = toLs.desc[toLs.sels[0]]
		}
		if (len(frLs.sels) == 0) && (frLs.tax.Id == to.Id) {
			break
		}
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		m.SetProperty(sparta.Caption, title)
		frWg.SetProperty(widget.ListList, nil)
		toWg.SetProperty(widget.ListList, nil)
		sparta.Block(nil)
		go func() {
			if len(frLs.sels) == 0 {
				txEdMove(frLs.tax, to)
			} else {
				for _, s := range frLs.sels {
					from := frLs.desc[s]
					txEdMove(from, to)
				}
			}
			localDB.Exec(jdh.Commit, "", nil)
			txEdUpdateList(m, frWg, frLs, true)
			txEdUpdateList(m, toWg, toLs, false)
			txEdSetCaption(m)
			sparta.Unblock(nil)
		}()
	case "syn":
		frWg := wnd["taxonList"]
		d := frWg.Property(widget.ListList)
		if d == nil {
			break
		}
		frLs := d.(*txList)
		if (len(frLs.sels) == 0) && (frLs.tax.Id == "0") {
			break
		}
		toWg := wnd["validList"]
		d = toWg.Property(widget.ListList)
		if d == nil {
			break
		}
		toLs := d.(*txList)
		var to *jdh.Taxon
		if len(toLs.sels) == 0 {
			to = toLs.tax
		} else {
			to = toLs.desc[toLs.sels[0]]
		}
		if to.Id == "0" {
			break
		}
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		m.SetProperty(sparta.Caption, title)
		frWg.SetProperty(widget.ListList, nil)
		toWg.SetProperty(widget.ListList, nil)
		sparta.Block(nil)
		go func() {
			if len(frLs.sels) == 0 {
				txEdSyn(frLs.tax, to)
			} else {
				for _, s := range frLs.sels {
					from := frLs.desc[s]
					txEdSyn(from, to)
				}
			}
			localDB.Exec(jdh.Commit, "", nil)
			txEdUpdateList(m, frWg, frLs, true)
			txEdUpdateList(m, toWg, toLs, false)
			txEdSetCaption(m)
			sparta.Unblock(nil)
		}()
	case "taxonList":
		d := ev.Source.Property(widget.ListList)
		if d == nil {
			break
		}
		data := d.(*txList)
		if ev.Value < 0 {
			i := -ev.Value - 1
			if i >= len(data.desc) {
				break
			}
			title := fmt.Sprintf("%s: please wait", cmd.Name)
			m.SetProperty(sparta.Caption, title)
			ev.Source.SetProperty(widget.ListList, nil)
			sparta.Block(nil)
			go func() {
				txEdInitList(m, ev.Source, data, i, true)
				txEdSetCaption(m)
				sparta.Unblock(nil)
			}()
			break
		}
		sel := true
		for j, s := range data.sels {
			if s == ev.Value {
				sel = false
				data.sels[j] = data.sels[len(data.sels)-1]
				data.sels = data.sels[:len(data.sels)-1]
				break
			}
		}
		if sel {
			data.sels = append(data.sels, ev.Value)
		}
		ev.Source.Update()
		break
	case "upTax":
		l := wnd["taxonList"]
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
		sparta.Block(nil)
		go func() {
			txEdAncList(m, l, data.tax, true)
			txEdSetCaption(m)
			sparta.Unblock(nil)
		}()
	case "upVal":
		l := wnd["validList"]
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
		sparta.Block(nil)
		go func() {
			txEdAncList(m, l, data.tax, false)
			txEdSetCaption(m)
			sparta.Unblock(nil)
		}()
	case "val":
		frWg := wnd["taxonList"]
		d := frWg.Property(widget.ListList)
		if d == nil {
			break
		}
		frLs := d.(*txList)
		if len(frLs.sels) == 0 {
			if frLs.tax.Id == "0" {
				break
			}
			if frLs.tax.IsValid {
				break
			}
		}
		toWg := wnd["validList"]
		d = toWg.Property(widget.ListList)
		if d == nil {
			break
		}
		toLs := d.(*txList)
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		m.SetProperty(sparta.Caption, title)
		frWg.SetProperty(widget.ListList, nil)
		toWg.SetProperty(widget.ListList, nil)
		sparta.Block(nil)
		go func() {
			if len(frLs.sels) == 0 {
				txEdVal(frLs.tax)
			} else {
				for _, s := range frLs.sels {
					from := frLs.desc[s]
					txEdVal(from)
				}
			}
			localDB.Exec(jdh.Commit, "", nil)
			txEdUpdateList(m, frWg, frLs, true)
			txEdUpdateList(m, toWg, toLs, false)
			txEdSetCaption(m)
			sparta.Unblock(nil)
		}()
	case "validList":
		d := ev.Source.Property(widget.ListList)
		if d == nil {
			break
		}
		data := d.(*txList)
		if ev.Value < 0 {
			i := -ev.Value - 1
			if i >= len(data.desc) {
				break
			}
			title := fmt.Sprintf("%s: please wait", cmd.Name)
			m.SetProperty(sparta.Caption, title)
			ev.Source.SetProperty(widget.ListList, nil)
			sparta.Block(nil)
			go func() {
				txEdInitList(m, ev.Source, data, i, false)
				txEdSetCaption(m)
				sparta.Unblock(nil)
			}()
			break
		}
		if data.IsSel(ev.Value) {
			data.sels = nil
		} else {
			data.sels = []int{ev.Value}
		}
		ev.Source.Update()
	}
	return true
}

func txEdSetCaption(m sparta.Widget) {
	d := m.Property(sparta.Data)
	if d == nil {
		m.SetProperty(sparta.Caption, "no data")
		return
	}
	dt := d.(*txList)
	title := fmt.Sprintf("%s: %s [id: %s]", cmd.Name, dt.tax.Name, dt.tax.Id)
	m.SetProperty(sparta.Caption, title)
}

func txEdUpdateList(m, l sparta.Widget, data *txList, syns bool) {
	if data == nil {
		data = newTxList(nil, localDB, syns)
	} else {
		d := newTxList(data.tax, data.db, syns)
		if len(d.desc) == 0 {
			txEdAncList(m, l, data.tax, syns)
			return
		}
		data = d
	}
	if l.Property(sparta.Name).(string) == "taxonList" {
		m.SetProperty(sparta.Data, data)
	}
	l.SetProperty(widget.ListList, data)
}

func txEdInitList(m, l sparta.Widget, data *txList, i int, syns bool) {
	if data == nil {
		data = newTxList(nil, localDB, syns)
	} else {
		d := newTxList(data.desc[i], data.db, syns)
		if len(d.desc) == 0 {
			if l.Property(sparta.Name).(string) == "validList" {
				data.sels = []int{i}
			} else {
				sel := false
				for _, s := range data.sels {
					if s == i {
						sel = true
						break
					}
				}
				if !sel {
					data.sels = append(data.sels, i)
				}
			}
		} else {
			data = d
		}
	}
	if l.Property(sparta.Name).(string) == "taxonList" {
		m.SetProperty(sparta.Data, data)
	}
	l.SetProperty(widget.ListList, data)
}

func txEdAncList(m, l sparta.Widget, tax *jdh.Taxon, syns bool) {
	var p *jdh.Taxon
	if len(tax.Parent) > 0 {
		p = taxon(cmd, localDB, tax.Parent)
		if len(p.Id) == 0 {
			p = nil
		}
	}
	data := newTxList(p, localDB, syns)
	for i, d := range data.desc {
		if d.Id == tax.Id {
			data.sels = []int{i}
			break
		}
	}
	if l.Property(sparta.Name).(string) == "taxonList" {
		m.SetProperty(sparta.Data, data)
	}
	l.SetProperty(widget.ListList, data)
}

func txEdMove(from, to *jdh.Taxon) {
	if from.Id == "0" {
		return
	}
	if (from.Id == to.Id) || (from.Parent == to.Id) {
		return
	}
	if (to.Id == "0") && (!from.IsValid) {
		return
	}
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, from.Id)
	if to.Id != "0" {
		vals.Add(jdh.TaxParent, to.Id)
	} else {
		vals.Add(jdh.TaxParent, "")
	}
	localDB.Exec(jdh.Set, jdh.Taxonomy, vals)
}

func txEdSyn(from, to *jdh.Taxon) {
	if from.Id == "0" {
		return
	}
	if from.Id == to.Id {
		return
	}
	if (from.Parent == to.Id) && (!from.IsValid) {
		return
	}
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, from.Id)
	vals.Add(jdh.TaxSynonym, to.Id)
	localDB.Exec(jdh.Set, jdh.Taxonomy, vals)
}

func txEdVal(from *jdh.Taxon) {
	if from.IsValid {
		return
	}
	vals := new(jdh.Values)
	vals.Add(jdh.KeyId, from.Id)
	vals.Add(jdh.TaxValid, "true")
	localDB.Exec(jdh.Set, jdh.Taxonomy, vals)
}
