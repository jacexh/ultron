package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimerConverterUniformRandomTimer(t *testing.T) {
	t1 := &UniformRandomTimer{MinWait: 100 * time.Millisecond, MaxWait: 200 * time.Microsecond}
	converter := newTimeConveter()
	dto, err := converter.ConvertTimer(t1)
	assert.Nil(t, err)
	t2, err := converter.ConvertDTO(dto)
	assert.Nil(t, err)
	assert.EqualValues(t, t1, t2)
}

func TestTimerConverterGaussianRandomTimer(t *testing.T) {
	t1 := &GaussianRandomTimer{DesiredMean: 100.11, StdDev: 15.3}
	converter := newTimeConveter()
	dto, err := converter.ConvertTimer(t1)
	assert.Nil(t, err)
	t2, err := converter.ConvertDTO(dto)
	assert.Nil(t, err)
	assert.EqualValues(t, t1, t2)
}

func TestTimerConverterNonStopTimer(t *testing.T) {
	t1 := NonstopTimer{}
	converter := newTimeConveter()
	dto, err := converter.ConvertTimer(t1)
	assert.Nil(t, err)
	t2, err := converter.ConvertDTO(dto)
	assert.Nil(t, err)
	assert.EqualValues(t, t1, t2)
}
