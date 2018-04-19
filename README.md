gtss
====

[![GoDoc](https://godoc.org/github.com/logrusorgru/gtss?status.svg)](https://godoc.org/github.com/logrusorgru/gtss)
[![WTFPL License](https://img.shields.io/badge/license-wtfpl-blue.svg)](http://www.wtfpl.net/about/)
[![Build Status](https://travis-ci.org/logrusorgru/gtss.svg)](https://travis-ci.org/logrusorgru/gtss)
[![Coverage Status](https://coveralls.io/repos/logrusorgru/gtss/badge.svg?branch=master)](https://coveralls.io/r/logrusorgru/gtss?branch=master)
[![GoReportCard](https://goreportcard.com/badge/logrusorgru/gtss)](https://goreportcard.com/report/logrusorgru/gtss)
[![Gitter](https://img.shields.io/badge/chat-on_gitter-46bc99.svg?logo=data:image%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGhlaWdodD0iMTQiIHdpZHRoPSIxNCI%2BPGcgZmlsbD0iI2ZmZiI%2BPHJlY3QgeD0iMCIgeT0iMyIgd2lkdGg9IjEiIGhlaWdodD0iNSIvPjxyZWN0IHg9IjIiIHk9IjQiIHdpZHRoPSIxIiBoZWlnaHQ9IjciLz48cmVjdCB4PSI0IiB5PSI0IiB3aWR0aD0iMSIgaGVpZ2h0PSI3Ii8%2BPHJlY3QgeD0iNiIgeT0iNCIgd2lkdGg9IjEiIGhlaWdodD0iNCIvPjwvZz48L3N2Zz4%3D&logoWidth=10)](https://gitter.im/logrusorgru/gtss)

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


### Licensing

Copyright &copy; 2016 Konstantin Ivanov <kostyarin.ivanov@gmail.com>  
This work is free. You can redistribute it and/or modify it under the
terms of the Do What The Fuck You Want To Public License, Version 2,
as published by Sam Hocevar. See the LICENSE file for more details.
