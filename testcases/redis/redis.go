/*
This is an example about how to build a redis benchmark to using fperf

A fperf testcase in fact is an implementation of client.UnaryClient.
The client has two method:

	Dial(addr string) error
	Request() error

Dial connect to the server address witch set by fperf option "-server". fperf will exit and print
the error message if error occurs.

Request is the method to fperf uses to issue an request. The returned error would be printed and
fperf would continue.
*/
package redis

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/shafreeck/fperf/client"
)

var seq func() string = seqCreater(0)
var random func() string = randCreater(10000000000000000)

//A test case can have itself options witch would be passed by fperf
type options struct {
	verbose bool
}

//A client is a struct that should implement client.UnaryClient
type redisClient struct {
	args    []string   //the args of client, we use redis command as args
	rds     redis.Conn //the redis connection, should be created when call Dial
	options options    //the options user set
}

//newRedisClient create the client object. The function should be
//registered to fperf, fperf -h will list all the registered clients(testcases)
func newRedisClient(flag *client.FlagSet) client.Client {
	c := new(redisClient)
	flag.BoolVar(&c.options.verbose, "v", false, "verbose")

	//Customize the usage output
	flag.Usage = func() {
		fmt.Printf("Usage: redis [options] [cmd] [args...], use __rand_int__ or __seq_int__ to generate random or sequence keys\noptions:\n")
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

//Dial to redis server. The addr is set by the fperf option "-server"
func (c *redisClient) Dial(addr string) error {
	rds, err := redis.DialURL(addr)
	if err != nil {
		return err
	}
	c.rds = rds
	return nil
}

func seqCreater(begin int64) func() string {
	// filled map, filled generated to 16 bytes
	l := []string{
		"",
		"0",
		"00",
		"000",
		"0000",
		"00000",
		"000000",
		"0000000",
		"00000000",
		"000000000",
		"0000000000",
		"00000000000",
		"000000000000",
		"0000000000000",
		"00000000000000",
		"000000000000000",
	}
	v := begin
	m := &sync.Mutex{}
	return func() string {
		m.Lock()
		s := strconv.FormatInt(v, 10)
		v += 1
		m.Unlock()

		filled := len(l) - len(s)
		if filled <= 0 {
			return s
		}
		return l[filled] + s
	}
}

func randCreater(max int64) func() string {
	// filled map, filled generated to 16 bytes
	l := []string{
		"",
		"0",
		"00",
		"000",
		"0000",
		"00000",
		"000000",
		"0000000",
		"00000000",
		"000000000",
		"0000000000",
		"00000000000",
		"000000000000",
		"0000000000000",
		"00000000000000",
		"000000000000000",
	}
	var v int64
	m := &sync.Mutex{}
	return func() string {
		m.Lock()
		v = rand.Int63n(max)
		s := strconv.FormatInt(v, 10)
		m.Unlock()

		filled := len(l) - len(s)
		if filled <= 0 {
			return s
		}
		return l[filled] + s
	}
}

func replaceSeq(s string) string {
	return strings.Replace(s, "__seq_int__", seq(), -1)
}
func replaceRand(s string) string {
	return strings.Replace(s, "__rand_int__", random(), -1)
}

//Request send a redis request and return the error if there is
func (c *redisClient) Request() error {
	var args []interface{}

	//Build the redis cmd and args
	cmd := c.args[0]
	for _, arg := range c.args[1:] {
		if strings.Index(arg, "__seq_int__") >= 0 {
			arg = replaceSeq(arg)
		}
		if strings.Index(arg, "__rand_int__") >= 0 {
			arg = replaceRand(arg)
		}
		args = append(args, arg)
	}
	_, err := c.rds.Do(cmd, args...)
	return err
}

//Register to fperf
func init() {
	rand.Seed(time.Now().UnixNano())
	client.Register("redis", newRedisClient, "redis performance benchmark")
}
