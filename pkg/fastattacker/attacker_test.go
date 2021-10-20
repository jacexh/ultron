package fastattacker

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestFastHTTPAttacker_Fire(t *testing.T) {
	attacker := NewFastHTTPAttacker("foobar")
	attacker.Apply(
		WithPrepareFunc(func(r *fasthttp.Request) error {
			r.SetRequestURI("https://httpbin.org/user-agent")
			r.Header.SetMethod(fasthttp.MethodGet)
			return nil
		}),
		WithCheckFunc(CheckHTTPStatusCode, func(r *fasthttp.Response) error {
			log.Println(string(r.Body()))
			return nil
		}),
	)
	err := attacker.Fire(context.Background())
	if err != nil {
		t.FailNow()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = attacker.Fire(ctx)
	if err == nil {
		t.FailNow()
	}
	if !errors.Is(err, context.Canceled) {
		t.FailNow()
	}
}
