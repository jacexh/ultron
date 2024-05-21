package ultron

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	MasterRunner interface {
		Launch(...grpc.ServerOption) error   // 服务启动
		StartPlan(Plan) error                // 开始执行某个测试计划
		StopPlan()                           // 停止当前计划
		SubscribeReport(...ReportHandleFunc) // 订阅聚合报告
	}

	SlaveRunner interface {
		Connect(string, ...grpc.DialOption) error // 连接master
		SubscribeResult(...ResultHandleFunc)      // 订阅Attacker的执行结果
		Assign(Task)                              // 指派压测任务
	}

	LocalRunner interface {
		Launch() error
		Assign(Task)
		SubscribeReport(...ReportHandleFunc)
		SubscribeResult(...ResultHandleFunc)
		StartPlan(Plan) error
		StopPlan()
	}

	masterRunner struct {
		scheduler  *scheduler
		plan       Plan
		eventbus   *eventbus
		supervisor *slaveSupervisor
		rpc        *grpc.Server
		rest       *http.Server
		mu         sync.RWMutex
	}

	localRunner struct {
		master *masterRunner
		slave  *slaveRunner
	}
)

func NewMasterRunner() MasterRunner {
	runner := newMasterRunner()
	go func(r *masterRunner) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-sigs
		Logger.Warn("caught quit signal, try to shutdown ultron master server", zap.String("signal", sig.String()))

		if r.scheduler != nil {
			if err := r.scheduler.stop(false); err != nil {
				Logger.Error("failed to interrupt current test plan", zap.Error(err))
				os.Exit(1)
			}
			r.eventbus.close()
		}
		os.Exit(0)
	}(runner)
	return runner
}

func NewSlaveRunner() SlaveRunner {
	runner := newSlaveRunner()

	go func(r *slaveRunner) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-sigs
		Logger.Warn("caught quit signal, try to disconnect to ultron server", zap.String("signal", sig.String()))

		r.subscribeStream.CloseSend()
		os.Exit(0)
	}(runner)
	return runner
}

func newMasterRunner() *masterRunner {
	return &masterRunner{
		eventbus: defaultEventBus,
	}
}

// Launch 主线程，如果发生错误则关闭
func (r *masterRunner) Launch(opts ...grpc.ServerOption) error {
	Logger.Info("loaded configurations", zap.Any("configrations", loadedOption))
	serverOption := loadedOption.Server

	// eventbus初始化
	r.eventbus.subscribeReport(printReportToConsole(os.Stdout))
	r.eventbus.subscribeReport(printJsonReport(os.Stdout))
	r.eventbus.start()

	start := make(chan struct{}, 1)
	go func() { // http server
		router := buildHTTPRouter(r)
		r.rest = &http.Server{
			Addr:    serverOption.HTTPAddr,
			Handler: router,
		}
		Logger.Info("ultron http server is running", zap.String("address", serverOption.HTTPAddr))
		if err := r.rest.ListenAndServe(); err != nil {
			Logger.Fatal("a error has occurend inside http server", zap.Error(err))
		}
	}()

	go func() { // grpc server
		lis, err := net.Listen("tcp", serverOption.GRPCAddr)
		if err != nil {
			Logger.Fatal("failed to launch grpc server", zap.Error(err))
		}
		r.rpc = grpc.NewServer(opts...)
		r.supervisor = newSlaveSupervisor()
		genproto.RegisterUltronAPIServer(r.rpc, r.supervisor)
		Logger.Info("ultron grpc server is running", zap.String("connect_address", serverOption.GRPCAddr))

		start <- struct{}{}
		if err := r.rpc.Serve(lis); err != nil {
			Logger.Fatal("a error has occurend inside grpc server", zap.Error(err))
		}
	}()

	<-start
	return nil
}

func (r *masterRunner) StartPlan(p Plan) error {
	if p == nil {
		err := errors.New("empty plan")
		Logger.Error("cannot start with empty plan", zap.Error(err))
		return err
	}
	r.mu.Lock()
	if r.plan != nil && r.plan.Status() == StatusRunning {
		r.mu.Unlock()
		err := errors.New("cannot start a new plan before shutdown current running plan")
		Logger.Error("failed to start a new plan", zap.Error(err))
		return err
	}
	Logger.Info("start plan", zap.String("plan_name", p.Name()))
	scheduler := newScheduler(r.supervisor)
	r.scheduler = scheduler
	r.plan = p
	r.mu.Unlock()

	if err := scheduler.start(p.(*plan)); err != nil {
		Logger.Error("failed to start a new plan", zap.Error(err))
		return err
	}
	go func() {
		if err := scheduler.patrol(5 * time.Second); err != nil {
			Logger.Warn("patrol mission is complete", zap.Error(err))
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
		Logger.Error("failed to stop plan", zap.Error(err))
	}
}

func (r *masterRunner) SubscribeReport(fns ...ReportHandleFunc) {
	for _, fn := range fns {
		r.eventbus.subscribeReport(fn)
	}
}

var _ LocalRunner = (*localRunner)(nil)

func NewLocalRunner() LocalRunner {
	runner := &localRunner{
		master: newMasterRunner(),
		slave:  newSlaveRunner(),
	}

	go func(r *masterRunner) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-sigs
		Logger.Warn("caught quit signal, try to shutdown ultron master server", zap.String("signal", sig.String()))

		if r.scheduler != nil {
			if err := r.scheduler.stop(false); err != nil {
				Logger.Error("failed to interrupt current test plan", zap.Error(err))
				os.Exit(1)
			}
			r.eventbus.close()
		}
		os.Exit(0)
	}(runner.master)
	return runner
}

func (lr *localRunner) Launch() error {
	if err := lr.master.Launch(); err != nil {
		return err
	}
	return lr.slave.Connect(":2021", grpc.WithInsecure())
}

func (lr *localRunner) Assign(t Task) {
	lr.slave.Assign(t)
}

func (lr *localRunner) SubscribeResult(fns ...ResultHandleFunc) {
	lr.slave.SubscribeResult(fns...)
}

func (lr *localRunner) SubscribeReport(fns ...ReportHandleFunc) {
	lr.master.SubscribeReport(fns...)
}

func (lr *localRunner) StartPlan(p Plan) error {
	return lr.master.StartPlan(p)
}

func (lr *localRunner) StopPlan() {
	lr.master.StopPlan()
}
