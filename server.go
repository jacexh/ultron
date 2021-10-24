package ultron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
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

func newSlaveAgent(session *genproto.Session) *slaveAgent {
	return &slaveAgent{
		session:   session,
		input:     make(chan *genproto.Event, 1),
		callbacks: make(map[uint32]chan *statistics.StatisticianGroup),
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

// Submit SlaveAgent.Submit的实现
func (sa *slaveAgent) Submit(ctx context.Context, batch uint32) (*statistics.StatisticianGroup, error) {
	recv := make(chan *statistics.StatisticianGroup, 1)
	defer close(recv)

	sa.mu.Lock()
	_, exists := sa.callbacks[batch]
	if exists {
		sa.mu.Unlock()
		return nil, fmt.Errorf("slave agent %s received conflicted batch id: %d", sa.ID(), batch)
	}
	sa.callbacks[batch] = recv
	// 清理
	for old, ch := range sa.callbacks {
		if (batch - old) >= 5 {
			delete(sa.callbacks, old)
			close(ch)
		}
	}
	sa.mu.Unlock()

	defer func() {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		delete(sa.callbacks, batch)
	}()

	go func() {
		if err := sa.send(&genproto.Event{Type: genproto.EventType_STATS_AGGREGATE, Data: &genproto.Event_BatchId{BatchId: batch}}); err != nil {
			Logger.Error("the slave agent is closed, didnot send EventType_STATS_AGGREGATE", zap.String("slave_agent", sa.ID()), zap.Error(err))
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case sg := <-recv:
		return sg, nil
	}
}

func (sa *slaveAgent) keepAlives() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := sa.send(&genproto.Event{Type: genproto.EventType_PING}); err != nil {
			Logger.Info("the slave agent is closed, stop the ticker", zap.String("slave_id", sa.session.SlaveId), zap.Error(err))
			return
		}
	}
}

func NewUltronServer() genproto.UltronServiceServer {
	return &ultronServer{
		slaves: make(map[string]*slaveAgent),
	}
}

// Subscribe 具体实现
func (u *ultronServer) Subscribe(session *genproto.Session, sendStream genproto.UltronService_SubscribeServer) error {
	agent := newSlaveAgent(session)
	defer agent.close()

	if agent.ID() == "" {
		Logger.Error("cannot subscribe to ultron server with empty slave id")
		return errors.New("empty client id")
	}

	u.mu.Lock()
	_, exists := u.slaves[agent.ID()]
	if exists {
		Logger.Error("cannot subscribe to ultron server with conflicted slave id", zap.String("slave_id", agent.ID()))
		return fmt.Errorf("conflicted slave id: %s", agent.ID())
	}
	u.slaves[agent.ID()] = agent
	u.mu.Unlock()
	Logger.Info("a new slave is subscribing to ultron server", zap.String("slave_id", session.SlaveId), zap.Any("extras", session.Extras))

	defer func() {
		u.mu.Lock()
		defer u.mu.Unlock()
		delete(u.slaves, agent.ID())
	}() // TODO: SlaveAgent should convert as StatsProvider then register to StatsAggregator

	// 防止阻塞
	go func(agent *slaveAgent) {
		if err := agent.send(&genproto.Event{Type: genproto.EventType_CONNECTED}); err != nil {
			Logger.Error("the slave agent is closed, failed to send EventType_CONNECTED", zap.String("slave_id", agent.ID()), zap.Error(err))
			return
		}
		agent.keepAlives()
	}(agent)

subscribing:
	for {
		select {
		case <-sendStream.Context().Done():
			Logger.Error("the slave has disconnected to this ultron server", zap.String("slave_id", agent.ID()), zap.Error(sendStream.Context().Err()))
			break subscribing
		case event := <-agent.input:
			if err := sendStream.Send(event); err != nil {
				Logger.Error("failed to send message to slave", zap.String("slave_id", agent.ID()), zap.Any("event", event), zap.Error(err))
				return err
			}
			if event.Type == genproto.EventType_DISCONNECT {
				Logger.Warn("ultron server would disconnect from slave", zap.String("slave_id", agent.ID()))
				return nil
			}
		}
	}
	return io.EOF
}

// Submit 提交统计报告
func (u *ultronServer) Submit(ctx context.Context, report *genproto.RequestSubmit) (*genproto.ResponseSubmit, error) {
	u.mu.Lock()
	slave, ok := u.slaves[report.GetSlaveId()]
	u.mu.Unlock()

	if !ok {
		Logger.Error("unregistered slave submitted stats report", zap.String("slave_id", report.SlaveId), zap.Uint32("batch_id", report.BatchId))
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_UNREGISTERED_SLAVE}, fmt.Errorf("unknown slave id: %s", report.GetSlaveId())
	}

	sg, err := statistics.NewStatisticianGroupFromDTO(report.GetStats())
	if err != nil {
		Logger.Error("slave submitted bad report", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()), zap.Error(err))
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_BAD_SUBMISSION}, err
	}
	if err = slave.callback(report.GetBatchId(), sg); err != nil {
		Logger.Error("failed to handle stats report", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()), zap.Error(err))
		return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_BATCH_REJECTED}, err
	}
	Logger.Info("accepted stats report from slave", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()))
	return &genproto.ResponseSubmit{Result: genproto.ResponseSubmit_ACCEPTED}, nil
}
