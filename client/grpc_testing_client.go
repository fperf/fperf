package client

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	testpb "google.golang.org/grpc/benchmark/grpc_testing"
	"google.golang.org/grpc/grpclog"
)

func init() {
	Register("grpc_testing", newTestpbClient, "benchmark for grpc/benchmark/grpc_testing")
}

type testpbClient struct {
	cli testpb.TestServiceClient
}

func newTestpbClient(flag *FlagSet) Client {
	return new(testpbClient)
}

func (r *testpbClient) Dial(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	r.cli = testpb.NewTestServiceClient(conn)
	return nil
}

func (r *testpbClient) Request() error {
	pl := newPayload(0, 20)
	sr := &testpb.SimpleRequest{
		ResponseType: pl.Type,
		ResponseSize: int32(10),
		Payload:      pl,
	}
	if _, err := r.cli.UnaryCall(context.Background(), sr); err != nil {
		return err
	}
	return nil
}

func (r *testpbClient) CreateStream(ctx context.Context) (Stream, error) {
	s, err := r.cli.StreamingCall(ctx)
	pl := newPayload(0, 20)
	return &testpbStream{stream: s, pl: pl}, err
}

type testpbStream struct {
	stream testpb.TestService_StreamingCallClient
	pl     *testpb.Payload
}

func newPayload(t testpb.PayloadType, size int) *testpb.Payload {
	if size < 0 {
		grpclog.Fatalf("Requested a response with invalid length %d", size)
	}
	body := make([]byte, size)
	switch t {
	case testpb.PayloadType_COMPRESSABLE:
	case testpb.PayloadType_UNCOMPRESSABLE:
		grpclog.Fatalf("PayloadType UNCOMPRESSABLE is not supported")
	default:
		grpclog.Fatalf("Unsupported payload type: %d", t)
	}
	return &testpb.Payload{
		Type: t,
		Body: body,
	}
}

func (s *testpbStream) DoSend() error {
	sr := &testpb.SimpleRequest{
		ResponseType: s.pl.Type,
		ResponseSize: int32(10),
		Payload:      s.pl,
	}
	return s.stream.Send(sr)
}
func (s *testpbStream) DoRecv() error {
	_, err := s.stream.Recv()
	return err
}
