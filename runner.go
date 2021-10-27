package ultron

import (
	"google.golang.org/grpc"
)

type (
	MasterRunner interface {
		Launch(RunnerConfig, ...grpc.ServerOption) error // 服务启动
		StartPlan(Plan) error                            // 开始执行某个测试计划
		StopPlan()                                       // 停止当前计划
		SubscribeReport(...ReportHandleFunc)             // 订阅聚合报告
	}

	SlaveRunner interface {
		Connect(string, ...grpc.DialOption) error // 连接master
		SubscriberResult(...ResultHandleFunc)     // 订阅Attacker的执行结果
		Assign(*Task)                             // 指派压测任务
	}

	LocalRunner interface {
		Launch(RunnerConfig) error
		Assign(*Task)
		SubscribeReport(...ReportHandleFunc)
		SubscriberResult(...ResultHandleFunc)
		StartPlan(Plan)
		StopPlan()
	}

	RunnerConfig struct {
		WebConsole bool   `json:"web_console"`            // 是否打开启web控制台
		GRPCAddr   string `json:"listern_addr,omitempty"` // 服务监听地址
		RESTAddr   string `json:"rest_addr,omitempty"`    // restful监听地址
		RunOnce    bool   `json:"run_once"`               // 作用于LocalRunner，如果true，则执行完后退出ultron
	}
)

const (
	DefaultGRPC = ":2021"
	DefaultREST = "127.0.0.1:2017"
)

var (
	masterRunnerBuilder func() MasterRunner
)

// BuildMasterRunner todo:
func BuildMasterRunner() MasterRunner {
	return masterRunnerBuilder()
}

// BuildSlaveRunner todo:
func BuildSlaveRunner() SlaveRunner {
	return nil
}

// BuildLocalRunner todo:
func BuildLocalRunner() LocalRunner {
	return nil
}

func RegisterMasterRunnerBuilder(fn func() MasterRunner) {
	masterRunnerBuilder = fn
}
