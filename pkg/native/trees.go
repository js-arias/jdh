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

	"github.com/js-arias/jdh/pkg/jdh"
)

// Trees holds the phylogenetic trees of the database.
type trees struct {
	db      *DB
	ids     map[string]*phylogeny // map of id:phylogeny
	ls      *list.List            // list of phylogenies
	nodes   map[string]*node      // map of id:node
	changed bool                  // true if the database has changed
	nxTree  int64                 // next valid tree id
	nxNode  int64                 // next valid tree node id
}

// Phylogeny holds a phylogenetic tree.
type phylogeny struct {
	data  *jdh.Phylogeny
	root  *node
	nodes map[string]*node // map of id:node
	taxa  map[string]*node // map of taxon-id:node
	elem  *list.Element
}

// Node holds a node in a phylogenetic tree.
type node struct {
	data *jdh.Node

	phylog *phylogeny

	// relations
	parent *node
	childs []*node
}

// files
const treFile = "trees"
const nodFile = "nodes"

// OpenTrees open phylogenetic tree data.
func openTrees(db *DB) *trees {
	tr := &trees{
		db:     db,
		ids:    make(map[string]*phylogeny),
		nodes:  make(map[string]*node),
		ls:     list.New(),
		nxTree: 1,
		nxNode: 1,
	}
	if !tr.openTreFile() {
		return tr
	}
	tr.openNodFile()
	return tr
}

// OpenTreFile read the phylogenetic tree data.
func (tr *trees) openTreFile() bool {
	p := filepath.Join(tr.db.path, treFile)
	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		phy := &jdh.Phylogeny{}
		if err := dec.Decode(phy); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-trees: error: %v\n", err)
			break
		}
		tr.setNxPhy(phy.Id)
		if err := tr.valPhy(phy); err != nil {
			log.Printf("db-trees: error: %v\n", err)
			continue
		}
		tr.addValPhy(phy)
	}
	return true
}

// SetNxPhy sets the value of the next valid tree id.
func (tr *trees) setNxPhy(id string) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}
	if v >= tr.nxTree {
		tr.nxTree = v + 1
	}
}

// ValPhy validates that a tree is valid in the database, and set some
// canonical values. It returns an error if the tree is not valid.
func (tr *trees) valPhy(phy *jdh.Phylogeny) error {
	phy.Id = strings.TrimSpace(phy.Id)
	if len(phy.Id) == 0 {
		return errors.New("phylogeny without identification")
	}
	if _, ok := tr.ids[phy.Id]; ok {
		return fmt.Errorf("phylogeny id %s alredy in use", phy.Id)
	}
	phy.Name = strings.Join(strings.Fields(phy.Name), " ")
	phy.Root = ""
	ext := phy.Extern
	phy.Extern = nil
	for _, e := range ext {
		serv, id, err := jdh.ParseExtern(e)
		if err != nil {
			continue
		}
		if len(id) == 0 {
			continue
		}
		add := true
		for _, ex := range phy.Extern {
			if strings.HasPrefix(ex, serv) {
				add = false
				break
			}
		}
		if !add {
			continue
		}
		if _, ok := tr.ids[e]; !ok {
			phy.Extern = append(phy.Extern, e)
		}
	}
	return nil
}

// AddValPhy adds a validated phylogeny into database.
func (tr *trees) addValPhy(phy *jdh.Phylogeny) {
	ph := &phylogeny{
		data:  phy,
		nodes: make(map[string]*node),
		taxa:  make(map[string]*node),
	}
	tr.ids[phy.Id] = ph
	for _, e := range phy.Extern {
		tr.ids[e] = ph
	}
	ph.elem = tr.ls.PushBack(ph)
}

// OpenNodFile open nodes file.
func (tr *trees) openNodFile() {
	p := filepath.Join(tr.db.path, nodFile)
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	for {
		nod := &jdh.Node{}
		if err := dec.Decode(nod); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("db-trees: error: %v\n", err)
			break
		}
		tr.setNxNode(nod.Id)
		if err := tr.valNod(nod); err != nil {
			log.Printf("db-trees: error: %v\n", err)
		}
		tr.addValNode(nod)
	}
}

// SetNxNode sets the value of the next node id.
func (tr *trees) setNxNode(id string) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}
	if v >= tr.nxNode {
		tr.nxNode = v + 1
	}
}

// ValNode validates that a node is valid in the database, and set some
// canonical values. It returns an error if the node is not valid.
func (tr *trees) valNod(nod *jdh.Node) error {
	nod.Id = strings.TrimSpace(nod.Id)
	nod.Tree = strings.TrimSpace(nod.Tree)
	if (len(nod.Id) == 0) || (len(nod.Tree) == 0) {
		return errors.New("node without identification")
	}
	if _, ok := tr.nodes[nod.Id]; ok {
		return fmt.Errorf("node id %s already in use", nod.Id)
	}
	ph, ok := tr.ids[nod.Tree]
	if !ok {
		return fmt.Errorf("node %s in a not assigned phylogeny [id %s]", nod.Id, nod.Tree)
	}
	nod.Parent = strings.TrimSpace(nod.Parent)
	if len(nod.Parent) == 0 {
		if ph.root != nil {
			return fmt.Errorf("node %s without a parent in phylogeny %s", nod.Id, ph.data.Id)
		}
	} else {
		p, ok := ph.nodes[nod.Parent]
		if !ok {
			return fmt.Errorf("node %s without a parent in phylogeny %s", nod.Id, ph.data.Id)
		}
		if (nod.Age > 0) && (p.data.Age < nod.Age) {
			return fmt.Errorf("node %s age %d is older than parent %s age %d", nod.Id, nod.Age, p.data.Id, p.data.Age)
		}
	}
	nod.Taxon = strings.TrimSpace(nod.Taxon)
	if len(nod.Taxon) > 0 {
		if !tr.db.t.isInDB(nod.Taxon) {
			nod.Taxon = ""
		} else if _, ok := ph.taxa[nod.Taxon]; ok {
			return fmt.Errorf("taxon %s assigned to node %s already assigned to phylogeny %s", nod.Taxon, nod.Id, ph.data.Id)
		}
	}
	return nil
}

// AddValNode adds a new node to the database.
func (tr *trees) addValNode(nod *jdh.Node) {
	ph := tr.ids[nod.Tree]
	nd := &node{
		data:   nod,
		phylog: ph,
	}
	if len(nod.Parent) == 0 {
		ph.data.Root = nod.Id
		ph.root = nd
	} else {
		p := ph.nodes[nod.Parent]
		p.childs = append(p.childs, nd)
		nd.parent = p
	}
	ph.nodes[nod.Id] = nd
	if len(nod.Taxon) > 0 {
		ph.taxa[nod.Taxon] = nd
	}
	tr.nodes[nod.Id] = nd
}

// AddTree adds a new tree to the database.
func (tr *trees) addTree(phy *jdh.Phylogeny) (string, error) {
	id := strconv.FormatInt(tr.nxTree, 10)
	phy.Id = id
	if err := tr.valPhy(phy); err != nil {
		return "", err
	}
	tr.addValPhy(phy)
	tr.nxTree++
	tr.changed = true
	return id, nil
}

// AddNode adds a new node to the database.
func (tr *trees) addNode(nod *jdh.Node) (string, error) {
	id := strconv.FormatInt(tr.nxNode, 10)
	nod.Id = id
	if err := tr.valNod(nod); err != nil {
		return "", err
	}
	tr.addValNode(nod)
	tr.nxNode++
	tr.changed = true
	return id, nil
}

// Commit saves the trees into hard disk.
func (tr *trees) commit(e chan error) {
	if !tr.changed {
		e <- nil
		return
	}
	if err := tr.comTree(); err != nil {
		e <- err
		return
	}
	if err := tr.comNode(); err != nil {
		e <- err
		return
	}
	tr.changed = false
	e <- nil
}

// ComTree commits a phylogenetic tree.
func (tr *trees) comTree() error {
	p := filepath.Join(tr.db.path, treFile)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for e := tr.ls.Front(); e != nil; e = e.Next() {
		ph := e.Value.(*phylogeny)
		enc.Encode(ph.data)
	}
	return nil
}

// ComNode commits the nodes of a phylogeny.
func (tr *trees) comNode() error {
	p := filepath.Join(tr.db.path, nodFile)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for e := tr.ls.Front(); e != nil; e = e.Next() {
		ph := e.Value.(*phylogeny)
		ph.root.encode(enc)
	}
	return nil
}

// Encode encodes a node into a json blob in the database.
func (nd *node) encode(enc *json.Encoder) {
	enc.Encode(nd.data)
	for _, d := range nd.childs {
		d.encode(enc)
	}
}

// DeleteTree deletes a tree, or a taxon from a tree.
func (tr *trees) deleteTree(vals []jdh.KeyValue) error {
	id := ""
	tax := ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			id = kv.Value[0]
			break
		}
		if kv.Key == jdh.TreTaxon {
			tax = kv.Value[0]
			break
		}
	}
	if (len(id) == 0) && (len(tax) == 0) {
		return errors.New("tree without identification")
	}
	if len(tax) > 0 {
		tr.delTaxonFromAll(tax)
		return nil
	}
	ph, ok := tr.ids[id]
	if !ok {
		return nil
	}
	tax = ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.NodTaxon {
			tax = kv.Value[0]
			break
		}
	}
	if len(tax) > 0 {
		tr.delTaxon(ph, tax)
		return nil
	}
	tr.delTree(ph)
	tr.changed = true
	return nil
}

// DeleteNode deletes a node, or a taxon from a tree.
func (tr *trees) deleteNode(vals []jdh.KeyValue) error {
	id := ""
	coll := false
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.KeyId {
			id = kv.Value[0]
			break
		}
		if kv.Key == jdh.NodCollapse {
			id = kv.Value[0]
			coll = true
			break
		}
	}
	if len(id) == 0 {
		return errors.New("node without identification")
	}
	nd, ok := tr.nodes[id]
	if !ok {
		return nil
	}
	if coll {
		tr.colNode(nd)
		tr.changed = true
		return nil
	}
	tax := ""
	for _, kv := range vals {
		if len(kv.Value) == 0 {
			continue
		}
		if kv.Key == jdh.NodTaxon {
			tax = kv.Value[0]
			break
		}
	}
	if len(tax) > 0 {
		tr.delTaxon(nd.phylog, tax)
		tr.changed = true
		return nil
	}
	p := tr.delNode(nd)
	if (p != nil) && len(p.childs) < 2 {
		tr.colNode(p)
	}
	tr.changed = true
	return nil
}

// DelTaxonFromAll removes a taxon from all trees.
func (tr *trees) delTaxonFromAll(tax string) {
	if !tr.db.t.isInDB(tax) {
		return
	}
	for e := tr.ls.Front(); e != nil; e = e.Next() {
		ph := e.Value.(*phylogeny)
		tr.delTaxon(ph, tax)
	}
}

// DelTaxon removes a taxon from a tree.
func (tr *trees) delTaxon(ph *phylogeny, tax string) {
	nd, ok := ph.taxa[tax]
	if !ok {
		return
	}
	nd.data.Taxon = ""
	delete(ph.taxa, tax)
	tr.changed = true
}

// DelTree removes a tree from the database.
func (tr *trees) delTree(ph *phylogeny) {
	if ph.root != nil {
		tr.delNode(ph.root)
	}
	ph.root = nil
	ph.nodes = nil
	ph.taxa = nil
	delete(tr.ids, ph.data.Id)
	ph.data = nil
	tr.ls.Remove(ph.elem)
	ph.elem = nil
}

// DelNode recursively removes nodes from the database, and returns its
// parent.
func (tr *trees) delNode(nd *node) *node {
	for len(nd.childs) > 0 {
		tr.delNode(nd.childs[0])
	}
	nd.childs = nil
	ph := nd.phylog
	nod := nd.data
	delete(ph.nodes, nod.Id)
	if len(nod.Taxon) > 0 {
		delete(ph.taxa, nod.Taxon)
	}
	delete(tr.nodes, nod.Id)
	p := nd.parent
	if p != nil {
		p.childs = delNodeFromList(p.childs, nd)
	}
	nd.parent = nil
	nd.data = nil
	return p
}

// DelNodeFromList remves a node pointer from a list of nodes.
func delNodeFromList(ls []*node, nd *node) []*node {
	if len(ls) == 0 {
		return nil
	}
	if len(ls) == 1 {
		if ls[0] == nd {
			return nil
		}
		return ls
	}
	for i, on := range ls {
		if on == nd {
			copy(ls[i:], ls[i+1:])
			ls[len(ls)-1] = nil
			return ls[:len(ls)-1]
		}
	}
	return ls
}

// ColNode collapses a node, and then delete it.
func (tr *trees) colNode(nd *node) {
	if len(nd.childs) == 0 {
		return
	}
	p := nd.parent
	if p == nil {
		if len(nd.childs) == 1 {
			d := nd.childs[0]
			d.parent = nil
			d.data.Parent = ""
			nd.childs = nil
			nd.phylog.root = d
			tr.delNode(nd)
		}
		return
	}
	for _, d := range nd.childs {
		d.parent = p
		d.data.Parent = p.data.Id
		p.childs = append(p.childs, d)
	}
	nd.childs = nil
	tr.delNode(nd)
}

// GetTree returns a tree with a given id.
func (tr *trees) getTree(id string) (*jdh.Phylogeny, error) {
	if len(id) == 0 {
		return nil, errors.New("tree without identificiation")
	}
	ph, ok := tr.ids[id]
	if !ok {
		return nil, nil
	}
	return ph.data, nil
}

// GetNode returns a node with a given id.
func (tr *trees) getNode(id string) (*jdh.Node, error) {
	if len(id) == 0 {
		return nil, errors.New("node without identificiation")
	}
	nd, ok := tr.nodes[id]
	if !ok {
		return nil, nil
	}
	return nd.data, nil
}

// ListTree returns a list of trees.
func (tr *trees) listTree(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	// creates a list of taxons in a tree
	for _, kv := range vals {
		if kv.Key != jdh.TreTaxon {
			continue
		}
		if len(kv.Value) == 0 {
			return l, nil
		}
		v := strings.TrimSpace(kv.Value[0])
		if len(v) == 0 {
			return l, nil
		}
		ph, ok := tr.ids[v]
		if !ok {
			return l, nil
		}
		for tx, _ := range ph.taxa {
			l.PushBack(jdh.IdElement{Id: tx})
		}
		return l, nil
	}
	// creates the list of trees
	for e := tr.ls.Front(); e != nil; e = e.Next() {
		ph := e.Value.(*phylogeny)
		l.PushBack(ph.data)
	}
	return l, nil
}

// ListNode returns a list of nodes.
func (tr *trees) listNode(vals []jdh.KeyValue) (*list.List, error) {
	l := list.New()
	noVal := true
	for _, kv := range vals {
		switch kv.Key {
		case jdh.NodChildren:
			if len(kv.Value) == 0 {
				break
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				break
			}
			noVal = false
			nd, ok := tr.nodes[v]
			if !ok {
				break
			}
			for _, c := range nd.childs {
				l.PushBack(c.data)
			}
		case jdh.NodParent:
			if len(kv.Value) == 0 {
				break
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				break
			}
			noVal = false
			nd, ok := tr.nodes[v]
			if !ok {
				break
			}
			for p := nd.parent; p != nil; p = p.parent {
				if p.data == nil {
					break
				}
				l.PushBack(p.data)
			}
		case jdh.NodTree:
			if len(kv.Value) == 0 {
				break
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				break
			}
			noVal = false
			ph, ok := tr.ids[v]
			if !ok {
				break
			}
			if ph.root == nil {
				break
			}
			addNodeToList(l, ph.root)
		case jdh.NodTaxon:
			if len(kv.Value) == 0 {
				break
			}
			v := strings.TrimSpace(kv.Value[0])
			if len(v) == 0 {
				break
			}
			noVal = false
			if !tr.db.t.isInDB(v) {
				break
			}
			for e := tr.ls.Front(); e != nil; e = e.Next() {
				ph := e.Value.(*phylogeny)
				nd, ok := ph.taxa[v]
				if !ok {
					continue
				}
				l.PushBack(nd.data)
			}
		}
		if !noVal {
			break
		}
	}
	if noVal {
		return nil, errors.New("node without identification")
	}
	return l, nil
}

// addNodeToList adds a node, and all of its descendants to a list.
func addNodeToList(l *list.List, nd *node) {
	l.PushBack(nd.data)
	for _, c := range nd.childs {
		addNodeToList(l, c)
	}
}

// SetTree sets a value of a tree in the database.
func (tr *trees) setTree(vals []jdh.KeyValue) error {
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
		return errors.New("tree without identification")
	}
	ph, ok := tr.ids[id]
	if !ok {
		return nil
	}
	phy := ph.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.TreName:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.Join(strings.Fields(kv.Value[0]), " ")
			}
			if phy.Name == v {
				continue
			}
			phy.Name = v
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if phy.Comment == v {
				continue
			}
			phy.Comment = v
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
					if !tr.delExtern(ph, serv) {
						continue
					}
					ok = true
					continue
				}
				if tr.addExtern(ph, v) != nil {
					continue
				}
				ok = true
			}
			if !ok {
				continue
			}
		default:
			continue
		}
		tr.changed = true
	}
	return nil
}

// AddExtern adds an extern id to an specimen.
func (tr *trees) addExtern(ph *phylogeny, extern string) error {
	serv, id, err := jdh.ParseExtern(extern)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return nil
	}
	if or, ok := tr.ids[extern]; ok {
		return fmt.Errorf("extern id %s of %s alredy in use by %s", extern, ph.data.Id, or.data.Id)
	}
	// the service is already assigned, then overwrite
	for i, e := range ph.data.Extern {
		if strings.HasPrefix(e, serv) {
			delete(tr.ids, e)
			ph.data.Extern[i] = extern
			tr.ids[extern] = ph
			return nil
		}
	}
	ph.data.Extern = append(ph.data.Extern, extern)
	tr.ids[extern] = ph
	return nil
}

// DelExtern deletes an extern id of an specimen.
func (tr *trees) delExtern(ph *phylogeny, service string) bool {
	for i, e := range ph.data.Extern {
		if strings.HasPrefix(e, service) {
			delete(tr.ids, e)
			copy(ph.data.Extern[i:], ph.data.Extern[i+1:])
			ph.data.Extern[len(ph.data.Extern)-1] = ""
			ph.data.Extern = ph.data.Extern[:len(ph.data.Extern)-1]
			return true
		}
	}
	return false
}

// SetNode sets a value of a node in the database.
func (tr *trees) setNode(vals []jdh.KeyValue) error {
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
		return errors.New("node without identification")
	}
	nd, ok := tr.nodes[id]
	if !ok {
		return nil
	}
	ph := nd.phylog
	nod := nd.data
	for _, kv := range vals {
		switch kv.Key {
		case jdh.NodAge:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(v)
			}
			if len(v) == 0 {
				if nod.Age == 0 {
					continue
				}
				nod.Age = 0
				break
			}
			a, err := strconv.ParseUint(v, 10, 0)
			if err != nil {
				return err
			}
			nod.Age = uint(a)
		case jdh.NodLength:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(v)
			}
			if len(v) == 0 {
				if nod.Len == 0 {
					continue
				}
				nod.Len = 0
				break
			}
			l, err := strconv.ParseUint(v, 10, 0)
			if err != nil {
				return err
			}
			nod.Len = uint(l)
		case jdh.NodSister:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if len(v) == 0 {
				continue
			}
			sis, ok := ph.nodes[v]
			if !ok {
				return fmt.Errorf("node %s not in tree %s", v, ph.data.Id)
			}
			if !sis.isValidAnc(nd) {
				continue
			}
			p := nd.parent
			if p == nil {
				continue
			}
			if (p == sis.parent) && (len(p.childs) == 2) {
				continue
			}
			ph := nd.phylog
			if len(p.childs) == 2 {
				ch := p.childs[0]
				if ch == nd {
					ch = p.childs[1]
				}
				anc := p.parent
				if anc != nil {
					p.childs = delNodeFromList(p.childs, ch)
					anc.childs = delNodeFromList(anc.childs, p)
					anc.childs = append(anc.childs, ch)
					ch.parent = anc
					ch.data.Parent = anc.data.Id
				} else {
					if len(ch.childs) == 0 {
						continue
					}
					ch.parent = nil
					ch.data.Parent = ""
					ph.root = ch
					ph.data.Root = ch.data.Id
				}
				np := sis.parent
				if np != nil {
					np.childs = delNodeFromList(np.childs, sis)
					np.childs = append(np.childs, p)
					p.parent = np
					p.data.Parent = np.data.Id
					p.childs = append(p.childs, sis)
					sis.parent = p
					sis.data.Parent = p.data.Id
					break
				}
				p.childs = append(p.childs, sis)
				sis.parent = p
				sis.data.Parent = p.data.Id
				ph.root = p
				ph.data.Root = p.data.Id
				break
			}
			np := sis.parent
			nId := ""
			if np != nil {
				p.childs = delNodeFromList(p.childs, nd)
				np.childs = delNodeFromList(np.childs, sis)
				nu := &jdh.Node{
					Tree:   nod.Tree,
					Parent: np.data.Id,
				}
				nId, _ = tr.addNode(nu)
				p = tr.nodes[nId]
			} else {
				ph.root = nil
				nu := &jdh.Node{
					Tree: nod.Tree,
				}
				nId, _ = tr.addNode(nu)
				p = tr.nodes[nId]
				ph.data.Root = nId
			}
			p.childs = append(p.childs, sis)
			sis.parent = p
			sis.data.Parent = nId
			p.childs = append(p.childs, nd)
			nd.parent = p
			nod.Parent = nId
		case jdh.NodTaxon:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if nod.Taxon == v {
				continue
			}
			if len(v) > 0 {
				if _, ok := ph.taxa[v]; ok {
					return fmt.Errorf("taxon %s already in tree %s", v, nod.Tree)
				}
				if !tr.db.t.isInDB(v) {
					continue
				}
				if len(nod.Taxon) > 0 {
					delete(ph.taxa, nod.Taxon)
				}
				nod.Taxon = v
				ph.taxa[v] = nd
				break
			}
			if len(nod.Taxon) > 0 {
				delete(ph.taxa, nod.Taxon)
			}
			nod.Taxon = ""
		case jdh.KeyComment:
			v := ""
			if len(kv.Value) > 0 {
				v = strings.TrimSpace(kv.Value[0])
			}
			if nod.Comment == v {
				continue
			}
			nod.Comment = v
		default:
			continue
		}
		tr.changed = true
	}
	return nil
}

func (nd *node) isValidAnc(on *node) bool {
	if nd == on {
		return false
	}
	for p := nd.parent; p != nil; p = p.parent {
		if p == on {
			return false
		}
	}
	return true
}
