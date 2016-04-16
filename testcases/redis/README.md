# Build redis performance tool using fperf

`fperf` is a powerful and flexible framework allows you to build your own
benchmark and load testing tool easily. All the knowledge you should to
know is how to send a request and fperf will handles the concurrency and 
statistics.

The is an example about how to build a redis benchmark tool with fperf in 
3 steps. You can alse build tools for other common servers like mysql, nginx, 
memcached or your own servers. Maybe building benchmark tools for private servers 
is the most cases you need fperf. In fact, fperf indeed is created to benchmark our grpc 
servers at first.

## Coding: 3 steps to build a redis benchmark tool

You can find the source code from [testcases/redis/redis.go](testcases/redis/redis.go)

### Create the NewClientFunc function
```go
//newRedisClient create the client object. The function should be
//registered to fperf, fperf -h will list all the registered clients(testcases)
func newRedisClient(flag *client.FlagSet) client.Client {
    c := new(redisClient)
    
    //A client can have itself options. fperf use flag to process
    //the command args and options
    flag.BoolVar(&c.options.verbose, "v", false, "verbose")

    //Customize the usage output
    flag.Usage = func() {
        fmt.Printf("Usage: redis [options] [cmd] [args...]\noptions:\n")
        flag.PrintDefaults()
    }
    flag.Parse()

    args := flag.Args()
    //Set the default command if not be set
    if len(args) == 0 {
        args = []string{"SET", "fperf", "hello world"}
    }
    c.args = args

    if c.options.verbose {
        fmt.Println(c.args)
    }
    return c
}
```

### Implement Dial and Request methods

```go
//Dial to redis server. The addr is set by the fperf option "-server"
func (c *redisClient) Dial(addr string) error {
    rds, err := redis.DialURL(addr)
    if err != nil {
        return err
    }
    c.rds = rds
    return nil
}

//Request send a redis request and return the error if there is
func (c *redisClient) Request() error {
    var args []interface{}

    //Build the redis cmd and args
    cmd := c.args[0]
    for _, arg := range c.args[1:] {
        args = append(args, arg)
    }
    _, err := c.rds.Do(cmd, args...)
    return err
}
```

### Register to fperf

```go
//Register to fperf
//parameters: name to register, NewClientFunc, a description
func init() {
	client.Register("redis", newRedisClient, "redis performance benchmark")
}
```

## Build and Run

### Build your testcase using "buildtestcase.sh"

```shell
./buildtestcase.sh testcases/redis/
```

### Run the benchmark

```shell
./fperf -server redis://localhost:6379 redis hset key field value
2016/04/17 00:04:35 latency 51.447µs qps 18671 total 37343
2016/04/17 00:04:37 latency 33.881µs qps 27777 total 92898
Count: 101781  Min: 16491  Max: 42468733  Avg: 43377.29
------------------------------------------------------------
[     10000,      11000)       0    0.0%    0.0%
[     11000,      13800)       0    0.0%    0.0%
[     13800,      21639)   62151   61.1%   61.1%  ######
[     21639,      43590)   12092   11.9%   72.9%  #
[     43590,     105055)   15539   15.3%   88.2%  ##
[    105055,     277158)   11941   11.7%   99.9%  #
[    277158,     759048)      43    0.0%  100.0%
```

See the [quick start](docs/quickstart.md) to learn about the output of fperf.
