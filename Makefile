all:fperf
GC?=go
fperf:bin/fperf/main.go fperf.go client.go
	$(GC) build -o fperf ./bin/fperf

server:example/server/main.go
	$(GC) build -o example/server/server example/server/main.go

clean:
	rm -f fperf example/server/server
