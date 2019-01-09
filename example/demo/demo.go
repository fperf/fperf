package demo

import (
	"fmt"
	"github.com/fperf/fperf"
	"time"
)

type demoClient struct{}

func newDemoClient(flag *fperf.FlagSet) fperf.Client {
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
	fperf.Register("demo", newDemoClient, "This is a demo client discription")
}
