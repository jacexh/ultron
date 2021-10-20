package ultron

import (
	"math/rand"
	"time"
)

type (
	// Timer 延时器
	Timer interface {
		Sleep()
	}

	// UniformRandomTimer 平均随机数
	UniformRandomTimer struct {
		MinWait time.Duration `json:"min_wait,omitempty"`
		MaxWait time.Duration `json:"max_wait,omitempty"`
	}

	// GaussianRandomTimer 高斯分布
	GaussianRandomTimer struct {
		StdDev      float64 // 标准差
		DesiredMean float64 // 期望均值
	}
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
