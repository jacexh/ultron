package ultron

import (
	"google.golang.org/grpc"
)

type (
	MasterRunner interface {
		StartNewPlan()
		FinishCurrentPlan()
		Append(...Stage)
		SubscribeReport(...ReportHandleFunc)
		Run(bool)
	}

	SlaveRunner interface {
		Assign(*Task)
		SubscriberResult(...ResultHandleFunc)
		Connect(string, ...grpc.DialOption) error
	}

	LocalRunner interface {
		Append(...Stage)
		Assign(*Task)
		SubscribeReport(...ReportHandleFunc)
		SubscriberResult(...ResultHandleFunc)
		Run(bool)
	}
)

// BuildMasterRunner todo:
func BuildMasterRunner() MasterRunner {
	return nil
}

// BuildSlaveRunner todo:
func BuildSlaveRunner() SlaveRunner {
	return nil
}

// BuildLocalRunner todo:
func BuildLocalRunner() LocalRunner {
	return nil
}
