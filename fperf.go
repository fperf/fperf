/*
fperf allows you to build your performace tools easily

Three steps to create your own testcase

1. Create the "NewClient" function

	package demo

	import (
		"fmt"
		"github.com/shafreeck/fperf/client"
		"time"
	)

	type DemoClient struct{}

	func NewDemoClient(flag *client.FlagSet) client.Client {
		return &DemoClient{}
	}

2. Implement the UnaryClient or StreamClient

	func (c *DemoClient) Dial(addr string) error {
		fmt.Println("Dial to", addr)
		return nil
	}

	func (c *DemoClient) Request() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

3. Register to fperf

	func init() {
		client.Register("demo", NewDemoClient, "This is a demo client discription")
	}


Run the buildin testcase

http is a simple builtin testcase to benchmark http servers

	fperf -cpu 8 -connection 10 http http://example.com
*/
package main

import (
	"flag"
	"fmt"
	db "github.com/influxdata/influxdb/client/v2"
	"github.com/shafreeck/fperf/client"
	hist "github.com/shafreeck/fperf/stats"
	"golang.org/x/net/context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type setting struct {
	Connection int
	Stream     int
	Goroutine  int
	CPU        int
	Burst      int
	N          int //number of requests
	Tick       time.Duration
	Address    string
	Send       bool
	Recv       bool
	Delay      time.Duration
	Async      bool
	Target     string
	CallType   string
	InfluxDB   string
}

type statistics struct {
	latencies []time.Duration
	histogram *hist.Histogram
}

//roundtrip will be used in async mode
//the sender and receiver will be in seperate goroutines
type roundtrip struct {
	start time.Time
	end   time.Time
}

//create the testcase clients, n is the number of clients, set by
//flag -connection
func createClients(n int, addr string) []client.Client {
	clients := make([]client.Client, n)
	for i := 0; i < n; i++ {
		cli := client.NewClient(s.Target)
		if cli == nil {
			log.Fatalf("Can not find client %q for benchmark\n", s.Target)
		}

		if err := cli.Dial(addr); err != nil {
			log.Fatalln(err)
		}
		clients[i] = cli
	}
	return clients
}

//create streams for every client. n is the number of streams per client
func createStreams(n int, clients []client.Client) []client.Stream {
	streams := make([]client.Stream, n*len(clients))
	for cur, cli := range clients {
		for i := 0; i < n; i++ {
			if cli, ok := cli.(client.StreamClient); ok {
				stream, err := cli.CreateStream(context.Background())
				if err != nil {
					log.Fatalf("StreamCall faile to create new stream, %v", err)
				}
				streams[cur*n+i] = stream
			} else {
				log.Fatalln(s.Target, " do not implement the client.StreamClient")
			}
		}
	}
	return streams
}

//run benchmark for stream clients, can be in async or sync mode
func benchmarkStream(n int, streams []client.Stream) {
	var wg sync.WaitGroup
	for _, stream := range streams {
		for i := 0; i < n; i++ {
			//Notice here. we must pass stream as a parameter because the varibale stream
			//would be changed after the goroutine created
			if s.Async {
				wg.Add(2)
				go func(stream client.Stream) { send(nil, stream); wg.Done() }(stream)
				go func(stream client.Stream) { recv(nil, stream); wg.Done() }(stream)
			} else {
				wg.Add(1)
				go func(stream client.Stream) { run(nil, stream); wg.Done() }(stream)
			}
		}
	}
	go statPrint()
	wg.Wait()
}

//run benchmark for unary clients
func benchmarkUnary(n int, clients []client.Client) {
	var wg sync.WaitGroup
	for _, cli := range clients {
		for i := 0; i < n; i++ {
			wg.Add(1)
			if cli, ok := cli.(client.UnaryClient); ok {
				go func(cli client.UnaryClient) { runUnary(nil, cli); wg.Done() }(cli)
			} else {
				log.Fatalln(s.Target, " does not implement the client.UnaryClient")
			}
		}
	}
	go statPrint()
	wg.Wait()
}

func runUnary(done <-chan int, cli client.UnaryClient) {
	for i := 0; s.N == 0 || i < s.N; i++ {
		//select {
		//case <-done:
		//	log.Println("run goroutine exit done")
		//	return
		//default:
		start := time.Now()
		if err := cli.Request(); err != nil {
			log.Println(err)
		}
		eplase := time.Since(start)
		stats.latencies = append(stats.latencies, eplase)
		stats.histogram.Add(int64(eplase))
		if s.Delay > 0 {
			time.Sleep(s.Delay)
		}
		//	}
	}
}
func run(done <-chan int, stream client.Stream) {
	for i := 0; s.N == 0 || i < s.N; i++ {
		select {
		case <-done:
			log.Println("run goroutine exit done")
			return
		default:
			start := time.Now()
			if s.Send {
				stream.DoSend()
			}
			if s.Recv {
				stream.DoRecv()
			}
			eplase := time.Since(start)
			stats.latencies = append(stats.latencies, eplase)
			stats.histogram.Add(int64(eplase))
			if s.Delay > 0 {
				time.Sleep(s.Delay)
			}
		}
	}
}

func send(done <-chan int, stream client.Stream) {
	timer := time.NewTimer(time.Second)
	for i := 0; s.N == 0 || i < s.N; i++ {
		select {
		case <-done:
			log.Println("send goroutine exit, done")
			return
		default:
			select {
			case rtts <- &roundtrip{start: time.Now()}:
				timer.Reset(time.Second)
			case <-timer.C:
				log.Println("blocked on send rtts")
			}

			if burst != nil {
				select {
				case burst <- 0:
					timer.Reset(time.Second)
				case <-timer.C:
					log.Println("blocked on send burst chan")
				}
			}

			stream.DoSend()
			if s.Delay > 0 {
				time.Sleep(s.Delay)
			}
		}
	}
}
func recv(done <-chan int, stream client.Stream) {
	timer := time.NewTimer(time.Second)
	for i := 0; s.N == 0 || i < s.N; i++ {
		select {
		case <-done:
			log.Println("recv goroutine exit, done")
			return
		default:
			if burst != nil {
				select {
				case <-burst:
					timer.Reset(time.Second)
				case <-timer.C:
					log.Println("blocked on recv burst chan")
				}
			}
			err := stream.DoRecv()
			if err != nil {
				log.Println("recv goroutine exit", err)
				return
			}
			select {
			case rtt := <-rtts:
				timer.Reset(time.Second)
				eplase := time.Since(rtt.start)
				stats.latencies = append(stats.latencies, eplase)
				stats.histogram.Add(int64(eplase))
			case <-timer.C:
				log.Println("blocked on recv rtts")
			}
			if s.Delay > 0 {
				time.Sleep(s.Delay)
			}
		}
	}
}

func statPrint() {
	tickc := time.Tick(s.Tick)
	var latencies []time.Duration
	total := int64(0)
	for {
		select {
		case <-tickc:
			latencies = stats.latencies
			stats.latencies = stats.latencies[:0]

			sum := time.Duration(0)
			for _, eplase := range latencies {
				total++
				sum += eplase
			}
			count := len(latencies)
			if count != 0 {
				log.Printf("latency %v qps %d total %v\n", sum/time.Duration(count), int64(float64(count)/float64(s.Tick)*float64(time.Second)), total)
				if influxdb != nil {
					bp, _ := db.NewBatchPoints(db.BatchPointsConfig{
						Database:  "fperf",
						Precision: "s",
					})
					tags := map[string]string{"latency": "latency", "qps": "qps"}
					fields := map[string]interface{}{
						"latency": float64(sum) / float64(count) / 1000.0,
						"qps":     int64(float64(count) / float64(s.Tick) * float64(time.Second)),
					}
					pt, err := db.NewPoint("benchmark", tags, fields, time.Now())
					if err != nil {
						log.Println("Error: ", err.Error())
					}
					bp.AddPoint(pt)

					// Write the batch
					err = influxdb.Write(bp)
					if err != nil {
						log.Println("Error: ", err.Error())
					}
				}
			} else {
				log.Printf("blocking...")
			}
		}
	}
}

var s setting
var stats statistics
var rtts = make(chan *roundtrip, 10*1024*1024)
var mutex sync.RWMutex
var burst chan int
var influxdb db.Client

func usage() {
	fmt.Printf("Usage: %v [options] <client>\noptions:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Println("clients:")
	for name, desc := range client.AllClients() {
		fmt.Printf(" %s", name)
		if len(desc) > 0 {
			fmt.Printf("\t: %s", desc)
		}
		fmt.Println()
	}
}

func main() {
	flag.IntVar(&s.Connection, "connection", 1, "number of connection")
	flag.IntVar(&s.Stream, "stream", 1, "number of streams per connection")
	flag.IntVar(&s.Goroutine, "goroutine", 1, "number of goroutines per stream")
	flag.IntVar(&s.CPU, "cpu", 0, "set the GOMAXPROCS, use go default if 0")
	flag.IntVar(&s.Burst, "burst", 0, "burst a number of request, use with -async=true")
	flag.IntVar(&s.N, "N", 0, "number of request per goroutine")
	flag.BoolVar(&s.Send, "send", true, "perform send action")
	flag.BoolVar(&s.Recv, "recv", true, "perform recv action")
	flag.DurationVar(&s.Delay, "delay", 0, "wait delay time before send the next request")
	flag.DurationVar(&s.Tick, "tick", 2*time.Second, "interval between statistics")
	flag.StringVar(&s.Address, "server", "127.0.0.1:8804", "address of the target server")
	flag.BoolVar(&s.Async, "async", false, "send and recv in seperate goroutines")
	flag.StringVar(&s.CallType, "type", "auto", "set the call type:unary, stream or auto. default is auto")
	flag.StringVar(&s.InfluxDB, "influxdb", "", "writing stats to influxdb, specify the address in this option")
	flag.Usage = usage
	flag.Parse()

	s.Target = flag.Arg(0)
	if len(s.Target) == 0 {
		flag.Usage()
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-c
		stats.histogram.Value().Print(os.Stdout)
		os.Exit(0)
	}()

	runtime.GOMAXPROCS(s.CPU)
	go func() {
		runtime.SetBlockProfileRate(1)
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	if s.Burst > 0 {
		burst = make(chan int, s.Burst)
	}

	if len(s.InfluxDB) > 0 {
		c, err := db.NewHTTPClient(db.HTTPConfig{
			Addr: s.InfluxDB,
		})
		if err != nil {
			log.Fatalf("Error creating InfluxDB Client: %v", err.Error())
		}
		defer c.Close()
		q := db.NewQuery("CREATE DATABASE fperf", "", "")
		if response, err := c.Query(q); err == nil && response.Error() == nil {
			log.Println(response.Results)
		}
		influxdb = c
	}

	stats.latencies = make([]time.Duration, 0, 500000)
	histopt := hist.HistogramOptions{
		NumBuckets:         16,
		GrowthFactor:       1.8,
		SmallestBucketSize: 1000,
		MinValue:           10000,
	}
	stats.histogram = hist.NewHistogram(histopt)
	clients := createClients(s.Connection, s.Address)
	cli := clients[0]
	switch s.CallType {
	case "auto":
		switch cli.(type) {
		case client.StreamClient:
			streams := createStreams(s.Stream, clients)
			benchmarkStream(s.Goroutine, streams)
		case client.UnaryClient:
			benchmarkUnary(s.Goroutine, clients)
		}
	case "stream":
		streams := createStreams(s.Stream, clients)
		benchmarkStream(s.Goroutine, streams)
	case "unary":
		benchmarkUnary(s.Goroutine, clients)
	}
}
