//
// Copyright (c) 2016 Konstantin Ivanov <kostyarin.ivanov@gmail.com>.
// All rights reserved. This program is free software. It comes without
// any warranty, to the extent permitted by applicable law. You can
// redistribute it and/or modify it under the terms of the Do What
// The Fuck You Want To Public License, Version 2, as published by
// Sam Hocevar. See LICENSE file for more details or see below.
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
	"os"
	"os/signal"
)

func ExampleGrace() {
	var s *Server

	// initialize server

	var g Grace
	g.ListenAndServe(s) // start server in separate goroutine

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt) // subscribe to INT signal

	select {
	case <-sig: // got signal INT, exiting
		g.Close()
		<-g.Done()
	case <-g.Done(): // server failed
	}
	if err := g.Err(); err != nil {
		// handle error
	}
}

func ExampleServer() {
	s := &Server{
		Net:             "tcp4",
		Addr:            "127.0.0.1:9000",
		WorkersLimit:    1000,
		WriteBufferSize: No,
		ReadBufferSize:  8192,
		Handlers: []Handler{
			func(ctx *Context) {
				// prepare
			},
			func(ctx *Context) {
				// stuff
			},
			func(ctx *Context) {
				// finialize
			},
		},
	}
	s.ListenAndServe()
}
