package master

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	ultronServer struct {
		slaves map[string]*slaveAgent
		mu     sync.Mutex
	}

	slaveAgent struct {
		slaveID   string
		extras    map[string]string // todo: 之后再实现
		input     chan *genproto.SubscribeResponse
		callbacks map[uint32]chan *statistics.StatisticianGroup
		closed    uint32
		mu        sync.Mutex
	}
)

var _ genproto.UltronAPIServer = (*ultronServer)(nil)

func newSlaveAgent(session *genproto.SubscribeRequest) *slaveAgent {
	return &slaveAgent{
		slaveID:   session.SlaveId,
		extras:    session.Extras,
		input:     make(chan *genproto.SubscribeResponse, 1),
		callbacks: make(map[uint32]chan *statistics.StatisticianGroup),
	}
}

func (sa *slaveAgent) ID() string {
	return sa.slaveID
}

func (sa *slaveAgent) close() {
	if atomic.CompareAndSwapUint32(&sa.closed, 0, 1) {
		sa.input <- &genproto.SubscribeResponse{Type: genproto.EventType_DISCONNECT}
		close(sa.input)
	}
}

func (sa *slaveAgent) send(event *genproto.SubscribeResponse) error {
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
	for old := range sa.callbacks {
		if (batch - old) >= 5 {
			delete(sa.callbacks, old)
		}
	}
	sa.mu.Unlock()

	defer func() {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		delete(sa.callbacks, batch)
	}()

	go func() {
		if err := sa.send(&genproto.SubscribeResponse{
			Type: genproto.EventType_STATS_AGGREGATE,
			Data: &genproto.SubscribeResponse_BatchId{BatchId: batch}}); err != nil {
			ultron.Logger.Error("the slave agent is closed, didnot send EventType_STATS_AGGREGATE", zap.String("slave_agent", sa.ID()), zap.Error(err))
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
		if err := sa.send(&genproto.SubscribeResponse{Type: genproto.EventType_PING}); err != nil {
			ultron.Logger.Info("the slave agent is closed, stop the ticker", zap.String("slave_id", sa.ID()), zap.Error(err))
			return
		}
	}
}

func NewUltronServer() genproto.UltronAPIServer {
	return &ultronServer{
		slaves: make(map[string]*slaveAgent),
	}
}

// Subscribe 具体实现
func (u *ultronServer) Subscribe(session *genproto.SubscribeRequest, sendStream genproto.UltronAPI_SubscribeServer) error {
	agent := newSlaveAgent(session)
	defer agent.close()

	if agent.ID() == "" {
		ultron.Logger.Error("cannot subscribe to ultron server with empty slave id")
		return errors.New("empty client id")
	}

	u.mu.Lock()
	_, exists := u.slaves[agent.ID()]
	if exists {
		ultron.Logger.Error("cannot subscribe to ultron server with conflicted slave id", zap.String("slave_id", agent.ID()))
		return fmt.Errorf("conflicted slave id: %s", agent.ID())
	}
	u.slaves[agent.ID()] = agent
	u.mu.Unlock()
	ultron.Logger.Info("a new slave is subscribing to ultron server", zap.String("slave_id", session.SlaveId), zap.Any("extras", session.Extras))

	defer func() {
		u.mu.Lock()
		defer u.mu.Unlock()
		delete(u.slaves, agent.ID())
	}() // TODO: SlaveAgent should convert as StatsProvider then register to StatsAggregator

	// 防止阻塞
	go func(agent *slaveAgent) {
		if err := agent.send(&genproto.SubscribeResponse{Type: genproto.EventType_CONNECTED}); err != nil {
			ultron.Logger.Error("the slave agent is closed, failed to send EventType_CONNECTED", zap.String("slave_id", agent.ID()), zap.Error(err))
			return
		}
		agent.keepAlives()
	}(agent)

subscribing:
	for {
		select {
		case <-sendStream.Context().Done():
			ultron.Logger.Error("the slave has disconnected to this ultron server", zap.String("slave_id", agent.ID()), zap.Error(sendStream.Context().Err()))
			break subscribing
		case event := <-agent.input:
			if err := sendStream.Send(event); err != nil {
				ultron.Logger.Error("failed to send message to slave", zap.String("slave_id", agent.ID()), zap.Any("event", event), zap.Error(err))
				return err
			}
			if event.Type == genproto.EventType_DISCONNECT {
				ultron.Logger.Warn("ultron server would disconnect from slave", zap.String("slave_id", agent.ID()))
				return nil
			}
		}
	}
	return io.EOF
}

// Submit 提交统计报告
func (u *ultronServer) Submit(ctx context.Context, report *genproto.SubmitRequest) (*emptypb.Empty, error) {
	u.mu.Lock()
	slave, ok := u.slaves[report.GetSlaveId()]
	u.mu.Unlock()

	if !ok {
		ultron.Logger.Error("unregistered slave submitted stats report", zap.String("slave_id", report.SlaveId), zap.Uint32("batch_id", report.BatchId))
		return &emptypb.Empty{}, fmt.Errorf("unknown slave id: %s", report.GetSlaveId())
	}

	sg, err := statistics.NewStatisticianGroupFromDTO(report.GetStats())
	if err != nil {
		ultron.Logger.Error("slave submitted bad report", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()), zap.Error(err))
		return &emptypb.Empty{}, err
	}
	if err = slave.callback(report.GetBatchId(), sg); err != nil {
		ultron.Logger.Error("failed to handle stats report", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()), zap.Error(err))
		return &emptypb.Empty{}, err
	}
	ultron.Logger.Info("accepted stats report from slave", zap.String("slave_id", report.GetSlaveId()), zap.Uint32("batch_id", report.GetBatchId()))
	return &emptypb.Empty{}, nil
}
