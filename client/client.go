package client

import (
	"flag"
	"golang.org/x/net/context"
)

type NewClientFunc func(*FlagSet) Client
type FlagSet struct {
	*flag.FlagSet
}

var clients map[string]NewClientFunc = make(map[string]NewClientFunc)
var descriptions map[string]string = make(map[string]string)

func (f *FlagSet) Parse() {
	f.FlagSet.Parse(flag.Args()[1:])
}

func NewClient(name string) Client {
	if c := clients[name]; c != nil {
		subcmd := &FlagSet{flag.NewFlagSet(flag.Arg(0), flag.ExitOnError)}
		return c(subcmd)
	}
	return nil
}

func Register(name string, f NewClientFunc, desc ...string) {
	clients[name] = f
	if len(desc) > 0 {
		descriptions[name] = desc[0]
	}
}

//AllClients return the client name and its description
func AllClients() map[string]string {
	m := make(map[string]string)
	for k, _ := range clients {
		m[k] = descriptions[k]
	}
	return m
}

//The interface use can choose to implement
type Client interface {
	Dial(addr string) error
}
type UnaryClient interface {
	Client
	Request() error
}
type StreamClient interface {
	Client
	CreateStream(ctx context.Context) (Stream, error)
}
type Stream interface {
	DoSend() error
	DoRecv() error
}
