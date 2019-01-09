package main

import (
	"github.com/fperf/fperf"
	_ "github.com/fperf/http"
	_ "github.com/fperf/mysql"
	_ "github.com/fperf/redis"
)

func main() {
	fperf.Main()
}
