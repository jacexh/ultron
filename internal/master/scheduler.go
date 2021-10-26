package master

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/internal/eventbus"
)

type (
	// Scheduler master进程任务调度者
	Scheduler struct {
		plan       ultron.Plan
		planCtx    context.Context
		planCancel context.CancelFunc
		supervisor *slaveSupervisor
		eventbus   ultron.ReportBus
		ticker     *time.Ticker
		mu         sync.Mutex
	}
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		supervisor: newSlaveSupervisor(),
		eventbus:   eventbus.DefaultEventBus,
	}
}

func (s *Scheduler) CreateNewPlan() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.plan.Status() == ultron.StatusRunning {
		return errors.New("failed to create a new plan until stop current running plan")
	}

	if s.planCancel != nil {
		s.planCancel()
	}
	s.planCtx, s.planCancel = context.WithCancel(context.Background())
	s.plan = NewPlan()
	return nil
}

// func (s *Scheduler) Start() error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	if s.plan == nil {
// 		return errors.New("cannot start empty plan")
// 	}
// 	if err := s.plan.Start(); err != nil {
// 		return err
// 	}

// 	_, stageIndex, stage, err := s.plan.StopCurrentAndStartNext(-1, statistics.SummaryReport{})
// 	if err != nil {
// 		return err
// 	}
// 	stage.GetStrategy().Spawn()
// }
