package ultron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	// slaveRunner ultron slave的实现
	slaveRunner struct {
		id        string
		ctx       context.Context
		cancel    context.CancelFunc
		client    genproto.UltronAPIClient
		commander AttackStrategyCommander
		stats     *statistics.StatisticianGroup
		task      Task
		eventbus  resultBus
		mu        sync.RWMutex
	}
)

var _ SlaveRunner = (*slaveRunner)(nil)

func newSlaveRunner() *slaveRunner {
	return &slaveRunner{
		id:       uuid.NewString(),
		stats:    statistics.NewStatisticianGroup(),
		eventbus: defaultEventBus,
	}
}

func (sr *slaveRunner) Connect(addr string, opts ...grpc.DialOption) error {
	sr.ctx, sr.cancel = context.WithCancel(context.Background())
	conn, err := grpc.DialContext(sr.ctx, addr, opts...)
	if err != nil {
		Logger.Fatal("failed to connect ultron server", zap.Error(err))
		return err
	}
	client := genproto.NewUltronAPIClient(conn)
	streams, err := client.Subscribe(sr.ctx, &genproto.SubscribeRequest{SlaveId: sr.id})
	if err != nil {
		Logger.Fatal("failed to subscribe events from ultron server", zap.Error(err))
		return err
	}

	// 第一条消息接受
	resp, err := streams.Recv()
	if err != nil {
		Logger.Fatal("failed to receive event from ultron server", zap.Error(err))
		return err
	}
	if resp.GetType() != genproto.EventType_CONNECTED {
		err := fmt.Errorf("unexpected event type: %d", resp.Type)
		Logger.Fatal("the first arrvied event is not expected", zap.Error(err))
		return err
	}

	sr.client = client

stream:
	for {
		event, err := streams.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				Logger.Warn("ultron server shutdown this connection", zap.Error(err))
				os.Exit(1)
				return nil
			}
			Logger.Fatal("failed to receive events from ultron server", zap.Error(err))
			return err
		}

		Logger.Info("received a new event", zap.String("slave_id", sr.id), zap.Any("event", event))
		switch event.GetType() {
		case genproto.EventType_DISCONNECT:
			Logger.Warn("ultron server ask me to shutdown connection")
			return nil

		case genproto.EventType_STATS_AGGREGATE:
			dto, err := statistics.ConvertStatisticianGroup(sr.stats)
			if err != nil {
				Logger.Error("failed to convert StatisticianGroup", zap.String("slave_id", sr.id), zap.Uint32("batch", event.GetBatchId()), zap.Error(err))
				continue stream
			}
			go func() {
				if _, err := sr.client.Submit(sr.ctx,
					&genproto.SubmitRequest{SlaveId: sr.id, BatchId: event.GetBatchId(), Stats: dto},
					grpc.MaxCallSendMsgSize(32*1024*1024)); err != nil {
					Logger.Error("failed to submit stats", zap.Error(err))
				}
			}()

		case genproto.EventType_PLAN_FINISHED, genproto.EventType_PLAN_INTERRUPTED:
			sr.commander.Close()

		case genproto.EventType_PLAN_STARTED:
			go func() {
				sr.commander = newFixedConcurrentUsersStrategyCommander()
				retC := sr.commander.Open(sr.ctx, sr.task)
				for ret := range retC {
					sr.stats.Record(ret)
					sr.eventbus.publishResult(ret)
				}
			}()

		case genproto.EventType_NEXT_STAGE_STARTED:
			strategy, err := defaultAttackStrategyConverter.convertDTO(event.GetAttackStrategy())
			if err != nil {
				Logger.Error("failed to convert attack strategy", zap.Error(err))
				continue stream
			}
			timer, err := defaultTimerConverter.convertDTO(event.GetTimer())
			if err != nil {
				Logger.Error("failed to convert timer", zap.Error(err))
				continue stream
			}
			sr.commander.Command(strategy, timer)

		default:
			continue
		}
	}
}

func (sr *slaveRunner) Assign(t Task) {
	sr.task = t
}

func (sr *slaveRunner) SubscribeResult(fns ...ResultHandleFunc) {
	for _, fn := range fns {
		sr.eventbus.subscribeResult(fn)
	}
}
