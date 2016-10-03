package main

import (
	"github.com/shafreeck/fperf"
	_ "github.com/shafreeck/fperf/clients/http"
	_ "github.com/shafreeck/fperf/clients/mqtt"
	_ "github.com/shafreeck/fperf/clients/redis"
)

func main() {
	fperf.Main()
}
