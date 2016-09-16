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

// Package gtss implements TCP server skeleton with
// some features
package gtss

import (
	"bufio"                    // buffering
	"crypto/tls"               // encryption
	"errors"                   // report errors
	"golang.org/x/net/netutil" // limit connections
	"io"                       // Reader and Writer interfaces, Copy
	"log"                      // used for EchoHandler
	"net"                      // tcp server
	"sync"                     // sync primitives
)

const (
	defaultWorkersLimit = 1 * 1024 * 1024
	defaultNet          = "tcp"
	defaultAddress      = "0.0.0.0:3000"
)

const (
	No      int = -1 // no limit or no buffers
	Default int = 0  // default workers limit or default buffer size
)

// A Handler is any function that handle incoming connection
type Handler func(in io.Reader, out io.Writer, conn net.Conn)

// EchoHandler is default handler that copies input to output
func EchoHandler(in io.Reader, out io.Writer, conn net.Conn) {
	remote := conn.RemoteAddr()
	log.Println("incoming connection from:", remote)
	n, err := io.Copy(out, in)
	if err != nil && err != io.EOF {
		log.Printf("io.Copy from/to remote %s, %d bytes copied, error: %s",
			remote.String(),
			n,
			err.Error())
	} else {
		log.Printf("io.Copy from/to remote %s, %d bytes copied, successful",
			remote.String(), n)
	}
}

// A Server implements TCP server. It's safe to change Server values
// before start only
type Server struct {
	// internals

	l      net.Listener  // listener interface
	closed chan struct{} // listener is closed
	done   chan struct{} // listener is done

	Net       string      // "tcp", "tcp4" or "tcp6", defaults to "tcp"
	Addr      string      // address, defaults to "0.0.0.0:3000"
	TLSConfig *tls.Config // optional TLS config

	// Handler function (that handle connections). Defaults to
	// EchoHandler
	Handler Handler
	// WorkersLimit is limitation of simultaneous connections.
	// Use `Default' for default or `No' for unlimited
	WorkersLimit int
	// ReadBufferSize is size of reading buffer for each connection.
	// Use `Default' for default or `No' to avoid buffering
	ReadBufferSize int
	// WriteBufferSize is size of writing buffer for each connection.
	// Use `Default' for default or `No' to avoid buffering
	WriteBufferSize int
}

// ListenAndServe starts server
func (s *Server) ListenAndServe() (err error) {
	if err = s.prepare(); err != nil {
		return
	}
	// accept connections in another goroutine
	go s.listern()
	//
	return
}

func (s *Server) prepare() {
	// set default values where need
	if s.Net == "" {
		s.Net = defaultNet
	}
	if s.Addr == "" {
		s.Addr = defaultAddress
	}
	if s.Handler == nil {
		s.Handler = EchoHandler
	}
	if s.WorkersLimit == Default {
		s.WorkersLimit = defaultWorkersLimit
	}
	// check ReadBufferSize and WriteBufferSize
	if s.ReadBufferSize != No && s.ReadBufferSize < 0 {
		err = errors.New("negative size of reading buffer")
		return
	}
	if s.WriteBufferSize != No && s.WriteBufferSize < 0 {
		err = errors.New("negative size of writing buffer")
		return
	}
	// create listener
	var l net.Listener
	// use TLS if tls config is given, otherwise pure TCP
	if s.TLSConfig != nil {
		l, err = tls.Listen(s.Net, s.Addr, s.TLSConfig)
	} else {
		l, err = net.Listen(s.Net, s.Addr)
	}
	if err != nil {
		return
	}
	// set workers limit if need
	if s.WorkersLimit != No {
		if s.WorkersLimit < 0 {
			err = errors.New("negative workers limit") // avoid panic
			return
		}
		l = netutil.LimitListener(l, s.WorkersLimit) // set limit
	}
	// make internal channels
	s.closed = make(chan struct{})
	s.done = make(chan struct{})
	// store listener inside Server struct
	s.l = l
}

func (s *Server) listern() (err error) {
	var conn net.Conn
	var in io.Reader
	var out io.Writer
	// waiting group
	wg := new(sync.WaitGroup)
	// accept loop
	for {
		conn, err = s.l.Accept()
		if err != nil {
			select {
			case <-s.closed:
				// prevent "use of closed network connection" error
				// when listener is closed manually
				err = nil
			}
			break // break loop
		}
		// setup buffers
		in, out = s.buffered(conn)
		// handle connection in another goroutine
		wg.Add(1)
		go func(in io.Reader, out io.Writer,
			wg *sync.WaitGroup, conn net.Conn) {
			// group defered calls for performance reason
			defer func() {
				conn.Close() // close connection
				wg.Done()    // done
			}()
			s.Handler(in, out, conn)
		}(in, out, wg, conn)
		// next loop
	}
	wg.Wait() // waiting for all handlers
	return
}

// unbuffered is used to provide fake Flush method for unbuffered
// writing
type unbuffered struct {
	io.Writer
}

func (u *unbuffered) Flush() error {
	return nil
}

func (s *Server) buffered(conn net.Conn) (in io.Reader, out io.Writer) {
	// read buffer
	if s.ReadBufferSize != No {
		if s.ReadBufferSize == Default {
			in = bufio.NewReader(conn) // default buffer size (4096)
		} else {
			in = bufio.NewReaderSize(conn, s.ReadBufferSize)
		}
	} else {
		in = conn // not buffered reading
	}
	// write buffer
	if s.WriteBufferSize != No {
		if s.WriteBufferSize == Default {
			out = bufio.NewWriter(conn) // default buffer size (4096)
		} else {
			out = bufio.NewWriterSize(conn, s.WriteBufferSize)
		}
	} else {
		out = &unbuffered{conn} // not buffered writing
	}
	return
}

// stop server
func (s *Server) Close() (err error) {
	if s.l != nil {
		select {
		case <-s.closed:
		default:
			close(s.closed)
		}
		return s.l.Close()
	}
	return nil
}
