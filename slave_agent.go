package ultron

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
)

type (
	// SlaveAgent 定义master侧的slave对象
	SlaveAgent interface {
		ID() string
		Extras() map[string]string
	}

	slaveAgent struct {
		slaveID string
		extras  map[string]string // todo: 之后再实现
		input   chan *genproto.SubscribeResponse
		closed  uint32
	}
)

var (
	_ SlaveAgent = (*slaveAgent)(nil)
)

func newSlaveAgent(req *genproto.SubscribeRequest) *slaveAgent {
	return &slaveAgent{
		slaveID: req.SlaveId,
		extras:  req.Extras,
		input:   make(chan *genproto.SubscribeResponse, 1),
	}
}

func (sa *slaveAgent) ID() string {
	return sa.slaveID
}

func (sa *slaveAgent) Extras() map[string]string {
	ret := make(map[string]string)
	for k, v := range sa.extras {
		ret[k] = v
	}
	return ret
}

func (sa *slaveAgent) close() error {
	if atomic.CompareAndSwapUint32(&sa.closed, 0, 1) {
		select {
		case sa.input <- &genproto.SubscribeResponse{Type: genproto.EventType_DISCONNECT}:
			close(sa.input)
		case <-time.After(3 * time.Second):
			return fmt.Errorf("the input channel is blocked: %s", sa.ID())
		}
	}
	return nil
}

// send 返回是否发送（不代表发送成功）
func (sa *slaveAgent) send(event *genproto.SubscribeResponse) error {
	if atomic.LoadUint32(&sa.closed) == 0 {
		sa.input <- event
		return nil
	}
	return fmt.Errorf("slave agent is closed, cannot send event out: %d", event.Type)
}

func (sa *slaveAgent) keepAlives() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := sa.send(&genproto.SubscribeResponse{Type: genproto.EventType_PING}); err != nil {
			log.Info("the slave agent is closed, stop the ticker", zap.String("slave_id", sa.ID()), zap.Error(err))
			return
		}
	}
}
