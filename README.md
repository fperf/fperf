# Framework of performance testing 

fperf is a framework which allows you to develop your own benchmark tools has never been easier.

## Quick Start
If you can not wait to run `fperf` to see how it works, follow the [quickstart](docs/quickstart.md)
 here.

## Customize testcase
You can build your own testcase based on fperf framework. A testcase in fact is a client that
implement the client.Client or to say more precisely client.UnaryClient or client.StreamClient.

An unary client is a client to send requests. It works in request-reply model. For example,
HTTP benchmark client is an unary client. See [http_testing](client/http_testing_client.go).
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

### Three steps to create your own testcase
1.Create the "NewClient" function

```go
package demo

import (
	"fmt"
	"github.com/shafreeck/fperf/client"
	"time"
)

type demoClient struct{}

func newDemoClient(flag *client.FlagSet) client.Client {
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
	client.Register("demo", dewDemoClient, "This is a demo client discription")
}
```

### Build custom testcases
You testcase should be in the same workspace(same $GOPATH) with fperf. There are two ways to build the custom testcase.  
#### Import the testcase package explicitly

Edit fperf.go and insert
```go
import _ "github.com/shafreeck/fperf/example/demo"
```

#### Use buildtestcase.sh

`buildtestcase.sh` is a script to build custom testcases. It accepts a path of your package and
create file `autoimport.go` which imports all your testcases when build fperf, then cleanup the
generated files after buiding.

```shell
./buildtestcase.sh example/demo
```

## Run benchmark
### Options
* **-N : number of the requests issued per goroutine**
* **-async=false: send and recv in seperate goroutines**  
Only used in stream mode. Use two groutine per stream, one for sending and another for recving.
defualt is false.

* **-burst=0: burst a number of request, use with**

* **-async=true**  
Used with -async option. use birst to limit the counts of request the sending goroutine issued.
default is unlimited.

* **-connection=1: number of connection**  
Set the number of connection will be created.

* **-cpu=0: set the GOMAXPROCS, use go default if 0**  
Set the GOMAXPROCS, default is 0, which means use the golang default option.

* **-goroutine=1: number of goroutines per stream**  
Set the number of goroutines, should only be used when testing is unary mode or streaming mode
with only 1 stream. This is because the stream is not thread-safty

* **-influxdb="": writing stats to influxdb, specify the address in this option**  
Set the influxdb address, fperf will automatic create a fperf table and inserting 
into the qps and lantency metrics.

* **-delay=0: wait <delay> time before send the next request**  
Set the time to wait before send a request

* **-recv=true: perform recv action**  
Only used in stream mode. Just enable the recving goroutine.

* **-send=true: perform send action**
Only used in stream mode. Just enable the sending goroutine.

* **-server="127.0.0.1:8804": address of the remote server**  
Set the address of the target server

* **-stream=1: number of streams per connection**  
Set the number of stream will be created. Only being used in stream mode.

* **-tick=2s: interval between statistics**  
Set the interval time between output the qps and latency metrics

* **-type="auto": set the call type:unary, stream or auto. default is auto**  
Set the type of your testcase. This option can be used when your testcase implement
unary and stream client at the same time and in this case fperf can not judge the type
automaticlly

### Draw live graph with grafana

TODO export data into influxdb and draw graph with grafana
