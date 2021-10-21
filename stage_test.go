package ultron

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStageConfigurations(t *testing.T) {
	conf := FixedUserStage{
		FixedUsers: FixedUsers{},
		UniformRandomTimer: UniformRandomTimer{
			MinWait: 2 * time.Second,
			MaxWait: 3 * time.Second,
		},
		UniversalExitConditions: UniversalExitConditions{
			Requests: 1100,
			Duration: 5 * time.Minute,
		},
	}

	data, err := json.Marshal(&conf)
	assert.Nil(t, err)
	log.Println(string(data))
}
