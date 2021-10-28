package ultron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	slaveSupervisor struct {
		counter           uint32
		toleranceForDelay uint32 // 容忍延后的批次
		slaveAgents       map[string]*slaveAgent
		buffer            map[uint32]map[string]*statsCallback
		mu                sync.RWMutex
	}

	statsCallback struct {
		stats      *statistics.StatisticianGroup
		signal     chan struct{}
		agent      *slaveAgent
		batch      uint32
		callbacked uint32
	}
)

var (
	_ genproto.UltronAPIServer = (*slaveSupervisor)(nil)
)

func newStatsCallback(agent *slaveAgent, batch uint32) *statsCallback {
	return &statsCallback{
		signal: make(chan struct{}, 1),
		batch:  batch,
		agent:  agent,
	}
}

func (cb *statsCallback) callback(id string, batch uint32, stats *statistics.StatisticianGroup) error {
	if atomic.CompareAndSwapUint32(&cb.callbacked, 0, 1) {
		cb.stats = stats
		cb.signal <- struct{}{}
		return nil
	}
	return fmt.Errorf("do not callback again: %s", cb.id())
}

func (cb *statsCallback) blockUntilCallbacked() <-chan struct{} {
	return cb.signal
}

func (cb *statsCallback) id() string {
	return cb.agent.ID()
}

func (cb *statsCallback) close() {
	atomic.StoreUint32(&cb.callbacked, 1)
	close(cb.signal)
}

func newSlaveSupervisor() *slaveSupervisor {
	return &slaveSupervisor{
		slaveAgents:       make(map[string]*slaveAgent),
		buffer:            make(map[uint32]map[string]*statsCallback),
		toleranceForDelay: 3,
	}
}

func (sup *slaveSupervisor) Subscribe(req *genproto.SubscribeRequest, stream genproto.UltronAPI_SubscribeServer) error {
	agent := newSlaveAgent(req)
	if err := sup.Add(agent); err != nil {
		log.Error("cannot subscribe to ultron server", zap.String("slave_id", agent.ID()), zap.Error(err))
	}
	log.Info("a new slave is subscribing to ultron server", zap.String("slave_id", agent.ID()), zap.Any("extras", agent.extras))

	defer func() {
		sup.Remove(agent.ID())
		if err := agent.close(); err != nil {
			log.Error("failed to close slave agent", zap.String("slave_id", agent.ID()), zap.Error(err))
		}
	}()

	go func() {
		if err := agent.send(&genproto.SubscribeResponse{Type: genproto.EventType_CONNECTED}); err != nil {
			log.Error("the slave agent is closed, failed to send EventType_CONNECTED", zap.String("slave_id", agent.ID()), zap.Error(err))
			return
		}
		agent.keepAlives()
	}()

subscribing:
	for {
		select {
		case <-stream.Context().Done():
			log.Error("the slave has disconnected to ultron server", zap.String("slave_id", agent.ID()), zap.Error(stream.Context().Err()))
			break subscribing

		case event := <-agent.input:
			if err := stream.Send(event); err != nil {
				log.Error("failed to send event to slave", zap.String("slave_id", agent.ID()), zap.Any("event", event), zap.Error(err))
				return err
			}
			if event.Type == genproto.EventType_DISCONNECT {
				log.Warn("ultron server would disconnect from slave", zap.String("slave_id", agent.ID()))
				return nil
			}
		}
	}
	return io.EOF
}

func (sup *slaveSupervisor) Submit(ctx context.Context, req *genproto.SubmitRequest) (*empty.Empty, error) {
	sg, err := statistics.NewStatisticianGroupFromDTO(req.GetStats())
	if err != nil {
		log.Error("slave submitted bad stats report", zap.String("slave_id", req.GetSlaveId()), zap.Uint32("batch_id", req.GetBatchId()), zap.Error(err))
		return &emptypb.Empty{}, err
	}
	sup.mu.RLock()
	defer sup.mu.RUnlock()

	if batchStats, ok := sup.buffer[req.BatchId]; ok {
		if callback, ok := batchStats[req.SlaveId]; ok {
			if err = callback.callback(req.SlaveId, req.BatchId, sg); err != nil {
				log.Error("failed to handle stats report", zap.String("slave_id", req.GetSlaveId()), zap.Uint32("batch_id", req.GetBatchId()), zap.Error(err))
				return &emptypb.Empty{}, err
			}
			log.Info("accepted stats report from slave", zap.String("slave_id", req.GetSlaveId()), zap.Uint32("batch_id", req.BatchId))
			return &emptypb.Empty{}, nil
		}
	}
	log.Warn("ultron server reject this request, there is no matched slaveID or batchID founded", zap.String("slave_id", req.SlaveId), zap.Uint32("batch_id", req.BatchId))
	return &emptypb.Empty{}, fmt.Errorf("submittion rejected: %s", req.SlaveId)
}

func (sup *slaveSupervisor) Aggregate(fullHistory bool) (statistics.SummaryReport, error) {
	sup.mu.Lock()
	batch := sup.counter
	sup.counter++

	if len(sup.slaveAgents) == 0 {
		sup.mu.Unlock()
		return statistics.SummaryReport{}, errors.New("failed to aggregate stats report without slaves")
	}
	sup.buffer[batch] = make(map[string]*statsCallback)
	for _, agent := range sup.slaveAgents {
		sup.buffer[batch][agent.ID()] = newStatsCallback(agent, batch)
	}

	// copy
	callbacks := make([]*statsCallback, len(sup.buffer[batch]))
	i := 0
	for _, callback := range sup.buffer[batch] {
		callbacks[i] = callback
		i++
	}
	sup.mu.Unlock()

	defer func() { // 程序退出前清理批次数据
		sup.mu.Lock()
		delete(sup.buffer, batch)
		sup.mu.Unlock()
	}()

	// 开始接收各个provider上报的数据
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for _, callback := range callbacks {
		callback := callback // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			if err := callback.agent.send(
				&genproto.SubscribeResponse{Type: genproto.EventType_STATS_AGGREGATE, Data: &genproto.SubscribeResponse_BatchId{BatchId: batch}}); err != nil {
				return fmt.Errorf("[%s] %v", callback.agent.ID(), err)
			}
			select {
			case <-callback.blockUntilCallbacked():
				callback.close()
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}
	err := g.Wait()
	if err != nil {
		return statistics.SummaryReport{}, fmt.Errorf("batch-%d: %w", batch, err)
	}

	sup.mu.Lock()
	if (sup.counter - (batch + 1)) > sup.toleranceForDelay { // too late
		sup.mu.Unlock()
		return statistics.SummaryReport{}, fmt.Errorf("batch-%d: too late to accept summary report", batch)
	}
	callbacker := sup.buffer[batch]
	delete(sup.buffer, batch)
	sup.mu.Unlock()

	// 检查是否完成
	sg := statistics.NewStatisticianGroup()
	for _, callback := range callbacker {
		if callback.stats == nil {
			return statistics.SummaryReport{}, fmt.Errorf("not submitted by the deadline: batch-%d, slave: %s", batch, callback.id())
		} else {
			sg.Merge(callback.stats)
		}
	}
	return sg.Report(fullHistory), nil
}

func (sup *slaveSupervisor) Exists(id string) bool {
	sup.mu.RLock()
	defer sup.mu.RUnlock()
	_, exists := sup.slaveAgents[id]
	return exists
}

func (sup *slaveSupervisor) Remove(id string) {
	sup.mu.Lock()
	defer sup.mu.Unlock()

	delete(sup.slaveAgents, id)
}

func (sup *slaveSupervisor) Get(id string) SlaveAgent {
	sup.mu.RLock()
	defer sup.mu.RUnlock()
	sa, exists := sup.slaveAgents[id]
	if exists {
		return sa
	}
	return nil
}

func (sup *slaveSupervisor) Add(sa *slaveAgent) error {
	if sa.ID() == "" || sa == nil {
		return errors.New("empty slave id")
	}

	sup.mu.Lock()
	defer sup.mu.Unlock()

	if _, ok := sup.slaveAgents[sa.ID()]; ok {
		return fmt.Errorf("duplicated slave id: %s", sa.ID())
	}
	sup.slaveAgents[sa.ID()] = sa
	return nil
}

func (sup *slaveSupervisor) Slaves() []SlaveAgent {
	sup.mu.RLock()
	defer sup.mu.RUnlock()
	ret := make([]SlaveAgent, len(sup.slaveAgents))
	i := 0
	for _, agent := range sup.slaveAgents {
		ret[i] = agent
		i++
	}
	return ret
}

func (sup *slaveSupervisor) batchSend(ctx context.Context, event *genproto.SubscribeResponse) error {
	sup.mu.RLock()
	defer sup.mu.RUnlock()

	if len(sup.slaveAgents) == 0 {
		return errors.New("cannot batch send event to empty slave agent")
	}
	eg, _ := errgroup.WithContext(ctx)
	for _, sa := range sup.slaveAgents {
		agent := sa
		eg.Go(func() error {
			return agent.send(event)
		})
	}
	return eg.Wait()
}

func (sup *slaveSupervisor) StartNewPlan(ctx context.Context, name string) error {
	return sup.batchSend(ctx, &genproto.SubscribeResponse{
		Type: genproto.EventType_PLAN_STARTED,
		Data: &genproto.SubscribeResponse_PlanName{PlanName: name},
	})
}

func (sup *slaveSupervisor) NextStage(ctx context.Context, strategy AttackStrategy, t Timer) error {
	if t == nil {
		t = NonstopTimer{}
	}

	sup.mu.RLock()
	slaves := make([]*slaveAgent, len(sup.slaveAgents))
	i := 0
	for _, sa := range sup.slaveAgents {
		slaves[i] = sa
		i++
	}
	sup.mu.RUnlock()

	strategies := strategy.Split(len(slaves)) // 数量可能少于 len(slaves)
	eg, _ := errgroup.WithContext(ctx)
	for i, strategy := range strategies {
		i := i
		strategy := strategy
		eg.Go(func() error {
			var err error
			event := &genproto.SubscribeResponse{Type: genproto.EventType_NEXT_STAGE_STARTED}
			event.Timer, err = defaultTimerConverter.ConvertTimer(t)
			if err != nil {
				return err
			}
			as, err := defaultAttackStrategyConverter.ConvertAttackStrategy(strategy)
			if err != nil {
				return err
			}
			event.Data = &genproto.SubscribeResponse_AttackStrategy{AttackStrategy: as}
			return slaves[i].send(event)
		})
	}
	return eg.Wait()
}

func (sup *slaveSupervisor) Stop(ctx context.Context, done bool) error {
	event := &genproto.SubscribeResponse{Type: genproto.EventType_PLAN_FINISHED}
	if !done {
		event.Type = genproto.EventType_PLAN_INTERRUPTED
	}
	return sup.batchSend(ctx, event)
}
