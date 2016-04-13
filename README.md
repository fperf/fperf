# Framework of performance testing 

This is a framework for common performance testing evolving from my another project `grpcperf`

## Quick Start
If you can not wait to run `fperf` to see how it works, you can follow the [quickstart](docs/quickstart.md)
 here.

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
