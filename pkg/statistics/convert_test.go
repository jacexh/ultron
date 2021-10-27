package statistics

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertStatisticianGroup(t *testing.T) {
	entity := NewStatisticianGroup()
	entity.Record(AttackResult{Name: "foobar", Duration: 3 * time.Millisecond})
	dto, err := ConvertStatisticianGroup(entity)
	assert.Nil(t, err)

	entity1, err := NewStatisticianGroupFromDTO(dto)
	assert.Nil(t, err)
	d1, _ := json.Marshal(entity.Report(true))
	d2, _ := json.Marshal(entity1.Report(true))
	assert.EqualValues(t, d1, d2)
}
