package ultron

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	// slaveRunner ultron slave的实现
	slaveRunner struct {
		id              string
		ctx             context.Context
		cancel          context.CancelFunc
		client          genproto.UltronAPIClient
		commander       AttackStrategyCommander
		stats           *statistics.StatisticianGroup
		task            Task
		eventbus        *eventbus
		subscribeStream genproto.UltronAPI_SubscribeClient
	}
)

var _ SlaveRunner = (*slaveRunner)(nil)

const (
	KeyPlan     = "plan"
	KeyAttacker = "attacker"
)

func newSlaveRunner() *slaveRunner {
	return &slaveRunner{
		id:       uuid.NewString(),
		stats:    statistics.NewStatisticianGroup(),
		eventbus: defaultEventBus,
	}
}

func (sr *slaveRunner) Connect(addr string, opts ...grpc.DialOption) error {
	if sr.task == nil {
		Logger.Fatal("you should assign a task before call connect function")
		return errors.New("you should assgin a task before connect")
	}
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
	sr.subscribeStream = streams
	go sr.working(streams)
	sr.eventbus.start()
	Logger.Info("salve is subscribing ultron server", zap.String("slave_id", sr.id))
	return nil
}

func (sr *slaveRunner) Assign(t Task) {
	sr.task = t
}

func (sr *slaveRunner) SubscribeResult(fns ...ResultHandleFunc) {
	for _, fn := range fns {
		sr.eventbus.subscribeResult(fn)
	}
}

func (sr *slaveRunner) working(streams genproto.UltronAPI_SubscribeClient) {
	for {
		event, err := streams.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				Logger.Warn("ultron server shutdown this connection", zap.Error(err))
				return
			}
			Logger.Fatal("failed to receive events from ultron server", zap.Error(err))
			return
		}

		Logger.Info("received a new event", zap.Any("event", event))

		// 在实现上确保串行
		switch event.GetType() {
		case genproto.EventType_DISCONNECT:
			Logger.Warn("ultron server ask me to shutdown connection")
			return

		case genproto.EventType_STATS_AGGREGATE:
			sr.submit(event.GetBatchId())

		case genproto.EventType_PLAN_FINISHED, genproto.EventType_PLAN_INTERRUPTED:
			sr.stopPlan()

		case genproto.EventType_PLAN_STARTED:
			sr.startPlan(event.GetPlanName())

		case genproto.EventType_NEXT_STAGE_STARTED:
			sr.startNextStage(event.GetAttackStrategy(), event.GetTimer())

		case genproto.EventType_STATUS_REPORT:
			sr.sendStatus()

		default:
			continue
		}
	}
}

func (sr *slaveRunner) startPlan(name string) {
	if sr.commander != nil {
		sr.commander.Close()
		sr.commander = nil
	}

	sr.stats.Reset()
	sr.stats.Attach(statistics.Tag{Key: KeyPlan, Value: name})
}

func (sr *slaveRunner) submit(batch uint32) {
	dto, err := statistics.ConvertStatisticianGroup(sr.stats)
	if err != nil {
		Logger.Error("failed to convert StatisticianGroup", zap.Uint32("batch", batch), zap.Error(err))
		return
	}
	go func() {
		if _, err := sr.client.Submit(sr.ctx, &genproto.SubmitRequest{SlaveId: sr.id, BatchId: batch, Stats: dto}); err != nil {
			Logger.Error("failed to submit stats", zap.Error(err))
		}
	}()
}

func (sr *slaveRunner) startNextStage(s *genproto.AttackStrategyDTO, t *genproto.TimerDTO) {
	strategy, err := defaultAttackStrategyConverter.convertDTO(s)
	if err != nil {
		Logger.Error("failed to start next stage", zap.Error(err))
		return
	}
	timer, err := defaultTimerConverter.convertDTO(t)
	if err != nil {
		Logger.Error("failed to start next stage", zap.Error(err))
		return
	}

	if sr.commander == nil {
		sr.commander = defaultCommanderFactory.build(strategy.(namedAttackStrategy).Name())
		output := sr.commander.Open(sr.ctx, sr.task)
		go func(c <-chan statistics.AttackResult) {
			for ret := range c {
				if ret.IsFailure() {
					Logger.Warn("received a failed attack result", zap.Error(ret.Error))
				}
				sr.stats.Record(ret)
				sr.eventbus.publishResult(ret)
			}
		}(output)
	}
	go sr.commander.Command(strategy, timer)
}

func (sr *slaveRunner) stopPlan() {
	if sr.commander != nil {
		sr.commander.Close()
		sr.commander = nil
	}
	Logger.Info("current plan is stopped")
}

func (sr *slaveRunner) sendStatus() {
	req := &genproto.SendStatusRequest{SlaveId: sr.id, ConcurrentUsers: 0}
	if sr.commander != nil {
		req.ConcurrentUsers = int32(sr.commander.ConcurrentUsers())
	}
	go func() {
		if _, err := sr.client.SendStatus(sr.ctx, req); err != nil {
			Logger.Error("failed to send status to ultron master server", zap.Error(err))
		}
	}()
}
