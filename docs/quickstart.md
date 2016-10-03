# Start
This is a tutorial about how to use the fperf command line. We will take the builtin testcase "http" as an example.

## Installing
`fperf` is developed in golang, make sure you have setup the go environment correctly.

then
```
go get github.com/shafreeck/fperf/bin/fperf
```

## Run benchmark
`fperf` in fact is a performance benchmark framework, you can use it to develop your own performance benchmark tools.
It has some demo implementation builtin including `http` and `grpc_testing`. We will take the `http` to
demonstrate how to use `fperf`

### Run by default
```
fperf http http://example.com
```

We call the `http` a testcase, `fperf` runs the testcase and pass an url to it. This will run the benchmark
with default fperf options(1 connection, 1 goroutine, no delay)

### Run by concurrency
```
fperf -cpu 8 -connection 100 -goroutine 2 -tick 1s http http://example.com
```
```
-cpu :set the GOMAXPROCS witch better be the number of core of cpu
-connection : the number of the connection
-goroutine : the number of goroutine per connection, so we have 200 concurrent goroutines in this example
-ticks : the interval between output and statistics
```

The result:
```
2016/04/13 19:34:47 latency 752.369972ms qps 198 total 198
2016/04/13 19:34:48 latency 813.625948ms qps 138 total 336
2016/04/13 19:34:49 latency 1.277809701s qps 105 total 441
2016/04/13 19:34:50 latency 1.533281819s qps 105 total 546
2016/04/13 19:34:51 latency 1.402846393s qps 152 total 698
2016/04/13 19:34:52 latency 1.308274418s qps 145 total 843
2016/04/13 19:34:53 latency 1.471700286s qps 119 total 963
2016/04/13 19:34:54 latency 1.717590479s qps 82 total 1045
Count: 1145  Min: 684840084  Max: 6674918683  Avg: 1275677320.56
------------------------------------------------------------
[     10000,      11000)     0    0.0%    0.0%
[     11000,      13800)     0    0.0%    0.0%
[     13800,      21639)     0    0.0%    0.0%
[     21639,      43590)     0    0.0%    0.0%
[     43590,     105055)     0    0.0%    0.0%
[    105055,     277158)     0    0.0%    0.0%
[    277158,     759048)     0    0.0%    0.0%
[    759048,    2108340)     0    0.0%    0.0%
[   2108340,    5886359)     0    0.0%    0.0%
[   5886359,   16464814)     0    0.0%    0.0%
[  16464814,   46084490)     0    0.0%    0.0%
[  46084490,  129019584)     0    0.0%    0.0%
[ 129019584,  361237849)     0    0.0%    0.0%
[ 361237849, 1011448991)   813   71.0%   71.0%  #######
[1011448991, 2832040189)   270   23.6%   94.6%  ##
[2832040189,        inf)    62    5.4%  100.0%  #
```
The result has two parts. The first part show the latency and qps witch will be printed
every <tick> time. The second part shows the histogram of latency. This will be outputed
only when you terminate fperf.
