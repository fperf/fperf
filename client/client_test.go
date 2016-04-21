package client

import (
	"flag"
	"os"
	"testing"
)

type testcli struct{}

func (c *testcli) Dial(addr string) error {
	return nil
}

func TestNewClient(t *testing.T) {
	if c := NewClient("test"); c != nil {
		t.Fail()
	}

	clients["test"] = func(flag *FlagSet) Client { return &testcli{} }
	t.Log(clients)
	if c := NewClient("test"); c == nil {
		t.Fail()
	}
}
func TestAllClients(t *testing.T) {
	for c := range AllClients() {
		t.Logf("%v\n", c)
	}
}

func TestRegister(t *testing.T) {
	Register("test", func(flag *FlagSet) Client { return &testcli{} }, "test description")
}

func TestParse(t *testing.T) {
	os.Args = []string{"fperf"}
	flag.CommandLine.Parse(os.Args)
	fs := &FlagSet{flag.NewFlagSet("test", flag.PanicOnError)}
	fs.Parse()
}
