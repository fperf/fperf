package main

import (
	"flag"
	"fmt"
	db "github.com/influxdb/influxdb/client/v2"
	"github.com/shafreeck/fperf/client"
	hist "github.com/shafreeck/fperf/stats"
	"golang.org/x/net/context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type setting struct {
	Connection     int
	Stream         int
	Goroutine      int
	Cpu            int
	Burst          int
	Tick           time.Duration
	Address        string
	Send           bool
	Recv           bool
	NoDelay        bool
	SendBufferSize int
	RecvBufferSize int
	Async          bool
	Target         string
	CallType       string
	InfluxDB       string
}

type statistics struct {
	latencies []time.Duration
	histogram *hist.Histogram
}

type roundtrip struct {
	start time.Time
	end   time.Time
	ack   bool
}

func dial(address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}

	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetNoDelay(s.NoDelay)
	if s.RecvBufferSize > 0 {
		tcpConn.SetReadBuffer(s.RecvBufferSize)
	}
	if s.SendBufferSize > 0 {
		tcpConn.SetWriteBuffer(s.SendBufferSize)
	}
	return conn, nil
}

func createClients(n int, addr string) []client.Client {
	clients := make([]client.Client, n)
	for i := 0; i < n; i++ {
		cli := client.NewClient(s.Target)
		if cli == nil {
			log.Fatalf("Can not find client %q for benchmark\n", s.Target)
		}

		cli.Dial(addr)
		clients[i] = cli
	}
	return clients
}
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

func benchmarkStream(n int, streams []client.Stream) {
	var wg sync.WaitGroup
	for _, stream := range streams {
		for i := 0; i < n; i++ {
			if s.Async {
				wg.Add(2)
				go send(nil, stream)
				go recv(nil, stream)
			} else {
				wg.Add(1)
				go run(nil, stream)
			}
		}
	}
	go statPrint()
	wg.Wait()
}
func benchmarkUnary(n int, clients []client.Client) {
	var wg sync.WaitGroup
	for _, cli := range clients {
		for i := 0; i < n; i++ {
			wg.Add(1)
			if cli, ok := cli.(client.UnaryClient); ok {
				go runUnary(nil, cli)
			} else {
				log.Fatalln(s.Target, " does not implement the client.UnaryClient")
			}
		}
	}
	go statPrint()
	wg.Wait()
}

func runUnary(done <-chan int, cli client.UnaryClient) {
	for {
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
		//	}
	}
}
func run(done <-chan int, stream client.Stream) {
	for {
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
		}
	}
}

func send(done <-chan int, stream client.Stream) {
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-done:
			log.Println("send goroutine exit, done")
			return
		default:
			select {
			case rtts <- &roundtrip{start: time.Now(), ack: false}:
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
		}
	}
}
func recv(done <-chan int, stream client.Stream) {
	timer := time.NewTimer(time.Second)
	for {
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
				rtt.ack = true
				eplase := time.Since(rtt.start)
				stats.latencies = append(stats.latencies, eplase)
				stats.histogram.Add(int64(eplase))
			case <-timer.C:
				log.Println("blocked on recv rtts")
			}
		}
	}
}

func statPrint() {
	tickc := time.Tick(s.Tick)
	var latencies []time.Duration
	for {
		select {
		case <-tickc:
			latencies = stats.latencies
			stats.latencies = stats.latencies[:0]

			sum := time.Duration(0)
			for _, eplase := range latencies {
				sum += eplase
			}
			count := len(latencies)
			if count != 0 {
				log.Printf("latency %v qps %d\n", sum/time.Duration(count), int64(float64(count)/float64(s.Tick)*float64(time.Second)))
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

func IdGen(min, max uint32) func() uint32 {
	id := min
	return func() uint32 {
		atomic.AddUint32(&id, 1)
		return id%max + min
	}
}

var s setting
var stats statistics
var rtts = make(chan *roundtrip, 10*1024*1024)
var mutex sync.RWMutex
var ider = IdGen(0, 10*1024*1024-1)
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
	flag.IntVar(&s.Cpu, "cpu", 0, "set the GOMAXPROCS, use go default if 0")
	flag.IntVar(&s.Burst, "burst", 0, "burst a number of request, use with -async=true")
	flag.BoolVar(&s.Send, "send", true, "perform send action")
	flag.BoolVar(&s.Recv, "recv", true, "perform recv action")
	flag.BoolVar(&s.NoDelay, "nodelay", true, "nodelay means sending requests ASAP")
	flag.DurationVar(&s.Tick, "tick", 2*time.Second, "interval between statistics")
	flag.StringVar(&s.Address, "server", "127.0.0.1:8804", "address of the target server")
	flag.IntVar(&s.SendBufferSize, "sndbuf", 0, "send buffer size")
	flag.IntVar(&s.RecvBufferSize, "rcvbuf", 0, "recv buffer size")
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

	runtime.GOMAXPROCS(s.Cpu)
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
			log.Fatalf("Error creating InfluxDB Client: ", err.Error())
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
