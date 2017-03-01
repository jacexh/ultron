package ultron

import (
	"time"
)

const (
	// ZeroDuration 无等待,用于一些特殊判断
	ZeroDuration time.Duration = time.Duration(0)
	// DefaultMinWait 默认最小等待时间
	DefaultMinWait time.Duration = time.Second * 1
	// DefaultMaxWait 默认最大等待时间
	DefaultMaxWait time.Duration = time.Second * 5
	// DefaultConcurrence 默认并发数
	DefaultConcurrence = 100
)

type (
	// Query .
	Query interface {
		SetTaskSet(*TaskSet)
		Name() string
		Fire() error
	}
)
