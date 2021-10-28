package types

import "context"

// Attacker 定义一个事务、请求，需要确保实现上是goroutine-safe
type Attacker interface {
	Name() string
	Fire(context.Context) error
}
