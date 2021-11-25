package ultron

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
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

func dialer(srv genproto.UltronAPIServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if srv == nil {
		genproto.RegisterUltronAPIServer(server, newSlaveSupervisor())
	} else {
		genproto.RegisterUltronAPIServer(server, srv)
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
	srv := newSlaveSupervisor()
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
	assert.EqualValues(t, len(srv.Slaves()), 0)
}

func TestSlaverSupervisor_Aggregate(t *testing.T) {
	srv := newSlaveSupervisor()
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

	sa := srv.Get(session.SlaveId)
	assert.NotNil(t, sa)
	go func() {
		<-time.After(1 * time.Second)
		sg := statistics.NewStatisticianGroup()
		sg.Record(statistics.AttackResult{Name: "foobar", Duration: 10 * time.Millisecond})
		dto, err := statistics.ConvertStatisticianGroup(sg)
		if err != nil {
			Logger.Error("bad sg", zap.Error(err))
		}
		assert.Nil(t, err)
		_, err = client.Submit(context.Background(), &genproto.SubmitRequest{
			SlaveId: session.GetSlaveId(),
			BatchId: 0,
			Stats:   dto,
		})
		assert.Nil(t, err)
	}()

	start := time.Now()
	_, err = srv.Aggregate(true)
	duration := time.Since(start)
	assert.Nil(t, err)
	Logger.Info(fmt.Sprintf("Aggregate() coast %s", duration.String()))
}

func TestSlaverSupervisor_Aggregate_FuzzTesting(t *testing.T) {
	srv := newSlaveSupervisor()
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

	sa := srv.Get(session.SlaveId)
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
			_, err := srv.Aggregate(true)
			if err != nil {
				atomic.AddUint32(&canceled, 1)
			} else {
				atomic.AddUint32(&submitted, 1)
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
	assert.Greater(t, submitted, uint32(0))
	assert.EqualValues(t, submitted+canceled, 10)
}
