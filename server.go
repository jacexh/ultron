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
		mu     sync.RWMutex
	}

	slaveAgent struct {
		session   *genproto.Session
		input     chan *genproto.Event
		callbacks map[uint32]chan *statistics.StatisticianGroup
		closed    uint32
		mu        sync.Mutex
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
	if atomic.CompareAndSwapUint32(&sa.closed, 0, 1) {
		sa.input <- &genproto.Event{Type: genproto.EventType_DISCONNECT}
		close(sa.input)
	}
}

func (sa *slaveAgent) send(event *genproto.Event) error {
	if atomic.LoadUint32(&sa.closed) == 0 {
		sa.input <- event
		return nil
	}
	return fmt.Errorf("slave-%s is closed", sa.ID())
}

func (sa *slaveAgent) callback(batch uint32, sg *statistics.StatisticianGroup) error {
	sa.mu.Lock()
	if ch, ok := sa.callbacks[batch]; ok {
		delete(sa.callbacks, batch)
		sa.mu.Unlock()
		ch <- sg
		return nil
	}
	sa.mu.Unlock()
	return errors.New("batch not found")
}

// TODO:
func (sa *slaveAgent) Submit(ctx context.Context, batch uint32) (*statistics.StatisticianGroup, error) {
	recv := make(chan *statistics.StatisticianGroup, 1)
	defer close(recv)

	sa.mu.Lock()
	sa.callbacks[batch] = recv
	// 清理
	for old, ch := range sa.callbacks {
		if (batch - old) >= 5 {
			delete(sa.callbacks, old)
			close(ch)
		}
	}
	sa.mu.Unlock()

	if err := sa.send(&genproto.Event{Type: genproto.EventType_STATS_AGGREGATE, Data: &genproto.Event_BatchId{BatchId: batch}}); err != nil {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		delete(sa.callbacks, batch)
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case sg := <-recv:
		return sg, nil
	}
}

// TODO:
func (u *ultronServer) Subscribe(session *genproto.Session, sendStream genproto.UltronService_SubscribeServer) error {
	agent := newSlaveAgent(session, sendStream)
	defer agent.close()

	if agent.ID() == "" {
		return errors.New("empty client id")
	}

	u.mu.Lock()
	u.slaves[agent.ID()] = agent
	u.mu.Unlock()

	<-sendStream.Context().Done()
	return io.EOF

	defaultMessageBus.publish(&messageSlaveConnected{session: session})

	if err := sendStream.Send(&genproto.Event{Type: genproto.EventType_CONNECTED}); err != nil {
		return err
	}

	for event := range agent.input {
		if err := sendStream.Send(event); err != nil {
			return err
		}
		if event.Type == genproto.EventType_DISCONNECT {
			break
		}
	}
	return io.EOF
}

// TODO:
func (u *ultronServer) Submit(ctx context.Context, report *genproto.RequestSubmit) (*genproto.ResponseSubmit, error) {
	u.mu.Lock()
	slave, ok := u.slaves[report.GetSlaveId()]
	u.mu.Unlock()

	if !ok {
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_UNKNOWN_BATCH}, errors.New("reject report from slave")
	}

	sg, err := statistics.NewStatisticianGroupFromDTO(report.GetStats())
	if err != nil {
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_UNKNOWN_BATCH}, err
	}
	if err = slave.callback(report.GetBatchId(), sg); err != nil {
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_UNKNOWN_BATCH}, err
	}
	return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_ACCEPTED}, nil
}
