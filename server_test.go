package ultron

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestSlaveAgent_Submit(t *testing.T) {
	agent := newSlaveAgent(&genproto.Session{SlaveId: uuid.NewString()})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := agent.Submit(ctx, 0)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	genproto.RegisterUltronServiceServer(server, newUltronServer())
	go func() {
		if err := server.Serve(listener); err != nil {
			Logger.Error("shutdown ultron server", zap.Error(err))
		}
	}()
	return func(c context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestUltronServer_Subscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.NewTimer(1 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	assert.Nil(t, err)

	client := genproto.NewUltronServiceClient(conn)
	streams, err := client.Subscribe(context.Background(), &genproto.Session{SlaveId: uuid.NewString()})
	assert.Nil(t, err)
	msg, err := streams.Recv()
	assert.Nil(t, err)
	assert.EqualValues(t, msg.Type, genproto.EventType_CONNECTED)
}
