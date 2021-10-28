package ultron

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
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

	masterRunner struct {
		scheduler  *scheduler
		plan       Plan
		eventbus   *eventbus
		supervisor *slaveSupervisor
		mu         sync.RWMutex
	}
)

const (
	DefaultGRPC = ":2021"
	DefaultREST = "127.0.0.1:2017"
)

func NewMasterRunner() MasterRunner {
	runner := &masterRunner{
		eventbus: defaultEventBus,
	}

	return runner
}

// Launch 主线程，如果发生错误则关闭
func (r *masterRunner) Launch(con RunnerConfig, opts ...grpc.ServerOption) error {
	if con.GRPCAddr == "" {
		con.GRPCAddr = DefaultGRPC
	}
	if con.RESTAddr == "" {
		con.RESTAddr = DefaultREST
	}

	lis, err := net.Listen("tcp", con.GRPCAddr)
	if err != nil {
		log.Fatal("failed to launch ultron server", zap.Error(err))
	}
	grpcServer := grpc.NewServer(opts...)
	r.supervisor = newSlaveSupervisor()
	genproto.RegisterUltronAPIServer(grpcServer, r.supervisor)
	log.Info("ultron grpc server is running", zap.String("connect_address", con.GRPCAddr))

	// eventbus初始化
	r.eventbus.subscribeReport(printReportToConsole(os.Stdout))
	r.eventbus.start()
	log.Info("report bus is working")

	block := make(chan error, 1)
	go func() {
		block <- grpcServer.Serve(lis)
	}()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-sigs
		log.Warn("caught quit signal, try to shutdown ultron server", zap.String("signal", sig.String()))

		if r.scheduler != nil {
			if err := r.scheduler.stop(false); err != nil {
				log.Error("failed to interrupt current test plan", zap.Error(err))
				os.Exit(1)
			}
		}
		grpcServer.GracefulStop()
		os.Exit(0)
	}()
	err = <-block
	log.Fatal("ultron runner is shutdown", zap.Error(err))
	return err
}

func (r *masterRunner) StartPlan(p Plan) error {
	r.mu.Lock()
	if r.plan != nil && r.plan.Status() == StatusRunning {
		r.mu.Unlock()
		err := errors.New("cannot start a new plan before shutdown current running plan")
		log.Error("failed to start a new plan", zap.Error(err))
		return err
	}
	scheduler := newScheduler(r.supervisor)
	r.scheduler = scheduler
	r.plan = p
	r.mu.Unlock()

	if err := scheduler.start(p.(*plan)); err != nil {
		log.Error("failed to start a new plan", zap.Error(err))
		return err
	}
	go func() {
		if err := scheduler.patrol(5 * time.Second); err != nil {
			log.Warn("patrol mission is complete", zap.Error(err))
		}
	}()
	return nil
}

func (r *masterRunner) StopPlan() {
	r.mu.RLock()
	if r.scheduler == nil {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()
	if err := r.scheduler.stop(false); err != nil {
		log.Error("failed to stop plan", zap.Error(err))
	}
}

func (r *masterRunner) SubscribeReport(fns ...ReportHandleFunc) {
	for _, fn := range fns {
		r.eventbus.subscribeReport(fn)
	}
}
