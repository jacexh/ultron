package ultron

import (
	"encoding/json"
	"math/rand"
	"time"
)

type (
	// Timer 延时器
	Timer interface {
		Sleep()
	}

	TimerDTO interface {
		Timer
		json.Marshaler
		json.Unmarshaler
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
)

func (urt UniformRandomTimer) Sleep() {
	if urt.MaxWait > 0 {
		time.Sleep(urt.MinWait + time.Duration(rand.Int63n(int64(urt.MaxWait-urt.MinWait)+1)))
	}
}

func (grt GaussianRandomTimer) Sleep() {
	t := time.Duration(rand.NormFloat64()*grt.StdDev + grt.DesiredMean)
	if t > 0 {
		time.Sleep(t * time.Millisecond)
	}
}

func (ns NonstopTimer) Sleep() {}
