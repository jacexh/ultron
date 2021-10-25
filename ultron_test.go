package ultron

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
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

func dialer(srv genproto.UltronServiceServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if srv == nil {
		genproto.RegisterUltronServiceServer(server, NewUltronServer())
	} else {
		genproto.RegisterUltronServiceServer(server, srv)
	}

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
	timer := time.NewTimer(3 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(nil)))
	assert.Nil(t, err)

	client := genproto.NewUltronServiceClient(conn)
	streams, err := client.Subscribe(context.Background(), &genproto.Session{SlaveId: uuid.NewString()})
	assert.Nil(t, err)

	msg, err := streams.Recv()
	assert.Nil(t, err)
	assert.EqualValues(t, msg.Type, genproto.EventType_CONNECTED)

	_ = streams.CloseSend()
}

// func TestClient(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	conn, err := grpc.DialContext(ctx, "127.0.0.1:2021", grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 15 * time.Second, Timeout: 3 * time.Second, PermitWithoutStream: false}))
// 	if err != nil {
// 		Logger.Error("failed to connect to server", zap.Error(err))
// 	}
// 	client := genproto.NewUltronServiceClient(conn)
// 	stream, err := client.Subscribe(ctx, &genproto.Session{SlaveId: uuid.NewString(), Extras: map[string]string{"foobar": "hello world"}})
// 	if err != nil {
// 		Logger.Error("got error", zap.Error(err))
// 	}

// 	go func() {
// 		for {
// 			event, err := stream.Recv()
// 			if err != nil {
// 				Logger.Error("got error", zap.Error(err))
// 				return
// 			}
// 			Logger.Info("event", zap.Any("event", event))
// 		}
// 	}()

// 	// time.Sleep(3 * time.Second)
// 	// stream.CloseSend()

// 	time.Sleep(60 * time.Minute)
// }

func Test_ultronServer_Submit(t *testing.T) {
	srv := NewUltronServer()
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.NewTimer(4 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(srv)))
	assert.Nil(t, err)

	client := genproto.NewUltronServiceClient(conn)
	session := &genproto.Session{SlaveId: uuid.NewString()}
	streams, err := client.Subscribe(context.Background(), session)
	assert.Nil(t, err)
	msg, err := streams.Recv()
	assert.Nil(t, err)
	assert.EqualValues(t, msg.Type, genproto.EventType_CONNECTED)

	sa, exists := srv.(*ultronServer).slaves[session.SlaveId]
	assert.True(t, exists)
	assert.NotNil(t, sa)
	go func() {
		<-time.After(1 * time.Second)
		sg := statistics.NewStatisticianGroup()
		sg.Record(statistics.AttackResult{Name: "foobar", Duration: 10 * time.Millisecond})
		dto, err := statistics.ConvertStatisticianGroup(sg)
		Logger.Info("ready to submit", zap.Any("report", sg.Report(true)))
		if err != nil {
			Logger.Error("bad sg", zap.Error(err))
		}
		assert.Nil(t, err)
		res, err := client.Submit(context.Background(), &genproto.RequestSubmit{
			SlaveId: session.GetSlaveId(),
			BatchId: 1,
			Stats:   dto,
		})
		assert.Nil(t, err)
		assert.EqualValues(t, res.Result, genproto.ResponseSubmit_ACCEPTED)
	}()

	sg, err := sa.Submit(ctx, 1)
	assert.Nil(t, err)
	Logger.Info("accepted report", zap.Any("report", sg.Report(true)))
}

func Test_ultronServer_Submit_FuzzTesting(t *testing.T) {
	srv := NewUltronServer()
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.NewTimer(15 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(srv)))
	assert.Nil(t, err)

	client := genproto.NewUltronServiceClient(conn)
	session := &genproto.Session{SlaveId: uuid.NewString()}
	streams, err := client.Subscribe(context.Background(), session)
	assert.Nil(t, err)
	msg, err := streams.Recv()
	assert.Nil(t, err)
	assert.EqualValues(t, msg.Type, genproto.EventType_CONNECTED)

	sa, exists := srv.(*ultronServer).slaves[session.SlaveId]
	assert.True(t, exists)
	assert.NotNil(t, sa)

	// get ready
	sg := statistics.NewStatisticianGroup()
	sg.Record(statistics.AttackResult{Name: "foobar", Duration: 10 * time.Millisecond})
	dto, err := statistics.ConvertStatisticianGroup(sg)
	assert.Nil(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(batch uint32) {
			defer wg.Done()
			Logger.Info("start to batch", zap.Uint32("batch_id", batch))
			accepted, err := sa.Submit(ctx, batch)
			if err == nil {
				Logger.Info("submitted", zap.Uint32("batch", batch))
				report := accepted.Report(true)
				assert.EqualValues(t, report.TotalRequests, 1)
				assert.EqualValues(t, report.Reports["foobar"].Min, 10*time.Millisecond)
				assert.EqualValues(t, report.Reports["foobar"].Max, 10*time.Millisecond)
			} else {
				Logger.Error("failed to submit", zap.Error(err))
			}
		}(uint32(i))
	}

	go func() {
		<-time.After(1 * time.Second)
		block := make(chan struct{}, 3)
		for {
			block <- struct{}{}
			go func() {
				defer func() {
					<-block
				}()

				batch := uint32(rand.Intn(10))
				Logger.Info("client send report", zap.Uint32("batch", batch))
				_, err := client.Submit(context.Background(), &genproto.RequestSubmit{
					SlaveId: session.GetSlaveId(),
					BatchId: batch,
					Stats:   dto,
				})
				if err != nil {
					Logger.Error("client got error", zap.Error(err))
				}
			}()
		}
	}()
	wg.Wait()
}
