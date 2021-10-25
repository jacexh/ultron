package master

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestSlaveAgent_Submit(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := agent.Submit(ctx, 0)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func dialer(srv genproto.UltronAPIServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if srv == nil {
		genproto.RegisterUltronAPIServer(server, NewUltronServer())
	} else {
		genproto.RegisterUltronAPIServer(server, srv)
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			ultron.Logger.Error("shutdown ultron server", zap.Error(err))
		}
	}()
	return func(c context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestUltronServer_Subscribe(t *testing.T) {
	srv := NewUltronServer()
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.NewTimer(3 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(srv)))
	assert.Nil(t, err)

	client := genproto.NewUltronAPIClient(conn)
	clientCtx, clientCancel := context.WithCancel(context.Background())
	streams, err := client.Subscribe(clientCtx, &genproto.SubscribeRequest{SlaveId: uuid.NewString()})
	assert.Nil(t, err)

	msg, err := streams.Recv()
	assert.Nil(t, err)
	assert.EqualValues(t, msg.Type, genproto.EventType_CONNECTED)

	// slave主动断开
	clientCancel()
	<-time.After(1 * time.Second)
	u := srv.(*ultronServer)
	u.mu.Lock()
	defer u.mu.Unlock()
	assert.EqualValues(t, len(u.slaves), 0)
}

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

	client := genproto.NewUltronAPIClient(conn)
	session := &genproto.SubscribeRequest{SlaveId: uuid.NewString()}
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
		ultron.Logger.Info("ready to submit", zap.Any("report", sg.Report(true)))
		if err != nil {
			ultron.Logger.Error("bad sg", zap.Error(err))
		}
		assert.Nil(t, err)
		_, err = client.Submit(context.Background(), &genproto.SubmitRequest{
			SlaveId: session.GetSlaveId(),
			BatchId: 1,
			Stats:   dto,
		})
		assert.Nil(t, err)
	}()

	sg, err := sa.Submit(ctx, 1)
	assert.Nil(t, err)
	ultron.Logger.Info("accepted report", zap.Any("report", sg.Report(true)))
}

func Test_ultronServer_Submit_FuzzTesting(t *testing.T) {
	srv := NewUltronServer()
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.NewTimer(3 * time.Second)
	go func() {
		<-timer.C
		timer.Stop()
		cancel()
	}()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(srv)))
	assert.Nil(t, err)

	client := genproto.NewUltronAPIClient(conn)
	session := &genproto.SubscribeRequest{SlaveId: uuid.NewString()}
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

	var submitted uint32
	var canceled uint32

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(batch uint32) {
			defer wg.Done()
			accepted, err := sa.Submit(ctx, batch)
			if err == nil {
				atomic.AddUint32(&submitted, 1)
				report := accepted.Report(true)
				assert.EqualValues(t, report.TotalRequests, 1)
				assert.EqualValues(t, report.Reports["foobar"].Min, 10*time.Millisecond)
				assert.EqualValues(t, report.Reports["foobar"].Max, 10*time.Millisecond)
			} else {
				atomic.AddUint32(&canceled, 1)
			}
		}(uint32(i))
	}

	go func() {
		// <-time.After(1 * time.Second)
		block := make(chan struct{}, 3)
		for {
			block <- struct{}{}
			go func() {
				defer func() {
					<-block
				}()

				batch := uint32(rand.Intn(10))
				slaveID := session.GetSlaveId()
				if rand.Float64() <= 0.15 {
					slaveID = uuid.NewString()
				}
				client.Submit(context.Background(), &genproto.SubmitRequest{
					SlaveId: slaveID,
					BatchId: batch,
					Stats:   dto,
				})
			}()
		}
	}()
	wg.Wait()
	assert.LessOrEqual(t, submitted, uint32(5))
	assert.EqualValues(t, submitted+canceled, 10)
}
