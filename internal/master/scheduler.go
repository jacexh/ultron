package master

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/internal/eventbus"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	// Scheduler master进程任务调度者
	Scheduler struct {
		planCtx    context.Context
		planCancel context.CancelFunc
		plan       ultron.Plan
		supervisor *slaveSupervisor
		eventbus   ultron.ReportBus
		mu         sync.Mutex
	}
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		supervisor: newSlaveSupervisor(),
		eventbus:   eventbus.DefaultEventBus,
	}
}

func (s *Scheduler) CreateNewPlan(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.plan.Status() == ultron.StatusRunning {
		return errors.New("failed to create a new plan until stop current running plan")
	}

	if s.planCancel != nil {
		s.planCancel()
	}
	s.planCtx, s.planCancel = context.WithCancel(context.Background())
	s.plan = NewPlan(name)
	return nil
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.plan == nil {
		return errors.New("cannot start empty plan")
	}
	if err := s.plan.Start(); err != nil {
		return err
	}

	if err := s.supervisor.StartNewPlan(); err != nil {
		return err
	}

	_, _, stage, err := s.plan.StopCurrentAndStartNext(-1, statistics.SummaryReport{})
	if err != nil {
		return err
	}

	if err := s.supervisor.Refresh(stage.GetStrategy(), stage.GetTimer()); err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) FinishPlan() {

}

func (s *Scheduler) nextStage(stage ultron.Stage) {}

func (s *Scheduler) patrol(ctx context.Context, every time.Duration, plan ultron.Plan) error {
	ticker := time.NewTimer(every)
	defer ticker.Stop()

	if plan.Status() != ultron.StatusRunning {
		return errors.New("failed to patrol cause the plan is not running")
	}

	stageIndex, _ := plan.Current()
patrol:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			report, err := s.supervisor.Aggregate(true)
			if err != nil {
				ultron.Logger.Warn("failed to aggregate stats report", zap.Error(err))
				continue patrol
			}
			stopped, next, stage, err := plan.StopCurrentAndStartNext(stageIndex, report)

			switch {
			case err != nil && errors.Is(err, ultron.ErrPlanClosed) && stopped: // 当前在最后一个阶段并且执行完成了
				s.FinishPlan()

			case err != nil && errors.Is(err, ultron.ErrPlanClosed) && !stopped: // 计划早已经结束，不干了
				ultron.Logger.Info("this plan is complete, stop patrol")
				return nil

			case err != nil && !errors.Is(err, ultron.ErrPlanClosed):
				ultron.Logger.Error("occur error on checking the test plan", zap.Error(err))
				continue patrol

			case err == nil && stopped: // 下一阶段
				s.nextStage(stage)
				stageIndex = next

			default: // 继续巡查
			}
		}
	}
}
