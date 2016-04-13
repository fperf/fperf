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

func (c *DemoClient) Dial(addr string) error {
	fmt.Println("Dial to", addr)
	return nil
}

func (c *DemoClient) Request() error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

func init() {
	client.Register("demo", NewDemoClient, "This is a demo client discription")
}
