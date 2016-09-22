package gtss

import (
	"os"
	"os/signal"
)

func Example_Grace() {
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
