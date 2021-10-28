package master

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/internal/eventbus"
	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	MasterRunner struct {
		scheduler  *scheduler
		plan       ultron.Plan
		eventbus   *eventbus.IEventBus
		supervisor *slaveSupervisor
		mu         sync.RWMutex
	}
)

var (
	_ ultron.MasterRunner = (*MasterRunner)(nil)
)

func NewMasterRunner() ultron.MasterRunner {
	return &MasterRunner{
		eventbus: eventbus.DefaultEventBus,
	}
}

// Launch 主线程，如果发生错误则关闭
func (r *MasterRunner) Launch(con ultron.RunnerConfig, opts ...grpc.ServerOption) error {
	if con.GRPCAddr == "" {
		con.GRPCAddr = ultron.DefaultGRPC
	}
	if con.RESTAddr == "" {
		con.RESTAddr = ultron.DefaultREST
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
	r.eventbus.SubscribeReport(eventbus.PrintReportToConsole(os.Stdout))
	r.eventbus.Start()
	log.Info("report bus is working")

	block := make(chan error, 1)
	go func() {
		block <- grpcServer.Serve(lis)
	}()
	err = <-block
	log.Fatal("ultron runner is shutdown", zap.Error(err))
	return err
}

func (r *MasterRunner) StartPlan(p ultron.Plan) error {
	r.mu.Lock()

	if r.plan != nil && r.plan.Status() == ultron.StatusRunning {
		r.mu.Unlock()
		err := errors.New("cannot start a new plan before shutdown current running plan")
		log.Error("failed to start a new plan", zap.Error(err))
		return err
	}
	scheduler := NewScheduler()
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

func (r *MasterRunner) StopPlan() {
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

func (r *MasterRunner) SubscribeReport(fns ...ultron.ReportHandleFunc) {
	for _, fn := range fns {
		r.eventbus.SubscribeReport(fn)
	}
}

func init() {
	ultron.RegisterMasterRunnerBuilder(NewMasterRunner)
}
