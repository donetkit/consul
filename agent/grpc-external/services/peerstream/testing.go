package peerstream

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/hashicorp/consul/proto/pbpeerstream"
)

type MockClient struct {
	mu sync.Mutex

	ErrCh             chan error
	ReplicationStream *MockStream
}

func (c *MockClient) Send(r *pbpeerstream.ReplicationMessage) error {
	c.ReplicationStream.recvCh <- r
	return nil
}

func (c *MockClient) Recv() (*pbpeerstream.ReplicationMessage, error) {
	return c.RecvWithTimeout(10 * time.Millisecond)
}

func (c *MockClient) RecvWithTimeout(dur time.Duration) (*pbpeerstream.ReplicationMessage, error) {
	select {
	case err := <-c.ErrCh:
		return nil, err
	case r := <-c.ReplicationStream.sendCh:
		return r, nil
	case <-time.After(dur):
		return nil, io.EOF
	}
}

func (c *MockClient) Close() {
	close(c.ReplicationStream.recvCh)
}

func NewMockClient(ctx context.Context) *MockClient {
	return &MockClient{
		ReplicationStream: newTestReplicationStream(ctx),
	}
}

// MockStream mocks peering.PeeringService_StreamResourcesServer
type MockStream struct {
	sendCh chan *pbpeerstream.ReplicationMessage
	recvCh chan *pbpeerstream.ReplicationMessage

	ctx context.Context
}

var _ pbpeerstream.PeerStreamService_StreamResourcesServer = (*MockStream)(nil)

func newTestReplicationStream(ctx context.Context) *MockStream {
	return &MockStream{
		sendCh: make(chan *pbpeerstream.ReplicationMessage, 1),
		recvCh: make(chan *pbpeerstream.ReplicationMessage, 1),
		ctx:    ctx,
	}
}

// Send implements pbpeerstream.PeeringService_StreamResourcesServer
func (s *MockStream) Send(r *pbpeerstream.ReplicationMessage) error {
	s.sendCh <- r
	return nil
}

// Recv implements pbpeerstream.PeeringService_StreamResourcesServer
func (s *MockStream) Recv() (*pbpeerstream.ReplicationMessage, error) {
	r := <-s.recvCh
	if r == nil {
		return nil, io.EOF
	}
	return r, nil
}

// Context implements grpc.ServerStream and grpc.ClientStream
func (s *MockStream) Context() context.Context {
	return s.ctx
}

// SendMsg implements grpc.ServerStream and grpc.ClientStream
func (s *MockStream) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg implements grpc.ServerStream and grpc.ClientStream
func (s *MockStream) RecvMsg(m interface{}) error {
	return nil
}

// SetHeader implements grpc.ServerStream
func (s *MockStream) SetHeader(metadata.MD) error {
	return nil
}

// SendHeader implements grpc.ServerStream
func (s *MockStream) SendHeader(metadata.MD) error {
	return nil
}

// SetTrailer implements grpc.ServerStream
func (s *MockStream) SetTrailer(metadata.MD) {}

// incrementalTime is an artificial clock used during testing. For those
// scenarios you would pass around the method pointer for `Now` in places where
// you would be using `time.Now`.
type incrementalTime struct {
	base time.Time
	next uint64
	mu   sync.Mutex
}

// Now advances the internal clock by 1 second and returns that value.
func (t *incrementalTime) Now() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next++

	dur := time.Duration(t.next) * time.Second

	return t.base.Add(dur)
}

// FutureNow will return a given future value of the Now() function.
// The numerical argument indicates which future Now value you wanted. The
// value must be > 0.
func (t *incrementalTime) FutureNow(n int) time.Time {
	if n < 1 {
		panic(fmt.Sprintf("argument must be > 1 but was %d", n))
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	dur := time.Duration(t.next+uint64(n)) * time.Second

	return t.base.Add(dur)
}
