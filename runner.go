package ultron

import (
	"google.golang.org/grpc"
)

type (
	MasterRunner interface {
		Launch(interface{})                  // 服务启动
		StartPlan(Plan)                      // 开始执行某个测试计划
		StopPlan()                           // 停止当前计划
		SubscribeReport(...ReportHandleFunc) // 订阅聚合报告
	}

	SlaveRunner interface {
		Connect(string, ...grpc.DialOption) error // 连接master
		SubscriberResult(...ResultHandleFunc)     // 订阅Attacker的执行结果
		Assign(*Task)                             // 指派压测任务
	}

	LocalRunner interface {
		Launch(interface{})
		Assign(*Task)
		SubscribeReport(...ReportHandleFunc)
		SubscriberResult(...ResultHandleFunc)
		StartPlan(Plan)
		StopPlan()
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
