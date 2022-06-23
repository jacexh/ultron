package fastattacker

import (
	"context"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wosai/ultron/v2"
)

type (
	// FastHTTPAttacker Attacker接口的fasthttp实现
	FastHTTPAttacker struct {
		name        string
		client      *fasthttp.Client
		prepareFunc FastHTTPPrepareFunc
		checkFuncs  []FastHTTPCheckFunc
	}

	// FastHTTPPrepareFunc 构造fasthttp.Request的请求
	FastHTTPPrepareFunc func(context.Context, *fasthttp.Request) error
	// FastHTTPCheckFunc fasthttp.Response检查函数
	FastHTTPCheckFunc func(context.Context, *fasthttp.Response) error
	// FastHTTPAttackerOption FastHTTPAttacker的配置项
	FastHTTPAttackerOption func(*FastHTTPAttacker)
)

const defaultUserAgent = "github.com/wosai/ultron"

var (
	defaultFastHTTPClient = &fasthttp.Client{
		Name:                defaultUserAgent,
		MaxConnsPerHost:     1000,
		MaxIdleConnDuration: 30 * time.Second,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
	}

	_ ultron.Attacker = (*FastHTTPAttacker)(nil)
)

func NewFastHTTPAttacker(name string) *FastHTTPAttacker {
	return &FastHTTPAttacker{
		name:       name,
		client:     defaultFastHTTPClient,
		checkFuncs: make([]FastHTTPCheckFunc, 0),
	}
}

func (fa *FastHTTPAttacker) Name() string {
	return fa.name
}

func (fa *FastHTTPAttacker) Fire(ctx context.Context) error {
	if fa.prepareFunc == nil {
		panic("call Apply(WithPrepareFunc()) first")
	}

	ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := fa.prepareFunc(ctx, req)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err = fa.client.Do(req, res); err != nil {
		return err
	}

	for _, check := range fa.checkFuncs {
		if err = check(ctx, res); err != nil {
			return err
		}
	}
	return nil
}

func (fa *FastHTTPAttacker) Apply(opts ...FastHTTPAttackerOption) {
	for _, opt := range opts {
		opt(fa)
	}
}

func WithClient(client *fasthttp.Client) FastHTTPAttackerOption {
	return func(fh *FastHTTPAttacker) {
		fh.client = client
	}
}

func WithCheckFunc(checks ...FastHTTPCheckFunc) FastHTTPAttackerOption {
	return func(fh *FastHTTPAttacker) {
		for _, check := range checks {
			if check == nil {
				panic("invalid FastHTTPCheckFunc")
			}
		}
		fh.checkFuncs = append(fh.checkFuncs, checks...)
	}
}

func WithPrepareFunc(prepare FastHTTPPrepareFunc) FastHTTPAttackerOption {
	return func(fh *FastHTTPAttacker) {
		if prepare == nil {
			panic("invalid FastHTTPPrepareFunc")
		}
		fh.prepareFunc = prepare
	}
}

func WithTimeout(t time.Duration) FastHTTPAttackerOption {
	return func(fh *FastHTTPAttacker) {
		fh.client.ReadTimeout = t
		fh.client.WriteTimeout = t
	}
}

func CheckHTTPStatusCode(_ context.Context, res *fasthttp.Response) error {
	if code := res.StatusCode(); code >= fasthttp.StatusBadRequest {
		return fmt.Errorf("bad status code: %d", code)
	}
	return nil
}
