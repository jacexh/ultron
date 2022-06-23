package ultron

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := AllocateStorageInContext(context.Background())
	StoreInContext(ctx, "foo", "bar")
	StoreInContext(ctx, "hello", "world")
	es, ok := ctx.(*executorSharedContext)
	assert.True(t, ok)
	assert.EqualValues(t, es.resources[0], map[string]interface{}{"foo": "bar", "hello": "world"})

	ClearStorageInContext(ctx)
	assert.EqualValues(t, len(es.resources[0]), 0)

	ctx = AllocateStorageInContext(ctx)
	StoreInContext(ctx, "foo", "bar")
	StoreInContext(ctx, "hello", "world")
	assert.EqualValues(t, es.resources[1], map[string]interface{}{"foo": "bar", "hello": "world"})
}
