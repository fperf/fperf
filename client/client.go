package client

import (
	"flag"
	"golang.org/x/net/context"
)

//NewClientFunc defines the function type a client should implement
type NewClientFunc func(*FlagSet) Client

//FlagSet combines the standard flag.FlagSet, this can be used to parse args by the client
type FlagSet struct {
	*flag.FlagSet
}

var clients = make(map[string]NewClientFunc)
var descriptions = make(map[string]string)

//Parse the command line args
func (f *FlagSet) Parse() {
	f.FlagSet.Parse(flag.Args()[1:])
}

//NewClient create a client by the name it registered
func NewClient(name string) Client {
	if c := clients[name]; c != nil {
		subcmd := &FlagSet{flag.NewFlagSet(flag.Arg(0), flag.ExitOnError)}
		return c(subcmd)
	}
	return nil
}

//Register attatch a client to fperf
func Register(name string, f NewClientFunc, desc ...string) {
	clients[name] = f
	if len(desc) > 0 {
		descriptions[name] = desc[0]
	}
}

//AllClients return the client name and its description
func AllClients() map[string]string {
	m := make(map[string]string)
	for k := range clients {
		m[k] = descriptions[k]
	}
	return m
}

//Client use Dial to connect to the server
type Client interface {
	Dial(addr string) error
}

//UnaryClient defines the request-reply access model
type UnaryClient interface {
	Client
	Request() error
}

//StreamClient used to create a stream
type StreamClient interface {
	Client
	CreateStream(ctx context.Context) (Stream, error)
}

//Stream use DoSend/DoRecv to send/recv data or message
type Stream interface {
	DoSend() error
	DoRecv() error
}
