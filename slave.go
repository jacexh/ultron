package ultron

import (
	"google.golang.org/grpc"
)

type (
	// SlaveAgent 在master的代理对象
	SlaveAgent interface {
		ID() string
	}

	Slave interface {
		Connect(string, ...grpc.CallOption) error
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
)
