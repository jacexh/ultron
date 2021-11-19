package ultron

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"google.golang.org/grpc"
)

func prepareGRPCServer() (genproto.UltronAPIServer, net.Listener) {
	lis, err := net.Listen("tcp", ":2021")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	ultron := newSlaveSupervisor()
	genproto.RegisterUltronAPIServer(s, ultron)

	go func() {
		if err := s.Serve(lis); err != nil {
			Logger.Error(err.Error())
		}
	}()
	return ultron, lis
}

func TestSlaveRunner_Connect(t *testing.T) {
	slave := newSlaveRunner()
	err := slave.Connect("")
	assert.NotNil(t, err)
	assert.EqualValues(t, err.Error(), "you should assgin a task before connect")

	slave.Assign(NewTask())

	_, lis := prepareGRPCServer()
	defer lis.Close()
	err = slave.Connect("127.0.0.1:2021", grpc.WithInsecure())
	assert.Nil(t, err)
}

func TestSlaveRunner_Working(t *testing.T) {
	slave := newSlaveRunner()
	task := NewTask()
	task.Add(&HTTPAttacker{}, 1)
	slave.Assign(task)
	ultron, lis := prepareGRPCServer()
	defer lis.Close()
	err := slave.Connect("127.0.0.1:2021", grpc.WithInsecure())
	assert.Nil(t, err)

	supervisor := ultron.(*slaveSupervisor)
	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_PLAN_STARTED, Data: &genproto.SubscribeResponse_PlanName{PlanName: "foobar"}})
	assert.Nil(t, err)

	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_STATS_AGGREGATE, Data: &genproto.SubscribeResponse_BatchId{BatchId: 0}})
	assert.Nil(t, err)

	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_STATUS_REPORT})
	assert.Nil(t, err)

	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_PING})
	assert.Nil(t, err)

	dto, _ := defaultAttackStrategyConverter.convertAttackStrategy(&FixedConcurrentUsers{ConcurrentUsers: 100, RampUpPeriod: 10})
	timer, _ := defaultTimerConverter.convertTimer(NonstopTimer{})
	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{
		Type:  genproto.EventType_NEXT_STAGE_STARTED,
		Data:  &genproto.SubscribeResponse_AttackStrategy{AttackStrategy: dto},
		Timer: timer,
	})
	assert.Nil(t, err)

	// err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_PLAN_INTERRUPTED})
	// assert.Nil(t, err)

	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_PLAN_FINISHED})
	assert.Nil(t, err)

	err = supervisor.batchSend(context.TODO(), &genproto.SubscribeResponse{Type: genproto.EventType_DISCONNECT})
	assert.Nil(t, err)

	<-time.After(1 * time.Second)
}
