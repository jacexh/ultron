package ultron

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
				Logger.Error("no consumer for the slave agent", zap.Error(err))
			}
		}()
	}
	wg.Wait()
}

func TestSlaveAgent_ID(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})
	assert.True(t, agent.ID() != "")

	extras := agent.Extras()
	assert.EqualValues(t, len(extras), 0)
}

func TestSlaveAgent_closeTimeout(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})
	agent.send(&genproto.SubscribeResponse{})
	err := agent.close()
	assert.NotNil(t, err)
	assert.EqualValues(t, err.Error(), fmt.Sprintf("the input channel is blocked: %s", agent.ID()))
}

func TestSlaveAgent_keepAliveOut(t *testing.T) {
	agent := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: uuid.NewString()})
	go func() {
		<-time.After(1 * time.Second)
		agent.close()
	}()
	agent.keepAlives()
}
