package ultron

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
)

func TestSlaveAgent_close(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.send(&genproto.SubscribeResponse{Type: genproto.EventType_PING})
			if err := agent.close(); err != nil {
				log.Info("no consumer for the slave agent", zap.Error(err))
			}
		}()
	}
	wg.Wait()
}
