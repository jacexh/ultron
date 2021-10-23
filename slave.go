package ultron

import (
	"github.com/wosai/ultron/v2/pkg/genproto"
	"google.golang.org/grpc"
)

type (
	// SlaveAgent 在master的代理对象
	SlaveAgent interface {
		ID() string
	}

	Slave interface {
		Connect(string, ...grpc.DialOption) error
		WithTask(*Task)
		SubscriberResult()
	}

	Master interface {
		StartNewPlan()
		FinishPlan()
		AddStages(...Stage)
		Start(bool)
		SubscribeReport()
	}

	Local interface {
		AddStages(...Stage)
		Start(bool)
		WithTask(*Task)
		SubscriberResult()
	}

	slave struct {
		client genproto.UltronServiceClient
	}
)
