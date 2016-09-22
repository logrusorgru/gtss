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

// Package gtss implements golang TCP server with minimal
// features such as: buffering, workers limitation, etc
// inclusive gracefull shutdown
package gtss

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/netutil"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

const debug bool = false

func debugf(format string, args ...interface{}) {
	if debug {
		log.Printf("[debug]: "+format, args...)
	}
}

// internal defaults

const (
	defaultWorkersLimit = 1024 * 1024 * 1
	defaultNet          = "tcp"
	defaultAddr         = "0.0.0.0:3000"

	minTempDelay time.Duration = 5 * time.Millisecond
	maxTempDelay time.Duration = 1 * time.Second
)

// global service/convience constants

const (
	No      int = -1 // avoid options
	Default int = 0  // default value
)

// context is a wrapped connection

// A Context represents buffered or not buffered
// connection.
type Context struct {
	in   io.Reader
	out  io.Writer
	bin  *bufio.Reader
	bout *bufio.Writer
	kv   map[interface{}]interface{}
	net.Conn
}

// Read wraps connection Read method. It refers to buffer
// if connection is buffered
func (c *Context) Read(p []byte) (n int, err error) {
	debugf("(*Context).Read: %v", c.RemoteAddr())
	return c.in.Read(p)
}

// Write wraps connection Write method. It refers to buffer
// if connection is buffered
func (c *Context) Write(p []byte) (n int, err error) {
	debugf("(*Context).Read: %v", c.RemoteAddr())
	return c.out.Write(p)
}

// Get underlying net.Conn interface. It may be *net.TCPConn
// or *tls.Conn if TLS is used
func (c *Context) Connection() net.Conn {
	debugf("(*Context).Connection: %v", c.RemoteAddr())
	return c.Conn
}

// Flush write buffer or do nothig
func (c *Context) Flush() (err error) {
	debugf("(*Context).Flush: %v", c.RemoteAddr())
	if bout := c.bout; bout != nil {
		err = bout.Flush()
	}
	return
}

// Flush write buffer (if buffered) and close connection. If some error
// occured during flush then this error returns (actually, connection will
// be closed anyway). Otherwise closing error returns if any
func (c *Context) Close() (err error) {
	debugf("(*Context).close: %v", c.RemoteAddr())
	if bout := c.bout; bout != nil {
		if err = bout.Flush(); err != nil {
			c.Conn.Close() // drop second error
			return
		}
	}
	return c.Conn.Close()
}

// Set any context value associated with given key. The value is alive
// while connection is alive. It's possible to delete value using Del
// method. The provided key must be comparable
func (c *Context) Set(key, value interface{}) {
	debugf("(*Context).set: %v; %v=%v", c.RemoteAddr(), key, value)
	if c.kv == nil {
		c.kv = make(map[interface{}]interface{})
	}
	c.kv[key] = value
}

// Get return stored value by given key. The provided key must be
// comparable
func (c *Context) Get(key interface{}) interface{} {
	debugf("(*Context).get: %v; %v", c.RemoteAddr(), key)
	if c.kv == nil {
		return nil
	}
	return c.kv[key]
}

// Del deletes stored value by given key. The provided key must be
// comparable
func (c *Context) Del(key interface{}) {
	debugf("(*Context).del: %v; %v", c.RemoteAddr(), key)
	if c.kv == nil {
		return
	}
	delete(c.kv, key)
}

// reset context to store it inside pool
func (c *Context) reset() {
	debugf("(*Context).reset")
	if c.bin != nil {
		c.bin.Reset(nil)
	}
	if c.bout != nil {
		c.bout.Reset(nil)
	}
	c.kv = nil
	c.Conn = nil
}

// A Handler implements a connection handler. It's possible to use
// many handlers one by one (such as prepare-stuff-finialize). Feel
// free to use context Set, Get and Del methods to share some values
// between Handlers when connection is alive
type Handler func(ctx *Context)

type Server struct {
	// Net is "tcp", "tcp4" or "tcp6", defaults to "tcp"
	Net string
	// Addr is TCP address to listen on, "0.0.0.0:3000" if empty
	Addr string
	// Handlers are successive handlers to invoke
	Handlers []Handler
	// WorkersLimit is a maximum number of simultaneous connections.
	// Use No to avoid limitation. Use Default to set default limit.
	// The limit must not be nagative (except No (-1))
	WorkersLimit int
	// ReadBufferSize. By default a connection is buffered with
	// default buffer size. Use No to avoid buffering. Provide any
	// positive integer value to set particular size. All connections
	// will have the same buffer size. Feel free to use Defualt for
	// readability of your code
	ReadBufferSize int
	// WriteBufferSize. By default a connection is buffered with
	// default buffer size. Use No to avoid buffering. Provide any
	// positive integer value to set particular size. All connections
	// will have the same buffer size. Feel free to use Defualt for
	// readability of your code.
	WriteBufferSize int
	// TLSConfig is optional TLS config, used by ListenAndServeTLS
	TLSConfig *tls.Config
	// ErrorLog specifies an optional logger for errors accepting
	// connections and unexpected behavior from handlers.
	// If nil, logging goes to os.Stderr via the log package's
	// standard logger.
	ErrorLog *log.Logger

	// avoid alloc/GC pressure if many short-lived buffered connections
	// are coming
	ctxPool sync.Pool // one pool per server (because of buffers sizes)
}

// used if not nil (for tests)
var testHookServerServe func(s *Server, l net.Listener)

// obtain context from pool or create a new one
func (s *Server) getContext() *Context {
	debugf("(*Server).getContext")
	if ifc := s.ctxPool.Get(); ifc != nil {
		return ifc.(*Context)
	}
	return new(Context)
}

// put context to pool
func (s *Server) putContext(ctx *Context) {
	debugf("(*Server).putContext")
	ctx.reset()
	s.ctxPool.Put(ctx)
}

// create context by connection and buffers sizes
func (s *Server) createContext(conn net.Conn, rbs, wbs int) (ctx *Context) {
	debugf("(*Server).createContext")
	ctx = s.getContext()
	ctx.Conn = conn
	// set up reader
	switch rbs {
	case No: // -1
		ctx.in = conn
	case Default: // 0
		// create new bufio.Reader
		if ctx.bin == nil {
			ctx.bin = bufio.NewReader(conn)
		} else { // use existed
			ctx.bin.Reset(conn)
		}
		ctx.in = ctx.bin
	default: // > 0
		// with particular size
		if ctx.bin == nil {
			ctx.bin = bufio.NewReaderSize(conn, rbs)
		} else { // use existed
			ctx.bin.Reset(conn)
		}
		ctx.in = ctx.bin
	}
	// set up writer
	switch wbs {
	case No: // -1
		ctx.out = conn
	case Default: // 0
		// create new bufio.Writer
		if ctx.bout == nil {
			ctx.bout = bufio.NewWriter(conn)
		} else {
			// use existed
			ctx.bout.Reset(conn)
		}
		ctx.out = ctx.bout
	default: // > 0
		// create new bufio.Writer
		if ctx.bout == nil {
			ctx.bout = bufio.NewWriterSize(conn, wbs)
		} else {
			// use existed
			ctx.bout.Reset(conn)
		}
		ctx.out = ctx.bout
	}
	return
}

// wrap listener with LimitListener if need
func (s *Server) limitWorkes(l net.Listener) (ll net.Listener, err error) {
	debugf("(*Server).limitWorkers")
	switch wl := s.WorkersLimit; {
	case wl == Default: // == 0
		wl = defaultWorkersLimit
		fallthrough
	case wl > 0: // > 0
		ll = netutil.LimitListener(l, s.WorkersLimit)
	case wl == No: // == -1
		ll = l // do nothing
	default: // < -1
		err = fmt.Errorf("negative (*Server).WorkersLimit: %d", s.WorkersLimit)
	}
	return
}

// log errors
func (s *Server) logf(format string, args ...interface{}) {
	debugf("(*Server).logf")
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// test buffer sizes
func (s *Server) bufferSizes() (rbs, wbs int, err error) {
	debugf("(*Server).bufferSizes")
	if s.ReadBufferSize < No {
		err = fmt.Errorf("negative (*Server).ReadBufferSize: %d",
			s.ReadBufferSize)
	}
	if s.WriteBufferSize < No {
		err = fmt.Errorf("negative (*Server).WriteBufferSize: %d",
			s.WriteBufferSize)
	}
	rbs = s.ReadBufferSize
	wbs = s.WriteBufferSize
	return
}

// address and network
func (s *Server) an() (a, n string) {
	debugf("(*Server).an")
	a, n = s.Addr, s.Net
	if a == "" {
		a = defaultAddr
	}
	if n == "" {
		n = defaultNet
	}
	return
}

func (s *Server) listen() (l net.Listener, err error) {
	debugf("(*Server).listen")
	a, n := s.an()
	l, err = net.Listen(n, a)
	return
}

// ListenAndServe listens on the TCP network (*Server).Net and address
// (*Server).Addr and then calls Serve to handle incoming connections.
// If (*Server).Net is blank, "tcp" is used. If (*Server).Addr is blank,
// "0.0.0.0:3000" is used. ListenAndServe always returns a non-nil error
func (s *Server) ListenAndServe() error {
	debugf("(*Server).ListenAndServe")
	l, err := s.listen()
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	debugf("cloneTLSConfig")
	if cfg == nil {
		return &tls.Config{}
	}
	return &tls.Config{
		Rand:                        cfg.Rand,
		Time:                        cfg.Time,
		Certificates:                cfg.Certificates,
		NameToCertificate:           cfg.NameToCertificate,
		GetCertificate:              cfg.GetCertificate,
		RootCAs:                     cfg.RootCAs,
		NextProtos:                  cfg.NextProtos,
		ServerName:                  cfg.ServerName,
		ClientAuth:                  cfg.ClientAuth,
		ClientCAs:                   cfg.ClientCAs,
		InsecureSkipVerify:          cfg.InsecureSkipVerify,
		CipherSuites:                cfg.CipherSuites,
		PreferServerCipherSuites:    cfg.PreferServerCipherSuites,
		SessionTicketsDisabled:      cfg.SessionTicketsDisabled,
		SessionTicketKey:            cfg.SessionTicketKey,
		ClientSessionCache:          cfg.ClientSessionCache,
		MinVersion:                  cfg.MinVersion,
		MaxVersion:                  cfg.MaxVersion,
		CurvePreferences:            cfg.CurvePreferences,
		DynamicRecordSizingDisabled: cfg.DynamicRecordSizingDisabled,
		Renegotiation:               cfg.Renegotiation,
	}
}

func (s *Server) listenTLS(certFile, keyFile string) (l net.Listener,
	err error) {
	debugf("(*Server).listenTLS")
	a, n := s.an()
	config := cloneTLSConfig(s.TLSConfig)
	configHasCert := len(config.Certificates) > 0 ||
		config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return
		}
	}
	l, err = tls.Listen(n, a, config)
	return
}

// ListenAndServeTLS listens on the TCP network (*Server).Net and address
// (*Server).Addr and then calls Serve to handle incoming TLS connections.
// ListenAndServeTLS always returns a non-nil error
//
// Filenames containing a certificate and matching private key for the server
// must be provided if neither the Server's TLSConfig.Certificates nor
// TLSConfig.GetCertificate are populated
//
// If (*Server).Net is blank, "tcp" is used. If (*Server).Addr is blank,
// "0.0.0.0:3000" is used. ListenAndServe always returns a non-nil error
//
// ListenAndServeTLS always returns a non-nil error.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	debugf("(*Server).ListenAndServeTLS")
	l, err := s.listenTLS(certFile, keyFile)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

// Serve accepts incoming connections on the Listener l, creating a new service
// goroutine for each. The service goroutines call (*Server).Handlers one
// by one to reply to them
//
// Serve always returns a non-nil error
func (s *Server) Serve(l net.Listener) (err error) {
	debugf("(*Server).Serve")
	// close the Listener after all
	defer l.Close()
	// invoke hook for tests
	if testHookServerServe != nil {
		testHookServerServe(s, l)
	}
	// test buffer size and store it into local scope
	var rbs, wbs int
	if rbs, wbs, err = s.bufferSizes(); err != nil {
		return
	}
	// set workers limit
	if l, err = s.limitWorkes(l); err != nil {
		return
	}
	// how long to sleep on accept failure
	var tempDelay time.Duration
	// accept loop
	debugf("(*Server).Serve start accept loop")
	for {
		conn, e := l.Accept()
		debugf("(*Server).Serve Accept")
		if e != nil {
			debugf("(*Server).Serve Accept returns an error: %v", e)
			// if it's temporary error (not a real error)
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = minTempDelay
				} else {
					tempDelay *= 2
				}
				if tempDelay > maxTempDelay {
					tempDelay = maxTempDelay
				}
				s.logf("(*Server).Serve Accept error: %v; retrying in %v", e,
					tempDelay)
				time.Sleep(tempDelay) // await
				continue              // try again
			}
			return e
		}
		tempDelay = 0
		debugf("(*Server).Serve accept connection")
		go s.serve(conn, rbs, wbs)
	}
	return
}

func (s *Server) serve(conn net.Conn, rbs, wbs int) {
	debugf("(*Server).serve")
	// create context
	ctx := s.createContext(conn, rbs, wbs)
	// finialize
	defer func() {
		// handle Handers' panics
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			s.logf("panic serving %v: %v\n%s", ctx.RemoteAddr(), err, buf)
		}
		// close connection
		if err := ctx.Close(); err != nil {
			s.logf("error closing connection: %v", err)
		}
		// reset context and put it into the pool
		s.putContext(ctx)
	}()
	// invoke handlers one by one
	for _, h := range s.Handlers {
		h(ctx)
	}
}

// // Copy connection data back
// func EchoHandler(ctx *Context) {
// 	debugf("EchoHandler")
// 	_, err := io.Copy(ctx, ctx)
// 	if err != nil {
// 		log.Printf("EchoHandler error:", err)
// 	}
// }

// A Grace wraps server to provide gracefull shutdown
type Grace struct {
	closed chan struct{}
	done   chan struct{}
	once   *sync.Once
	err    error
	l      net.Listener
}

func (g *Grace) prepare() {
	debugf("(*Grace).prepare")
	g.closed = make(chan struct{})
	g.once = new(sync.Once)
	g.err = nil
}

// Done is closed when server is closed
func (g *Grace) Done() <-chan struct{} {
	debugf("(*Grace).Done")
	return g.done
}

// LsiternAndServe in separate gorotine. It panics if 's' is nil
func (g *Grace) ListenAndServe(s *Server) {
	debugf("(*Grace).ListenAndServe")
	g.done = make(chan struct{})
	l, err := s.listen()
	if err != nil {
		g.err = err
		close(g.done)
		return
	}
	g.Serve(s, l)
}

// LsiternAndServeTLS in separate gorotine. It panics if 's' is nil
func (g *Grace) ListenAndServeTLS(s *Server, certFile, keyFile string) {
	debugf("(*Grace).ListenAndServeTLS")
	g.done = make(chan struct{})
	l, err := s.listenTLS(certFile, keyFile)
	if err != nil {
		g.err = err
		close(g.done)
		return
	}
	g.Serve(s, l)
}

// Serve in separate gorotine. It panics if 's' is nil
func (g *Grace) Serve(s *Server, l net.Listener) {
	debugf("(*Grace).Serve")
	g.prepare()
	go func() {
		g.l = l
		err := s.Serve(l)
		debugf("(*Grace).Serve: (*Server).Serve returns")
		select {
		case <-g.closed:
		default:
			g.err = err
		}
		close(g.done)
	}()
}

// Close stops server
func (g *Grace) Close() {
	debugf("(*Grace).Close")
	if g.once != nil && g.l != nil {
		g.once.Do(func() {
			debugf("(*Grace).Close once.Do func")
			close(g.closed)
			g.l.Close()
		})
	}
}

// Err returns server error when it's closed
func (g *Grace) Err() error {
	debugf("(*Grace).Err")
	return g.err
}
