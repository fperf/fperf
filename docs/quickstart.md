# Start
This is a tutorial about how to used the grpcperf command line. We will take the builtin testcase "grpc_testing" as the example.

## Run the server
build the example server in the example director
```
cd example/server
go build
./server
```

## Unary call benchmark 
```
grpcperf -server localhost:8808 -cpu 4 -type unary grpc_testing
```
## Stream call benchmark 
```
grpcperf -server localhost:8808 -cpu 4 -type stream grpc_testing
```
