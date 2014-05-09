// Copyright (c) 2014, J. Salvador Arias <jsalarias@csnat.unt.edu.ar>
// All rights reserved.
// Distributed under BSD2 license that can be found in LICENSE file.

// Package server implements the native jdh database server.
package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	ntv "github.com/js-arias/jdh/pkg/driver/native"
	"github.com/js-arias/jdh/pkg/jdh"
	"github.com/js-arias/jdh/pkg/native"
)

// server holds the information of a server
type server struct {
	ln   net.Listener
	conn chan net.Conn
	end  chan struct{}
	db   *native.DB
}

// Listen creates a server of a database in the local host.
func Listen(port, path string) error {
	srv := &server{
		conn: make(chan net.Conn, 10),
		end:  make(chan struct{}),
		db:   native.Open(path),
	}
	if len(port) == 0 {
		port = ntv.Port
	}
	var err error
	srv.ln, err = net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer srv.ln.Close()

	go func() {
		for {
			c, err := srv.ln.Accept()
			if err != nil {
				fmt.Fprintf(os.Stdout, "error [%s] %v\n", time.Now().Format("2006-Jan-2 15:04:05 -0700"), err)
				continue
			}
			select {
			case srv.conn <- c:
			case <-srv.end:
				return
			}
		}
	}()
	var done sync.WaitGroup
	for {
		select {
		case c := <-srv.conn:
			srv.handleConn(c, &done)
		case <-srv.end:
			done.Wait()
			return nil
		}
	}
}

// handleConn handles the connection
func (srv *server) handleConn(conn net.Conn, done *sync.WaitGroup) {
	remote := strings.Split(conn.RemoteAddr().String(), ":")[0]
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	req := &ntv.Request{}
	if err := dec.Decode(req); err != nil {
		ans := ntv.ErrAnswer(err.Error())
		enc.Encode(ans)
		log(remote, req, ans)
		conn.Close()
		return
	}
	table := req.Table
	switch req.Query {
	case jdh.Add:
		var ans *ntv.Answer
		if remote == "127.0.0.1" {
			if id, err := srv.db.Add(table, dec); err != nil {
				ans = ntv.ErrAnswer(err.Error())
			} else {
				ans = ntv.Success(id)
			}
		} else {
			ans = ntv.ErrAnswer("forbidden")
		}
		enc.Encode(ans)
		log(remote, req, ans)
	case jdh.Close:
		var ans *ntv.Answer
		if remote == "127.0.0.1" {
			ans = ntv.Success("ok")
		} else {
			ans = ntv.ErrAnswer("forbidden")
		}
		enc.Encode(ans)
		log(remote, req, ans)
		close(srv.end)
	case jdh.Commit:
		var ans *ntv.Answer
		if remote == "127.0.0.1" {
			if err := srv.db.Commit(); err != nil {
				ans = ntv.ErrAnswer(err.Error())
			} else {
				ans = ntv.Success("ok")
			}
		} else {
			ans = ntv.ErrAnswer("forbidden")
		}
		enc.Encode(ans)
		log(remote, req, ans)
	case jdh.Delete:
		var ans *ntv.Answer
		if remote == "127.0.0.1" {
			if len(req.Kvs) == 0 {
				ans = ntv.ErrAnswer("expecting arguments")
			} else {
				if err := srv.db.Delete(table, req.Kvs); err != nil {
					ans = ntv.ErrAnswer(err.Error())
				} else {
					ans = ntv.Success("ok")
				}
			}
		} else {
			ans = ntv.ErrAnswer("forbidden")
		}
		enc.Encode(ans)
		log(remote, req, ans)
	case jdh.Get:
		done.Add(1)
		go func() {
			defer conn.Close()
			defer done.Done()
			var id string
			for _, kv := range req.Kvs {
				if kv.Key == jdh.KeyId {
					id = kv.Value
				}
			}
			v, err := srv.db.Get(table, id)
			if err != nil {
				ans := ntv.ErrAnswer(err.Error())
				enc.Encode(ans)
				log(remote, req, ans)
				return
			}
			ans := ntv.Success("ok")
			enc.Encode(ans)
			enc.Encode(v)
			log(remote, req, ans)
		}()
		return
	case jdh.List:
		done.Add(1)
		go func() {
			defer conn.Close()
			defer done.Done()
			l, err := srv.db.List(table, req.Kvs)
			if err != nil {
				ans := ntv.ErrAnswer(err.Error())
				enc.Encode(ans)
				log(remote, req, ans)
				return
			}
			ans := ntv.Success("ok")
			enc.Encode(ans)
			for e := l.Front(); e != nil; e = e.Next() {
				enc.Encode(e.Value)
			}
			log(remote, req, ans)
		}()
		return
	case jdh.Set:
		var ans *ntv.Answer
		if remote == "127.0.0.1" {
			if len(req.Kvs) == 0 {
				ans = ntv.ErrAnswer("expecting arguments")
			} else {
				if err := srv.db.Set(table, req.Kvs); err != nil {
					ans = ntv.ErrAnswer(err.Error())
				} else {
					ans = ntv.Success("ok")
				}
			}
		} else {
			ans = ntv.ErrAnswer("forbidden")
		}
		enc.Encode(ans)
		log(remote, req, ans)
	default:
		ans := ntv.ErrAnswer("not implemented")
		enc.Encode(ans)
		log(remote, req, ans)
	}
	conn.Close()
}

func log(remote string, req *ntv.Request, ans *ntv.Answer) {
	msg := remote
	msg += fmt.Sprintf(" [%s]", time.Now().Format("2006-01-02 15:04:05 -0700"))
	msg += fmt.Sprintf(" \"%s", req.Query)
	if len(req.Table) > 0 {
		msg += fmt.Sprintf(" %s", req.Table)
	}
	for _, kv := range req.Kvs {
		msg += fmt.Sprintf(" %s=%s", kv.Key, kv.Value)
	}
	msg += `"`
	msg += fmt.Sprintf(" %s", ans.Message)
	fmt.Fprintf(os.Stdout, "%s\n", msg)
}
