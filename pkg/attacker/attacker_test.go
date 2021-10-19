package attacker

import (
	"context"
	"net/http"
	"testing"
)

func BenchmarkHTTPAttacker_Fire(b *testing.B) {
	attacker := NewHTTPAttacker("http-benchmark")
	attacker.Apply(
		WithPrepareFunc(func() (*http.Request, error) { return http.NewRequest(http.MethodGet, "https://www.google.com", nil) }),
	)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			attacker.Fire(context.Background())
		}
	})
}
