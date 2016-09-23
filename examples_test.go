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
