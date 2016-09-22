//
// Copyright (c) 2016 Konstanin Ivanov <kostyarin.ivanov@gmail.com>.
// All rights reserved. This program is free software. It comes without
// any warranty, to the extent permitted by applicable law. You can
// redistribute it and/or modify it under the terms of the Do What
// The Fuck You Want To Public License, Version 2, as published by
// Sam Hocevar. See LICENSE.md file for more details or see below.
//

//
//        DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
//                    Version 2, December 2004
//
// Copyright (C) 2004 Sam Hocevar <sam@hocevar.net>
//
// Everyone is permitted to copy and distribute verbatim or modified
// copies of this license document, and changing it is allowed as long
// as the name is changed.
//
//            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
//   TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION
//
//  0. You just DO WHAT THE FUCK YOU WANT TO.
//

package gtss

import (
	"testing"

	"crypto/tls"
	"io"
	"net"
)

const (
	listenOn = "127.0.0.1:0"
	maxTries = 5
)

// helper functions
func send(addr string, data []byte) (err error) {
	var c net.Conn
	if c, err = net.Dial("tcp", addr); err != nil {
		return
	}
	defer func() {
		if err != nil {
			c.Close()
		} else {
			err = c.Close()
		}
	}()
	_, err = c.Write(data)
	return
}

func sendTLS(addr string, data []byte) (err error) {
	var c net.Conn
	if c, err = tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	}); err != nil {
		return
	}
	defer func() {
		if err != nil {
			c.Close()
		} else {
			err = c.Close()
		}
	}()
	_, err = c.Write(data)
	return
}

// func sendRecv(addr string, data []byte, rn int) (reply []byte, err error) {
// 	var c net.Conn
// 	if c, err = net.Dial("tcp", addr); err != nil {
// 		return
// 	}
// 	defer func() {
// 		if err != nil {
// 			c.Close()
// 		} else {
// 			err = c.Close()
// 		}
// 	}()
// 	_, err = c.Write(data)
// 	reply = make([]byte, rn)
// 	_, err = io.ReadFull(c, reply)
// 	return
// }

func recv(addr string, rn int) (reply []byte, err error) {
	var c net.Conn
	if c, err = net.Dial("tcp", addr); err != nil {
		return
	}
	defer func() {
		if err != nil {
			c.Close()
		} else {
			err = c.Close()
		}
	}()
	reply = make([]byte, rn)
	_, err = io.ReadFull(c, reply)
	return
}

func open(addr string) (c net.Conn, err error) {
	c, err = net.Dial("tcp", addr)
	return
}

func hSend(data []byte, t *testing.T) Handler {
	return func(ctx *Context) {
		n, err := ctx.Write(data)
		if err != nil {
			t.Errorf("handler write error: %v", err)
		}
		if n != len(data) {
			t.Errorf("handler short white; expected %d, got %d", len(data), n)
		}
	}
}

func hRecv(want []byte, t *testing.T) Handler {
	return func(ctx *Context) {
		got := make([]byte, len(want))
		_, err := ctx.Read(got)
		if err != nil {
			t.Errorf("handler read error: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("wrong client response: expected %q, got %q", string(want),
				string(got))
		}
	}
}

func beforeServe() chan net.Listener {
	c := make(chan net.Listener, 1)
	testHookServerServe = func(_ *Server, l net.Listener) {
		c <- l
	}
	return c
}

func afterServe() { testHookServerServe = nil }

func Test_openConnection(t *testing.T) {
	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: No,
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	conn, err := open(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func TestContext_ConnectionSetGetDel(t *testing.T) {
	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: No,
			Handlers: []Handler{
				func(ctx *Context) {
					conn := ctx.Connection()
					if conn != ctx.Conn {
						t.Error("(*Context).Connection returns wrong")
					}
					ctx.Set(1, 1)
				},
				func(ctx *Context) {
					v := ctx.Get(1)
					if v == nil {
						t.Error("(*Context).Set-Get returns nil")
					}
					one, ok := v.(int)
					if !ok || one != 1 {
						t.Error("(*Context).Set-Get returns wrong")
					}
					ctx.Del(1)
				},
				func(ctx *Context) {
					if ctx.Get(1) != nil {
						t.Error("(*Context).Del did nothing")
					}
				},
			},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	conn, err := open(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func Test_clientToServer(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: No,
			Handlers:        []Handler{hRecv(data, t)},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	if err := send(ln.Addr().String(), data); err != nil {
		t.Fatal(err)
	}
}

func Test_serverToClient(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: No,
			Handlers:        []Handler{hSend(data, t)},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	reply, err := recv(ln.Addr().String(), len(data))
	if err != nil {
		t.Fatalf("client receiving error: %v", err)
	}
	if string(reply) != string(data) {
		t.Errorf("wrong msg from server: expected %q, got %q", string(data),
			string(reply))
	}
}

func Test_serverToClientBuffered(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: Default,
			Handlers:        []Handler{hSend(data, t)},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	reply, err := recv(ln.Addr().String(), len(data))
	if err != nil {
		t.Fatalf("client receiving error: %v", err)
	}
	if string(reply) != string(data) {
		t.Errorf("wrong msg from server: expected %q, got %q", string(data),
			string(reply))
	}
}

func Test_serverToClientBufferedFlushClose(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  No,
			WriteBufferSize: Default,
			Handlers: []Handler{
				func(ctx *Context) {
					n, err := ctx.Write(data)
					if err != nil {
						t.Errorf("handler write error: %v", err)
					}
					if n != len(data) {
						t.Errorf("handler short white; expected %d, got %d",
							len(data), n)
					}
					if err := ctx.Flush(); err != nil {
						t.Error("flush error:", err)
					}
					if err := ctx.Close(); err != nil {
						t.Error("close error:", err)
					}
				},
			},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	reply, err := recv(ln.Addr().String(), len(data))
	if err != nil {
		t.Fatalf("client receiving error: %v", err)
	}
	if string(reply) != string(data) {
		t.Errorf("wrong msg from server: expected %q, got %q", string(data),
			string(reply))
	}
}

func Test_clientToServerBuffered(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  Default,
			WriteBufferSize: No,
			Handlers:        []Handler{hRecv(data, t)},
		}
		errc := make(chan error, 1)
		go func() { errc <- s.ListenAndServe() }()
		select {
		case err := <-errc:
			t.Logf("On try #%v: %v", try+1, err)
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	defer ln.Close()
	if err := send(ln.Addr().String(), data); err != nil {
		t.Fatal(err)
	}
}

func Test_clientToServerGrace(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener
	var g Grace

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  Default,
			WriteBufferSize: No,
			Handlers:        []Handler{hRecv(data, t)},
		}
		g.ListenAndServe(s)
		select {
		case <-g.Done():
			t.Logf("On try #%v: %v", try+1, g.Err())
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	if err := send(ln.Addr().String(), data); err != nil {
		t.Fatal(err)
	}
	g.Close()
	<-g.Done()
	if g.Err() != nil {
		t.Errorf("server closing error: %v", g.Err())
	}
}

func Test_clientToServerGraceTLS(t *testing.T) {
	data := []byte("Hello")

	var ok bool
	var s *Server
	var ln net.Listener
	var g Grace

	serveNotify := beforeServe()
	defer afterServe()

Try:
	for try := 0; try < maxTries; try++ {
		s = &Server{
			Addr:            listenOn,
			WorkersLimit:    No,
			ReadBufferSize:  Default,
			WriteBufferSize: No,
			Handlers:        []Handler{hRecv(data, t)},
			TLSConfig:       tlsConfig(t),
		}
		g.ListenAndServeTLS(s, "", "")
		select {
		case <-g.Done():
			t.Logf("On try #%v: %v", try+1, g.Err())
			continue
		case ln = <-serveNotify:
			ok = true
			t.Logf("listening on: %v", ln.Addr())
			break Try
		}
	}
	if !ok {
		t.Fatalf("Failed to start up after %d tries", maxTries)
	}
	if err := sendTLS(ln.Addr().String(), data); err != nil {
		t.Fatal(err)
	}
	g.Close()
	<-g.Done()
	if g.Err() != nil {
		t.Errorf("server closing error: %v", g.Err())
	}
}

type dummyListener struct {
	count int
}

func (dummyListener) Accept() (net.Conn, error) { return nil, nil }
func (dummyListener) Close() error              { return nil }
func (dummyListener) Addr() net.Addr            { return nil }

func TestServer_limitWorkers(t *testing.T) {
	s := Server{}
	s.WorkersLimit = -2
	if _, err := s.limitWorkes(nil); err == nil {
		t.Error("missing error")
	}
	s.WorkersLimit = No
	d := net.Listener(&dummyListener{})
	l, err := s.limitWorkes(d)
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if d != l {
		t.Error("unknown hujnya")
	}
	s.WorkersLimit = Default
	l, err = s.limitWorkes(d)
	if err != nil {
		t.Error("unexpected error:", err)
	}
}

func tlsConfig(t *testing.T) *tls.Config {
	cert, err := tls.X509KeyPair([]byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`), []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
