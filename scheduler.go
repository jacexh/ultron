package ultron

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		eventbus   reportBus
	}
)

func newScheduler(sup *slaveSupervisor) *scheduler {
	return &scheduler{
		supervisor: sup,
		eventbus:   defaultEventBus,
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
	s.plan = plan
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
	if s.plan == nil {
		return nil
	}

	ps := s.plan.Status()

	switch {
	case ps == StatusFinished && done: // 正常结束

	case ps == StatusRunning && !done: // 被中断
		s.plan.interrupt()

	case !done && (ps == StatusReady || ps == StatusFinished || ps == StatusInterrupted):
		return nil

	default:
		Logger.Warn("unexpected parameters", zap.Bool("done", done), zap.Int("plan_status", int(ps)))
		return errors.New("unknown status")
	}

	var err error
	if err = s.supervisor.Stop(s.ctx, done); err != nil {
		Logger.Warn("failed to stop slaves", zap.Error(err))
	}
	s.cancel()
	Logger.Info("canceled all running jobs")

	report, aggErr := s.supervisor.Aggregate(true, statistics.Tag{Key: planKey, Value: s.plan.Name()})
	switch {
	case err == nil && aggErr != nil:
		return aggErr

	case err != nil && aggErr == nil:
		return err

	case err != nil && aggErr != nil:
		return fmt.Errorf("recent error: %w last error:%s", aggErr, err.Error())

	default:
		s.eventbus.publishReport(report)
		return nil
	}
}

func (s *scheduler) nextStage(stage Stage) error {
	return s.supervisor.NextStage(s.ctx, stage.GetStrategy(), stage.GetTimer())
}

// patrol scheduler核心逻辑
func (s *scheduler) patrol(every time.Duration) error {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	plan := s.plan

	if plan.Status() != StatusRunning {
		return errors.New("failed to patrol cause the plan is not running")
	}

	stageIndex, _ := plan.Current()
patrol:
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()

		case <-ticker.C:
			report, err := s.supervisor.Aggregate(false, statistics.Tag{Key: planKey, Value: plan.Name()})
			if err != nil {
				Logger.Warn("failed to aggregate stats report", zap.Error(err))
				continue patrol
			}
			s.eventbus.publishReport(report)

			stopped, next, stage, err := plan.stopCurrentAndStartNext(stageIndex, report)
			switch {
			case err != nil && errors.Is(err, ErrPlanClosed) && stopped: // 当前在最后一个阶段并且执行完成了，此时plan已经完成
				Logger.Info("current test plan is complete")
				s.stop(true) // TODO： 是否还要做点什么？不做的话会拿到下一次聚合报告？
				return nil

			case err != nil && errors.Is(err, ErrPlanClosed) && !stopped: // 计划早已经结束，不干了
				Logger.Info("this plan is complete, stop patrol")
				return nil

			case err != nil && !errors.Is(err, ErrPlanClosed):
				Logger.Error("occur error on checking the test plan", zap.Error(err))
				continue patrol

			case err == nil && stopped: // 下一阶段
				Logger.Info("start the next stage")
				if err := s.nextStage(stage); err != nil {
					Logger.Error("failed to send the configurations of next stage to slaves", zap.Error(err))
				}
				stageIndex = next

			default: // 继续巡查
			}
		default:
			continue
		}
	}
}
