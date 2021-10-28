package ultron

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/wosai/ultron/v2/pkg/genproto"
)

type (
	// Timer 延时器
	Timer interface {
		Sleep()
	}

	NamedTimer interface {
		Timer
		Name() string
	}

	// UniformRandomTimer 平均随机数
	UniformRandomTimer struct {
		MinWait time.Duration `json:"min_wait,omitempty"`
		MaxWait time.Duration `json:"max_wait,omitempty"`
	}

	// GaussianRandomTimer 高斯分布
	GaussianRandomTimer struct {
		StdDev      float64 `json:"std_dev"`      // 标准差
		DesiredMean float64 `json:"desired_mean"` // 期望均值
	}

	// NonstopTimer 不中断
	NonstopTimer struct{}

	TimerConverter struct {
		convertDTOFuncs map[string]ConvertDTOFunc
	}

	ConvertDTOFunc func([]byte) (Timer, error)
)

var (
	defaultTimerConverter *TimerConverter
)

func (urt *UniformRandomTimer) Sleep() {
	if urt.MaxWait > 0 {
		time.Sleep(urt.MinWait + time.Duration(rand.Int63n(int64(urt.MaxWait-urt.MinWait)+1)))
	}
}

func (urt *UniformRandomTimer) Name() string {
	return "uniform-random-timer"
}

func (grt *GaussianRandomTimer) Sleep() {
	t := time.Duration(rand.NormFloat64()*grt.StdDev + grt.DesiredMean)
	if t > 0 {
		time.Sleep(t * time.Millisecond)
	}
}

func (grt *GaussianRandomTimer) Name() string {
	return "gaussion-random-timer"
}

func (ns NonstopTimer) Sleep() {}

func (ns NonstopTimer) Name() string {
	return "non-stop-timer"
}

func newTimeConveter() *TimerConverter {
	return &TimerConverter{
		convertDTOFuncs: map[string]ConvertDTOFunc{
			"non-stop-timer": func([]byte) (Timer, error) { return NonstopTimer{}, nil },
			"gaussion-random-timer": func(data []byte) (Timer, error) {
				t := new(GaussianRandomTimer)
				err := json.Unmarshal(data, t)
				return t, err
			},
			"uniform-random-timer": func(data []byte) (Timer, error) {
				t := new(UniformRandomTimer)
				err := json.Unmarshal(data, t)
				return t, err
			},
		},
	}
}

func (tc *TimerConverter) ConvertDTO(dto *genproto.TimerDTO) (Timer, error) {
	fn, ok := tc.convertDTOFuncs[dto.Type]
	if !ok {
		return nil, errors.New("cannot find convert func")
	}
	return fn(dto.Timer)
}

func (tc *TimerConverter) ConvertTimer(t Timer) (*genproto.TimerDTO, error) {
	nt, ok := t.(NamedTimer)
	if !ok {
		return nil, errors.New("cannot convert timer")
	}

	data, err := json.Marshal(nt)
	if err != nil {
		return nil, err
	}
	return &genproto.TimerDTO{Type: nt.Name(), Timer: data}, nil
}

func init() {
	defaultTimerConverter = newTimeConveter()
}
