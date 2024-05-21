package ultron

import (
	"context"
	"time"
)

type (
	// executorSharedContext is designed to carry the fire-scoped values.
	executorSharedContext struct {
		parentCtx context.Context
		counter   int32
		resources map[int32]map[string]interface{}
	}
)

func newExecutorSharedContext(ctx context.Context) context.Context {
	return &executorSharedContext{
		parentCtx: ctx,
		resources: make(map[int32]map[string]interface{}),
	}
}

func (m *executorSharedContext) Deadline() (time.Time, bool) {
	return m.parentCtx.Deadline()
}

func (m *executorSharedContext) Done() <-chan struct{} {
	return m.parentCtx.Done()
}

func (m *executorSharedContext) Err() error {
	return m.parentCtx.Err()
}

func (m *executorSharedContext) Value(key interface{}) interface{} {
	return m.parentCtx.Value(key)
}

func FromContext(ctx context.Context, key string) (interface{}, bool) {
	entity, ok := ctx.(*executorSharedContext)
	if ok {
		if res, ok := entity.resources[entity.counter]; ok {
			return res[key], true
		}
	}
	return nil, false
}

func StoreInContext(ctx context.Context, key string, value interface{}) bool {
	if entity, ok := ctx.(*executorSharedContext); ok {
		if _, ok := entity.resources[entity.counter]; !ok {
			entity.resources[entity.counter] = make(map[string]interface{})
		}
		entity.resources[entity.counter][key] = value
		return true
	}
	return false
}

func ClearStorageInContext(ctx context.Context) {
	if entity, ok := ctx.(*executorSharedContext); ok {
		delete(entity.resources, entity.counter)
	}
}

func AllocateStorageInContext(ctx context.Context) context.Context {
	if entity, ok := ctx.(*executorSharedContext); ok {
		entity.counter++
	} else {
		ctx = newExecutorSharedContext(ctx)
	}
	return ctx
}
