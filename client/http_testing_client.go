package client

import (
	"net/http"
)

func init() {
	Register("http_testing", NewHTTPClient, "benchmark for http server")
}

type httpClient struct {
	cli *http.Client
	url string
}

func NewHTTPClient(flag *FlagSet) Client {
	c := new(httpClient)
	flag.StringVar(&c.url, "url", "", "set the url you request")
	flag.Parse()
	return c
}

func (c *httpClient) Dial(addr string) error {
	c.cli = new(http.Client)
	return nil
}

func (c *httpClient) Request() error {
	_, err := c.cli.Get(c.url)
	return err
}
