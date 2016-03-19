all:fperf
GC?=go
fperf:fperf.go client/*.go
	$(GC) build -o fperf fperf.go

server:example/server/main.go
	$(GC) build -o example/server/server example/server/main.go

clean:
	rm -f fperf example/server/server
