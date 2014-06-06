// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"github.com/js-arias/cmdapp"
	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/sparta"
	"github.com/js-arias/sparta/widget"
)

var trView = &cmdapp.Command{
	Name:     "tr.view",
	Synopsis: `[-p|--port value] [-s|--set]`,
	Short:    "displays a tree",
	Long: `
Description

Tr.view displays a tree. By default the tree will be fitted to the window.

If there are more than one tree in the database, you can use space, enter keys
to move to the next tree, and backspace to move to previous tree.

Options

    -p value
    --port value
      Sets the port in which the server will be listening. By default the
      value is ":16917"
    
    -s
    --set
      If set, the tree can be edited with the mouse.
	`,
}

func init() {
	trView.Flag.StringVar(&portFlag, "port", "", "")
	trView.Flag.StringVar(&portFlag, "p", "", "")
	trView.Flag.BoolVar(&setFlag, "set", false, "")
	trView.Flag.BoolVar(&setFlag, "s", false, "")
	trView.Run = trViewRun
}

func trViewRun(c *cmdapp.Command, args []string) {
	cmd = c
	openLocal(c)
	title := fmt.Sprintf("%s: please wait", c.Name)
	m := widget.NewMainWindow("main", title)
	geo := m.Property(sparta.Geometry).(image.Rectangle)

	tv := widget.NewCanvas(m, "tree", image.Rect(0, 0, geo.Dx(), geo.Dy()))
	wnd["tree"] = tv
	tv.Capture(sparta.Expose, trViewExpose)
	tv.Capture(sparta.KeyEv, trViewKey)
	tv.Capture(sparta.Mouse, trViewMouse)
	tv.Update()

	m.Capture(sparta.Configure, trViewConf)
	go trViewInitList(m, tv)

	sparta.Run()
}

func trViewConf(m sparta.Widget, e interface{}) bool {
	ev := e.(sparta.ConfigureEvent)
	tv := wnd["tree"]
	tv.SetProperty(sparta.Geometry, image.Rect(0, 0, ev.Rect.Dx(), ev.Rect.Dy()))
	return false
}

func trViewExpose(tv sparta.Widget, e interface{}) bool {
	dt := tv.Property(sparta.Data)
	if dt == nil {
		return false
	}
	data := dt.(*trData)
	draw := tv.(*widget.Canvas)
	for _, n := range data.node {
		draw.Draw(n.ancLine)
		if n.level > 0 {
			draw.Draw(n.descLine)
		} else if len(n.name.Text) > 0 {
			draw.Draw(n.name)
		}
	}
	if n := data.sel; n != nil {
		rect := widget.Rectangle{
			Rect: image.Rect(n.pos.X-3, n.pos.Y-3, n.pos.X+3, n.pos.Y+3),
		}
		draw.Draw(rect)
		rect.Rect = image.Rect(n.pos.X-2, n.pos.Y-2, n.pos.X+2, n.pos.Y+2)
		rect.Fill = true
		draw.SetColor(sparta.Foreground, color.RGBA{G: 255})
		draw.Draw(rect)
	}
	return false
}

func trViewKey(tv sparta.Widget, e interface{}) bool {
	dt := tv.Property(sparta.Data)
	if dt == nil {
		return true
	}
	data := dt.(*trData)
	rect := tv.Property(sparta.Geometry).(image.Rectangle)
	ev := e.(sparta.KeyEvent)
	switch ev.Key {
	case sparta.KeyDown:
		data.pos.Y -= 5
	case sparta.KeyUp:
		data.pos.Y += 5
	case sparta.KeyLeft:
		data.pos.X -= 5
	case sparta.KeyRight:
		data.pos.X += 5
	case sparta.KeyHome:
		data.pos = image.Pt(0, 0)
	case sparta.KeyPageUp:
		data.pos.Y += rect.Dy() - sparta.HeightUnit
	case sparta.KeyPageDown:
		data.pos.Y -= rect.Dy() - sparta.HeightUnit
	case ' ', sparta.KeyReturn:
		p := tv.Property(sparta.Parent).(sparta.Widget)
		d := p.Property(sparta.Data).(*trList)
		if (d.pos + 1) >= len(d.phyLs) {
			return false
		}
		d.pos++
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		p.SetProperty(sparta.Caption, title)
		tv.SetProperty(sparta.Data, nil)
		go trViewInitTree(p, tv)
	case sparta.KeyBackSpace:
		p := tv.Property(sparta.Parent).(sparta.Widget)
		d := p.Property(sparta.Data).(*trList)
		if (d.pos - 1) < 0 {
			return false
		}
		d.pos--
		tv.SetProperty(sparta.Data, nil)
		title := fmt.Sprintf("%s: please wait", cmd.Name)
		p.SetProperty(sparta.Caption, title)
		go trViewInitTree(p, tv)
	case '+':
		data.y = data.y * 5 / 4
	case '-':
		data.y = data.y * 4 / 5
	case '*':
		data.x = data.x * 5 / 4
	case '/':
		data.x = data.x * 4 / 5
	case '#':
		root := data.node[0]
		data.y = float32(rect.Dy()-10) / float32(root.terms+2)
		data.x = float32(rect.Dx()-10-(sparta.WidthUnit*32)) / float32(root.level)
	case '=':
		data.y = float32(sparta.HeightUnit)
		data.x = float32(sparta.WidthUnit * 2)
	case '>':
		if !data.aln {
			return false
		}
		data.aln = false
	case '<':
		if data.aln {
			return false
		}
		data.aln = true
	default:
		return true
	}
	data.putOnScreen()
	tv.Update()
	return false
}

func trViewMouse(tv sparta.Widget, e interface{}) bool {
	dt := tv.Property(sparta.Data)
	if dt == nil {
		return true
	}
	data := dt.(*trData)
	ev := e.(sparta.MouseEvent)
	switch ev.Button {
	case sparta.MouseRight:
		if !setFlag {
			return true
		}
		if data.sel == nil {
			return true
		}
		sel := trViewNearestNode(ev.Loc, data.node)
		if sel == nil {
			return true
		}
		x, y, pos := data.x, data.y, data.pos
		p := tv.Property(sparta.Parent).(sparta.Widget)
		d := p.Property(sparta.Data).(*trList)
		if sel == data.sel {
			vals := new(jdh.Values)
			vals.Add(jdh.NodCollapse, sel.id)
			localDB.Exec(jdh.Delete, jdh.Nodes, vals)
			localDB.Exec(jdh.Commit, "", nil)
		} else if !sel.isValidSis(data.sel) {
			return true
		} else {
			vals := new(jdh.Values)
			vals.Add(jdh.KeyId, data.sel.id)
			vals.Add(jdh.NodSister, sel.id)
			localDB.Exec(jdh.Set, jdh.Nodes, vals)
			localDB.Exec(jdh.Commit, "", nil)
		}
		rect := tv.Property(sparta.Geometry).(image.Rectangle)
		data = setTree(d.phyLs[d.pos], rect)
		data.x, data.y, data.pos = x, y, pos
		tv.SetProperty(sparta.Data, data)
		data.putOnScreen()
		tv.Update()
	case sparta.MouseLeft:
		data.sel = trViewNearestNode(ev.Loc, data.node)
		tv.Update()
	case -sparta.MouseWheel:
		data.pos.Y -= 5
		data.putOnScreen()
		tv.Update()
	case sparta.MouseWheel:
		data.pos.Y += 5
		data.putOnScreen()
		tv.Update()
	}
	return true
}

func trViewNearestNode(pt image.Point, node []*trNode) *trNode {
	var trn *trNode
	best := 11
	for _, n := range node {
		dif := pt.Sub(n.pos)
		if dif.X < 0 {
			dif.X = -dif.X
		}
		if dif.Y < 0 {
			dif.Y = -dif.Y
		}
		if d := dif.X + dif.Y; d < best {
			best = d
			trn = n
		}
	}
	return trn
}

func trViewInitList(m, tv sparta.Widget) {
	l, err := localDB.List(jdh.Trees, new(jdh.Values))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", cmd.ErrStr(err))
		return
	}
	data := &trList{}
	for {
		phy := &jdh.Phylogeny{}
		if err := l.Scan(phy); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", cmd.ErrStr(err))
			return
		}
		if (len(phy.Id) == 0) || (len(phy.Root) == 0) {
			continue
		}
		data.phyLs = append(data.phyLs, phy)
	}
	if len(data.phyLs) == 0 {
		return
	}
	m.SetProperty(sparta.Data, data)
	trViewInitTree(m, tv)
}

func trViewInitTree(m, tv sparta.Widget) {
	d := m.Property(sparta.Data).(*trList)
	if d.pos >= len(d.phyLs) {
		return
	}
	title := fmt.Sprintf("%s: %s [id: %s]", cmd.Name, d.phyLs[d.pos].Name, d.phyLs[d.pos].Id)
	m.SetProperty(sparta.Caption, title)
	rect := tv.Property(sparta.Geometry).(image.Rectangle)
	curTree := setTree(d.phyLs[d.pos], rect)
	curTree.putOnScreen()
	tv.SetProperty(sparta.Data, curTree)
	tv.Update()
}
