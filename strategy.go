package ultron

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// AttackStrategyCommander 压测策略
	AttackStrategyCommander interface {
		Open(context.Context, *Task) <-chan statistics.AttackResult
		Command(AttackStrategy, Timer)
		Close()
	}

	// AttackStrategy 压测策略描述
	AttackStrategy interface {
		Spawn() []*RampUpStep
		Switch(next AttackStrategy) []*RampUpStep
		Split(int) []AttackStrategy
	}

	// RampUpStep 增/降压描述
	RampUpStep struct {
		N        int           // 增、降的数量，>0 为加压， <0为降压
		Interval time.Duration // 间隔时间
	}

	// FixedConcurrentUsers 固定goroutine/线程/用户的并发策略
	FixedConcurrentUsers struct {
		ConcurrentUsers int `json:"conncurrent_users"`        // 并发用户数
		RampUpPeriod    int `json:"ramp_up_period,omitempty"` // 增压周期时长
	}

	AttackStrategyConverter struct {
		convertDTOFunc map[string]AttackStrategyConvertDTOFunc
	}

	AttackStrategyConvertDTOFunc func([]byte) (AttackStrategy, error)

	NamedAttackStrategy interface {
		AttackStrategy
		Name() string
	}
)

var (
	_ AttackStrategy = (*FixedConcurrentUsers)(nil)

	DefaultAttackStrategyConverter *AttackStrategyConverter
)

func (fc *FixedConcurrentUsers) spawn(current, expected, period, interval int) []*RampUpStep {
	var ret []*RampUpStep

	if current == expected {
		return ret
	}

	if period < interval {
		period = interval
	}

	steps := period / interval
	nPerStep := (expected - current) / steps

	if current < expected {
		for current <= expected-nPerStep {
			current += nPerStep
			ret = append(ret, &RampUpStep{N: nPerStep, Interval: time.Duration(interval) * time.Second})
		}
	} else {
		for current >= expected-nPerStep {
			current += nPerStep
			ret = append(ret, &RampUpStep{N: nPerStep, Interval: time.Duration(interval) * time.Second})
		}
	}

	if current != expected {
		ret[len(ret)-1].N += (expected - current)
	}
	return ret
}

// Spawn 增压、降压
func (fc *FixedConcurrentUsers) Spawn() []*RampUpStep {
	if fc.ConcurrentUsers <= 0 {
		panic("the number of concurrent users must be greater than 0")
	}
	return fc.spawn(0, fc.ConcurrentUsers, fc.RampUpPeriod, 1)
}

// Switch 转入下一个阶段
func (fc *FixedConcurrentUsers) Switch(next AttackStrategy) []*RampUpStep {
	n, ok := next.(*FixedConcurrentUsers)
	if !ok {
		panic("cannot switch to different type of AttackStrategyDescriber")
	}
	return fc.spawn(fc.ConcurrentUsers, n.ConcurrentUsers, n.RampUpPeriod, 1)
}

// Split 切分配置
func (fx *FixedConcurrentUsers) Split(n int) []AttackStrategy {
	if n <= 0 {
		panic("bad slices number")
	}
	ret := make([]AttackStrategy, n)
	for i := 0; i < n; i++ {
		ret[i] = &FixedConcurrentUsers{
			ConcurrentUsers: fx.ConcurrentUsers / n,
			RampUpPeriod:    fx.RampUpPeriod,
		}
	}
	if remainder := fx.ConcurrentUsers % n; remainder > 0 {
		for i := 0; i < remainder; i++ {
			ret[i].(*FixedConcurrentUsers).ConcurrentUsers++
		}
	}
	return ret
}

func (fx *FixedConcurrentUsers) Name() string {
	return "fixed-concurrent-users"
}

func newAttackStrategyConverter() *AttackStrategyConverter {
	return &AttackStrategyConverter{
		convertDTOFunc: map[string]AttackStrategyConvertDTOFunc{
			"fixed-concurrent-users": func(data []byte) (AttackStrategy, error) {
				as := new(FixedConcurrentUsers)
				err := json.Unmarshal(data, as)
				return as, err
			},
		},
	}
}

func (c *AttackStrategyConverter) ConvertDTO(dto *genproto.AttackStrategyDTO) (AttackStrategy, error) {
	fn, ok := c.convertDTOFunc[dto.Type]
	if !ok {
		return nil, errors.New("cannot found convertion function")
	}
	return fn(dto.AttackStrategy)
}

func (c *AttackStrategyConverter) ConvertAttackStrategy(as AttackStrategy) (*genproto.AttackStrategyDTO, error) {
	na, ok := as.(NamedAttackStrategy)
	if !ok {
		return nil, errors.New("cannot convert attack strategy")
	}
	data, err := json.Marshal(na)
	if err != nil {
		return nil, err
	}
	return &genproto.AttackStrategyDTO{Type: na.Name(), AttackStrategy: data}, nil
}
