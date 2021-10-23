package ultron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	ultronServer struct {
		slaves map[string]*slaveAgent
		stats  map[uint32]map[string]chan *statistics.AttackStatisticsDTO
		mu     sync.RWMutex
	}

	slaveAgent struct {
		session *genproto.Session
		// server  genproto.UltronService_SubscribeServer
		input  chan *genproto.Event
		closed uint32
	}
)

var _ genproto.UltronServiceServer = (*ultronServer)(nil)

func newSlaveAgent(session *genproto.Session, server genproto.UltronService_SubscribeServer) *slaveAgent {
	return &slaveAgent{
		session: session,
		// server:  server,
		input: make(chan *genproto.Event, 1),
	}
}

func (sa *slaveAgent) ID() string {
	return sa.session.SlaveId
}

func (sa *slaveAgent) close() {
	atomic.CompareAndSwapUint32(&sa.closed, 0, 1)
	sa.input <- &genproto.Event{Type: genproto.EventType_DISCONNECT}
	close(sa.input)
}

func (sa *slaveAgent) send(event *genproto.Event) error {
	if atomic.LoadUint32(&sa.closed) == 0 {
		sa.input <- event
		return nil
	}
	return fmt.Errorf("slave-%s is closed", sa.ID())
}

func (sa *slaveAgent) Submit(ctx context.Context, batch uint32) (*statistics.StatisticianGroup, error) {
	recv := make(chan *statistics.StatisticianGroup, 1)
	defer close(recv)

	if err := sa.send(&genproto.Event{Type: genproto.EventType_STATS_AGGREGATE, Data: &genproto.Event_BatchId{BatchId: batch}}); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case sg := <-recv:
		// statistics.AttackStatisticsDTO
		return sg, nil
	}
}

func (u *ultronServer) Subscribe(client *genproto.Session, events genproto.UltronService_SubscribeServer) error {
	agent := newSlaveAgent(client, events)
	defer agent.close()

	if agent.ID() == "" {
		return errors.New("empty client id")
	}
	u.slaves[agent.ID()] = agent

	if err := events.Send(&genproto.Event{Type: genproto.EventType_CONNECTED}); err != nil {
		return err
	}

	for event := range agent.input {
		if err := events.Send(event); err != nil {
			return err
		}
		if event.Type == genproto.EventType_DISCONNECT {
			break
		}
	}
	return io.EOF
}

func (u *ultronServer) Submit(ctx context.Context, report *genproto.RequestSubmit) (*genproto.ResponseSubmit, error) {
	if batch, ok := u.stats[report.GetBatchId()]; ok {
		if c, ok := batch[report.GetSlaveId()]; ok {
			// c <- report.GetStats()
			close(c)
		}
	}
	return nil, nil
}
