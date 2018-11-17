/*
This is an example about how to build a redis benchmark to using fperf

A fperf testcase in fact is an implementation of fperf.UnaryClient.
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
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/shafreeck/fperf"
)

const seqPlaceHolder = "__seq_int__"
const randPlaceHolder = "__rand_int__"

var seq func() string = seqCreater(0)
var random func() string = randCreater(10000000000000000)

//A test case can have itself options witch would be passed by fperf
type options struct {
	verbose bool
	auth    string
	load    string
}

type command struct {
	name string
	args []interface{}
}

//A client is a struct that should implement fperf.UnaryClient
type redisClient struct {
	args     []string   //the args of client, we use redis command as args
	rds      redis.Conn //the redis connection, should be created when call Dial
	options  options    //the options user set
	commands []command  //commands read from file
}

//newRedisClient create the client object. The function should be
//registered to fperf, fperf -h will list all the registered clients(testcases)
func newRedisClient(flag *fperf.FlagSet) fperf.Client {
	c := new(redisClient)
	flag.BoolVar(&c.options.verbose, "v", false, "verbose")
	flag.StringVar(&c.options.auth, "a", "", "auth of redis")
	flag.StringVar(&c.options.load, "load", "", "load commands from file")

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
	if c.options.load != "" {
		c.readFile()
	}
	return c
}

//Dial to redis server. The addr is set by the fperf option "-server"
func (c *redisClient) Dial(addr string) error {
	rds, err := redis.DialURL(addr)
	if err != nil {
		return err
	}
	if c.options.auth != "" {
		rds.Do("auth", c.options.auth)
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
	return strings.Replace(s, seqPlaceHolder, seq(), -1)
}
func replaceRand(s string) string {
	return strings.Replace(s, randPlaceHolder, random(), -1)
}

func (c *redisClient) readFile() error {
	file, err := os.Open(c.options.load)
	if err != nil {
		return err
	}
	defer file.Close()

	var commands []command
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		cmd := command{name: fields[0]}
		for _, arg := range fields[1:] {
			cmd.args = append(cmd.args, arg)
		}

		commands = append(commands, cmd)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.commands = commands
	return nil
}

func replace(s string) string {
	if strings.Index(s, seqPlaceHolder) >= 0 {
		s = replaceSeq(s)
	}
	if strings.Index(s, randPlaceHolder) >= 0 {
		s = replaceRand(s)
	}
	return s
}

func (c *redisClient) RequestBatch() error {
	for _, cmd := range c.commands {
		var args []interface{}
		name := replace(cmd.name)
		for _, arg := range cmd.args {
			args = append(args, replace(arg.(string)))
		}

		if err := c.rds.Send(name, args...); err != nil {
			return err
		}
	}
	if err := c.rds.Flush(); err != nil {
		return err
	}
	for _ = range c.commands {
		_, err := c.rds.Receive()
		if err != nil {
			return err
		}
	}
	return nil
}

//Request send a redis request and return the error if there is
func (c *redisClient) Request() error {
	if c.options.load != "" {
		return c.RequestBatch()
	}

	var args []interface{}

	//Build the redis cmd and args
	cmd := c.args[0]
	for _, arg := range c.args[1:] {
		if strings.Index(arg, seqPlaceHolder) >= 0 {
			arg = replaceSeq(arg)
		}
		if strings.Index(arg, randPlaceHolder) >= 0 {
			arg = replaceRand(arg)
		}
		args = append(args, arg)
	}
	_, err := c.rds.Do(cmd, args...)
	return err
}

//Register to fperf
func init() {
	//rand.Seed(time.Now().UnixNano())
	fperf.Register("redis", newRedisClient, "redis performance benchmark")
}
