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

func (c *demoClient) Dial(addr string) error {
	fmt.Println("Dial to", addr)
	return nil
}

func (c *demoClient) Request() error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

func init() {
	client.Register("demo", newDemoClient, "This is a demo client discription")
}
