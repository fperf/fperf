package client

import (
	"net/http"
)

func init() {
	Register("http_testing", newHTTPClient, "benchmark for http server")
}

type httpClient struct {
	cli *http.Client
	url string
}

func newHTTPClient(flag *FlagSet) Client {
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
	res, err := c.cli.Get(c.url)
	if err != nil {
		return err
	}
	err = res.Body.Close()
	return err
}
