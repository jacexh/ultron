package master

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/wosai/ultron/v2/pkg/genproto"
)

func Test_slaveAgent_close(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		func() {
			defer wg.Done()
			agent.send(&genproto.SubscribeResponse{Type: genproto.EventType_PING})
			agent.close()
		}()
	}
	wg.Wait()
}
