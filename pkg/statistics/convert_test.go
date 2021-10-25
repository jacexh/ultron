package statistics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertStatisticianGroup(t *testing.T) {
	entity := NewStatisticianGroup()
	entity.Record(AttackResult{Name: "foobar", Duration: 3 * time.Millisecond})
	dto, err := ConvertStatisticianGroup(entity)
	assert.Nil(t, err)
	assert.EqualValues(t, dto.Container["foobar"].Requests, 1)
	assert.EqualValues(t, dto.Container["foobar"].MinResponseTime, 3*time.Millisecond)
	assert.EqualValues(t, dto.Container["foobar"].MinResponseTime, dto.Container["foobar"].MaxResponseTime)
}
