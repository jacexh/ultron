package master

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/internal/eventbus"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	// scheduler master进程任务调度者
	scheduler struct {
		ctx        context.Context
		cancel     context.CancelFunc
		plan       *plan
		supervisor *slaveSupervisor
		eventbus   ultron.ReportBus
	}
)

func NewScheduler() *scheduler {
	return &scheduler{
		supervisor: newSlaveSupervisor(),
		eventbus:   eventbus.DefaultEventBus,
	}
}

func (s *scheduler) start(plan *plan) error {
	if plan == nil {
		return errors.New("cannot start with an empty plan")
	}
	if err := plan.check(); err != nil {
		return err
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())

	if err := s.supervisor.StartNewPlan(s.ctx, plan.Name()); err != nil {
		return err
	}

	_, _, stage, err := s.plan.stopCurrentAndStartNext(-1, statistics.SummaryReport{})
	if err != nil {
		return err
	}

	if err := s.supervisor.NextStage(s.ctx, stage.GetStrategy(), stage.GetTimer()); err != nil {
		return err
	}
	return nil
}

func (s *scheduler) stop(done bool) error {
	if !done {
		s.plan.interrupt()
	}

	var err error
	if err = s.supervisor.Stop(s.ctx, done); err != nil {
		ultron.Logger.Warn("failed to stop slaves", zap.Error(err))
	}
	s.cancel()
	ultron.Logger.Info("canceled all master running jobs")

	report, aggErr := s.supervisor.Aggregate(true)
	switch {
	case err == nil && aggErr != nil:
		return aggErr

	case err != nil && aggErr == nil:
		return err

	case err != nil && aggErr != nil:
		return fmt.Errorf("recent error: %w \tlast error:%s", aggErr, err.Error())

	default:
		s.eventbus.PublishReport(report)
		return nil
	}
}

func (s *scheduler) nextStage(stage ultron.Stage) error {
	return s.supervisor.NextStage(s.ctx, stage.GetStrategy(), stage.GetTimer())
}

// patrol scheduler核心逻辑
func (s *scheduler) patrol(every time.Duration) error {
	ticker := time.NewTimer(every)
	defer ticker.Stop()

	plan := s.plan

	if plan.Status() != ultron.StatusRunning {
		return errors.New("failed to patrol cause the plan is not running")
	}

	stageIndex, _ := plan.Current()
patrol:
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()

		case <-ticker.C:
			report, err := s.supervisor.Aggregate(false)
			if err != nil {
				ultron.Logger.Warn("failed to aggregate stats report", zap.Error(err))
				continue patrol
			}
			s.eventbus.PublishReport(report)

			stopped, next, stage, err := plan.stopCurrentAndStartNext(stageIndex, report)
			switch {
			case err != nil && errors.Is(err, ultron.ErrPlanClosed) && stopped: // 当前在最后一个阶段并且执行完成了
				s.stop(true) // TODO： 是否还要做点什么？不做的话会拿到下一次聚合报告？
				return nil

			case err != nil && errors.Is(err, ultron.ErrPlanClosed) && !stopped: // 计划早已经结束，不干了
				ultron.Logger.Info("this plan is complete, stop patrol")
				return nil

			case err != nil && !errors.Is(err, ultron.ErrPlanClosed):
				ultron.Logger.Error("occur error on checking the test plan", zap.Error(err))
				continue patrol

			case err == nil && stopped: // 下一阶段
				if err := s.nextStage(stage); err != nil {
					ultron.Logger.Error("failed to send the configurations of next stage to slaves", zap.Error(err))
				}
				stageIndex = next

			default: // 继续巡查
			}
		}
	}
}
