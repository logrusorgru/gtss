gtss
====

[![GoDoc](https://godoc.org/github.com/logrusorgru/gtss?status.svg)](https://godoc.org/github.com/logrusorgru/gtss)
[![WTFPL License](https://img.shields.io/badge/license-wtfpl-blue.svg)](http://www.wtfpl.net/about/)
[![Build Status](https://travis-ci.org/logrusorgru/gtss.svg)](https://travis-ci.org/logrusorgru/gtss)
[![Coverage Status](https://coveralls.io/repos/logrusorgru/gtss/badge.svg?branch=master)](https://coveralls.io/r/logrusorgru/gtss?branch=master)
[![GoReportCard](https://goreportcard.com/badge/logrusorgru/gtss)](https://goreportcard.com/report/logrusorgru/gtss)
[![Gitter](https://img.shields.io/badge/chat-on_gitter-46bc99.svg?logo=data:image%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGhlaWdodD0iMTQiIHdpZHRoPSIxNCI%2BPGcgZmlsbD0iI2ZmZiI%2BPHJlY3QgeD0iMCIgeT0iMyIgd2lkdGg9IjEiIGhlaWdodD0iNSIvPjxyZWN0IHg9IjIiIHk9IjQiIHdpZHRoPSIxIiBoZWlnaHQ9IjciLz48cmVjdCB4PSI0IiB5PSI0IiB3aWR0aD0iMSIgaGVpZ2h0PSI3Ii8%2BPHJlY3QgeD0iNiIgeT0iNCIgd2lkdGg9IjEiIGhlaWdodD0iNCIvPjwvZz48L3N2Zz4%3D&logoWidth=10)](https://gitter.im/logrusorgru/gtss?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) | 
[![paypal gratuity](https://img.shields.io/badge/paypal-gratuity-3480a1.svg?logo=data:image%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMDAwIDEwMDAiPjxwYXRoIGZpbGw9InJnYigyMjAsMjIwLDIyMCkiIGQ9Ik04ODYuNiwzMDUuM2MtNDUuNywyMDMuMS0xODcsMzEwLjMtNDA5LjYsMzEwLjNoLTc0LjFsLTUxLjUsMzI2LjloLTYybC0zLjIsMjEuMWMtMi4xLDE0LDguNiwyNi40LDIyLjYsMjYuNGgxNTguNWMxOC44LDAsMzQuNy0xMy42LDM3LjctMzIuMmwxLjUtOGwyOS45LTE4OS4zbDEuOS0xMC4zYzIuOS0xOC42LDE4LjktMzIuMiwzNy43LTMyLjJoMjMuNWMxNTMuNSwwLDI3My43LTYyLjQsMzA4LjktMjQyLjdDOTIxLjYsNDA2LjgsOTE2LjcsMzQ4LjYsODg2LjYsMzA1LjN6Ii8%2BPHBhdGggZmlsbD0icmdiKDIyMCwyMjAsMjIwKSIgZD0iTTc5MS45LDgzLjlDNzQ2LjUsMzIuMiw2NjQuNCwxMCw1NTkuNSwxMEgyNTVjLTIxLjQsMC0zOS44LDE1LjUtNDMuMSwzNi44TDg1LDg1MWMtMi41LDE1LjksOS44LDMwLjIsMjUuOCwzMC4ySDI5OWw0Ny4zLTI5OS42bC0xLjUsOS40YzMuMi0yMS4zLDIxLjQtMzYuOCw0Mi45LTM2LjhINDc3YzE3NS41LDAsMzEzLTcxLjIsMzUzLjItMjc3LjVjMS4yLTYuMSwyLjMtMTIuMSwzLjEtMTcuOEM4NDUuMSwxODIuOCw4MzMuMiwxMzAuOCw3OTEuOSw4My45TDc5MS45LDgzLjl6Ii8%2BPC9zdmc%2B)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=AVFWLEREA97PU)

Golang TCP server skeleton

![gtss logo](https://github.com/logrusorgru/gtss/blob/master/gopher_gtss.png)

# Installation

Get
```
go get -u github.com/logrusorgru/gtss
```
Test
```
go test github.com/logrusorgru/gtss
```

# Features

+ TLS connections
+ Usege is similar to `net/http` package
+ Limit number of simultaneous connections
+ Buffered reading and buffered writing
+ Share values between handlers
+ Buffers pool

# Examples

```go
package main

import (
	"github.com/logrusorgru/gtss"

	"log"
)

func main() {
	s := gtss.Server{
		Addr: "127.0.0.1:9000",
		WorkersLimit: 100 * 1000
		ReadBufferSize: gtss.Default,
		WriteBufferSize: gtss.No,
		Handlers: []gtss.Handler{
			func(ctx *gtss.Context) {
				// prepare
			},
			func(ctx *gtss.Context) {
				// stuff
			},
			func(ctx *gtss.Context) {
				// finialize
			},
		},
	}

	s.ListenAndServe()

	/*
		Also
		====

		log.Print(s.ListenAndServeTLS(certFile, keyFile))

		Or (gracefull shutdown)
		=======================

		g := Grace{}
		g.ListenAndServe(s)

		// additional imports: 'os' and 'os/signal'
		sig := make(chan struct{}, 2)
		singal.Notify(sig, os.Interrupt)

		select {
		case <-sig:
			log.Print("got signal INT, exiting...")
			g.Close()
			<-g.Done
		case <-g.Done():
		}

		if err := g.Err(); err != nil {
			log.Print("server error: ", err)
		}

	*/
}

```

# Installation

```bash
go get github.com/logrusorgru/gtss
```


### Licensing

Copyright &copy; 2015 Konstantin Ivanov <ivanov.konstantin@logrus.org.ru>  
This work is free. You can redistribute it and/or modify it under the
terms of the Do What The Fuck You Want To Public License, Version 2,
as published by Sam Hocevar. See the LICENSE.md file for more details.
