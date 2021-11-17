package ultron

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type (
	// Attacker 事务接口
	Attacker interface {
		Name() string
		Fire(context.Context) error
	}
	// HTTPPrepareFunc 构造http.Request函数，需要调用方定义，由HTTPAttacker来发送
	HTTPPrepareFunc func() (*http.Request, error)

	// HTTPCheckFunc http.Response校验函数，可由调用方自定义，如果返回error，则视为请求失败
	HTTPCheckFunc func(*http.Response, []byte) error

	// HTTPAttacker 内置net/http库对Attacker的实现
	HTTPAttacker struct {
		client      *http.Client
		name        string
		prepareFunc HTTPPrepareFunc
		checkFuncs  []HTTPCheckFunc
	}

	// HTTPAttackerOption HTTPAttacker配置项
	HTTPAttackerOption func(*HTTPAttacker)
)

const (
	defaultUserAgent = "github.com/wosai/ultron"
)

var (
	// defaultHTTPClient 默认http.Client
	// http://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/
	defaultHTTPClient = &http.Client{
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   1000,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	_ Attacker = (*HTTPAttacker)(nil)
)

func NewHTTPAttacker(name string) *HTTPAttacker {
	return &HTTPAttacker{
		client:     defaultHTTPClient,
		name:       name,
		checkFuncs: make([]HTTPCheckFunc, 0),
	}
}

func (ha *HTTPAttacker) Name() string {
	return ha.name
}

func (ha *HTTPAttacker) Fire(ctx context.Context) error {
	if ha.prepareFunc == nil {
		panic("call Apply(WithPrepareFunc()) first")
	}

	req, err := ha.prepareFunc()
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	// change user agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	}

	res, err := ha.client.Do(req)
	if err != nil {
		return err
	}

	if len(ha.checkFuncs) == 0 {
		io.Copy(io.Discard, res.Body) // no checker defined, discard body
		return res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	res.Body.Close()

	for _, check := range ha.checkFuncs {
		if err = check(res, body); err != nil {
			return err
		}
	}
	return nil
}

func (ha *HTTPAttacker) Apply(opts ...HTTPAttackerOption) {
	for _, opt := range opts {
		opt(ha)
	}
}

func WithClient(client *http.Client) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		h.client = client
	}
}

func WithPrepareFunc(prepare HTTPPrepareFunc) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		if prepare == nil {
			panic("invalid HTTPPrepareFunc")
		}
		h.prepareFunc = prepare
	}
}

func WithCheckFuncs(checks ...HTTPCheckFunc) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		for _, check := range checks {
			if check == nil {
				panic("invalid HTTPCheckFunc")
			}
		}
		h.checkFuncs = append(h.checkFuncs, checks...)
	}
}

func WithDisableKeepAlives(disbale bool) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		if tran, ok := h.client.Transport.(*http.Transport); ok {
			tran.DisableKeepAlives = disbale
		}
	}
}

func WithTimeout(t time.Duration) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		h.client.Timeout = t
	}
}

func WithProxy(proxy func(*http.Request) (*url.URL, error)) HTTPAttackerOption {
	return func(h *HTTPAttacker) {
		if transport, ok := h.client.Transport.(*http.Transport); ok {
			transport.Proxy = proxy
		}
	}
}

// CheckHTTPStatusCode 检查状态码是否>=400, 如果是则视为请求失败
func CheckHTTPStatusCode(res *http.Response, body []byte) error {
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return nil
}
