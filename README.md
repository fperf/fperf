# Framework of performance testing 
[![Build Status](https://travis-ci.org/fperf/fperf.svg?branch=master)](https://travis-ci.org/fperf/fperf)
[![Go Report Card](https://goreportcard.com/badge/github.com/shafreeck/fperf)](https://goreportcard.com/report/github.com/shafreeck/fperf)

fperf is a powerful and flexible framework which allows you to develop your own benchmark tools so much easy.
**You create the client and send requests, fperf do the concurrency and statistics, then give you a report about qps and latency.**
Any one can create powerful performance benchmark tools by fperf with only some knowledge about how to send a request.

## Build fperf with the builtin clients
```sh
go install ./bin/fperf 
```

or use fperf-build
```
go install ./bin/fperf-build

fperf-build ./clients/*
```

## Quick Start
If you can not wait to run `fperf` to see how it works, follow the [quickstart](docs/quickstart.md)
 here.

## Customize client 
You can build your own client based on fperf framework. A client in fact is a client that
implement the fperf.Client or to say more precisely fperf.UnaryClient or fperf.StreamClient.

An unary client is a client to send requests. It works in request-reply model. For example,
HTTP benchmark client is an unary client. See [http client](clients/http/httpclient.go).
```go
type Client interface {
        Dial(addr string) error
}
type UnaryClient interface {
        Client
        Request() error
}
```

A stream client is a client to send and receive data by stream or datagram. TCP and UDP nomarlly
can be implemented as stream client. Google's grpc has a stream mode and can be used as a stream
client. See [grpc_testing](client/grpc_testing_client.go)
```go
type StreamClient interface {
	Client
	CreateStream(ctx context.Context) (Stream, error)
}
type Stream interface {
	DoSend() error
	DoRecv() error
}
```

### Three steps to create your own client
1.Create the "NewClient" function

```go
package demo

import (
	"fmt"
	"github.com/shafreeck/fperf"
	"time"
)

type demoClient struct{}

func newDemoClient(flag *fperf.FlagSet) fperf.Client {
	return &demoClient{}
}
```

2.Implement the UnaryClient or StreamClient
```go
func (c *demoClient) Dial(addr string) error {
	fmt.Println("Dial to", addr)
	return nil
}

func (c *demoClient) Request() error {
	time.Sleep(100 * time.Millisecond)
	return nil
}
```

3.Register to fperf
```go
func init() {
	fperf.Register("demo", dewDemoClient, "This is a demo client discription")
}
```

### Building custom clients
You client should be in the same workspace(same $GOPATH) with fperf.

#### Using fperf-build

`fperf-build` is a tool to build custom clients. It accepts a path of your package and
create file `autoimport.go` which imports all your clients when build fperf, then cleanup the
generated files after buiding.

Installing from source
```
go install ./bin/fperf-build
```
or  installing from github
```
go get github.com/shafreeck/fperf/bin/fperf-build
```

```shell
fperf-build [packages]
```

`packages` can be go importpath(see go help importpath) or absolute path to your package

For example, build all clients alang with fperf(using relative importpath)

```
fperf-build ./clients/* 
```

## Run benchmark
### Options
```
Usage: ./fperf [options] <client>
options:
  -N int
        number of request per goroutine
  -async
        send and recv in seperate goroutines
  -burst int
        burst a number of request, use with -async=true
  -connection int
        number of connection (default 1)
  -cpu int
        set the GOMAXPROCS, use go default if 0
  -delay duration
        wait delay time before send the next request
  -goroutine int
        number of goroutines per stream (default 1)
  -influxdb string
        writing stats to influxdb, specify the address in this option
  -recv
        perform recv action (default true)
  -send
        perform send action (default true)
  -server string
        address of the target server (default "127.0.0.1:8804")
  -stream int
        number of streams per connection (default 1)
  -tick duration
        interval between statistics (default 2s)
  -type string
        set the call type:unary, stream or auto. default is auto (default "auto")
clients:
 http   : HTTP performanch benchmark client
 mqtt-publish   : benchmark of mqtt publish
 redis  : redis performance benchmark
```

### Draw live graph with grafana

TODO export data into influxdb and draw graph with grafana
