package ultron

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockAttacker struct{}

func (fs *mockAttacker) Name() string {
	return "fake"
}

func (fs *mockAttacker) Fire(ctx context.Context) error {
	req, _ := http.NewRequest(http.MethodGet, "https://www.google.com", nil)
	_ = req.WithContext(ctx)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func BenchmarkFakeAttacker(b *testing.B) {
	attacker := &mockAttacker{}

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			attacker.Fire(context.Background())
		}
	})
}

func BenchmarkHTTPAttacker_Fire(b *testing.B) {
	attacker := NewHTTPAttacker("http-benchmark")
	attacker.Apply(
		WithPrepareFunc(func() (*http.Request, error) { return http.NewRequest(http.MethodGet, "https://www.google.com", nil) }),
	)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if err := attacker.Fire(context.Background()); err != nil {
				Logger.Error("occur error", zap.Error(err))
			}
		}
	})
}

func TestHTTPAttacker_Fire(t *testing.T) {
	attacker := NewHTTPAttacker("http")
	attacker.Apply(
		WithPrepareFunc(func() (*http.Request, error) {
			return http.NewRequest(http.MethodGet, "https://httpbin.org/user-agent", nil)
		}),
		WithCheckFuncs(
			CheckHTTPStatusCode,
			func(r *http.Response, b []byte) error {
				Logger.Info("body", zap.ByteString("body", b))
				return nil
			}),
	)
	err := attacker.Fire(context.Background())
	assert.Nil(t, err)
}
