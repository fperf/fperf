package http

import (
	"fmt"
	"github.com/shafreeck/fperf"
	"net/http"
	"os"
	"time"
)

type options struct {
	keepalive bool
	url       string
	method    string
	userAgent string
	timeout   time.Duration
}
type httpClient struct {
	cli  http.Client
	opts options
}

func newHTTPClient(flag *fperf.FlagSet) fperf.Client {
	c := new(httpClient)
	flag.BoolVar(&c.opts.keepalive, "keepalive", true, "keep connection alive")
	flag.StringVar(&c.opts.method, "method", "GET", "method of HTTP request, methods:GET,POST,HEAD,OPTIONS,PUT,DELETE")
	flag.StringVar(&c.opts.userAgent, "user-agent", "fperf-http-client", "customize the header User-Agent")
	flag.DurationVar(&c.opts.timeout, "timeout", 10*time.Second, "timeout of request")
	flag.Usage = func() {
		fmt.Printf("Usage: http [options] <url>\noptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	c.opts.url = flag.Arg(0)
	return c
}

func (c *httpClient) Dial(addr string) error {
	tr := &http.Transport{
		DisableKeepAlives: !c.opts.keepalive,
	}
	c.cli = http.Client{
		Transport: tr,
		Timeout:   c.opts.timeout,
	}
	return nil
}

func (c httpClient) Request() error {
	req, err := http.NewRequest(c.opts.method, c.opts.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.opts.userAgent)

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
func init() {
	fperf.Register("http", newHTTPClient, "HTTP performanch benchmark client")
}
