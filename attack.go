package ultron

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
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
	// Attacker .
	Attacker interface {
		TaskSet() *TaskSet
		SetTaskSet(*TaskSet)
		Name() string
		Fire() (int, error)
	}

	// FastHTTPAttacker 结构体
	FastHTTPAttacker struct {
		client     *fasthttp.Client
		name       string
		parent     *TaskSet
		Prepare    func() *fasthttp.Request
		CheckChain []func(*fasthttp.Response) error
	}

	// HTTPAttacker net/http request
	HTTPAttacker struct {
		client     *http.Client
		name       string
		parent     *TaskSet
		Prepare    func() *http.Request
		CheckChain []func(*http.Response, []byte) error
	}
)

var (
	// DefaultFastHTTTPAttackerConfig 默认fasthttp配置
	DefaultFastHTTTPAttackerConfig = &fasthttp.Client{
		MaxConnsPerHost:     1000,
		MaxIdleConnDuration: time.Second * 30,
		ReadTimeout:         time.Second * 60,
		WriteTimeout:        time.Second * 30,
	}

	// DefaultHTTPAttackerConfig 默认配置
	DefaultHTTPAttackerConfig = &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 1000,
		},
	}
)

// NewFastHTTPAttacker 创建fasthttp实例
func NewFastHTTPAttacker(n string) *FastHTTPAttacker {
	return &FastHTTPAttacker{
		client: DefaultFastHTTTPAttackerConfig,
		name:   n,
		CheckChain: []func(*fasthttp.Response) error{
			func(r *fasthttp.Response) error { return checkStatusCode(r.StatusCode()) },
		},
	}
}

// Name 获取http请求名称
func (f *FastHTTPAttacker) Name() string {
	return f.name
}

// SetTaskSet 设置task
func (f *FastHTTPAttacker) SetTaskSet(t *TaskSet) {
	f.parent = t
}

// TaskSet 获取TaskSet
func (f *FastHTTPAttacker) TaskSet() *TaskSet {
	return f.parent
}

// Fire 发起请求
func (f *FastHTTPAttacker) Fire() (int, error) {
	if f.Prepare == nil {
		panic(errors.New("please imple Prepare() method"))
	}
	response := fasthttp.AcquireResponse()
	request := f.Prepare()

	if err := f.client.Do(request, response); err != nil {
		return 0, err
	}
	body := response.Body()

	for _, f := range f.CheckChain {
		err := f(response)
		if err != nil {
			return 0, err
		}
	}
	return len(body), nil
}

// NewHTTPAttacker create new HTTPRequest instance
func NewHTTPAttacker(n string) *HTTPAttacker {
	return &HTTPAttacker{
		client: DefaultHTTPAttackerConfig,
		name:   n,
		CheckChain: []func(*http.Response, []byte) error{
			func(r *http.Response, b []byte) error { return checkStatusCode(r.StatusCode) },
		},
	}
}

// Name return the name of HTTPRequest
func (h *HTTPAttacker) Name() string {
	return h.name
}

// Fire send request and read response
func (h *HTTPAttacker) Fire() (int, error) {
	if h.Prepare == nil {
		panic(errors.New("please impl Prepare()"))
	}
	resp, err := h.client.Do(h.Prepare())
	if err != nil {
		return 0, err
	}

	// defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	for _, check := range h.CheckChain {
		err := check(resp, body)
		if err != nil {
			return len(body), err
		}
	}

	return len(body), nil
}

// SetTaskSet set taskset
func (h *HTTPAttacker) SetTaskSet(t *TaskSet) {
	h.parent = t
}

// TaskSet 获取TaskSet
func (h *HTTPAttacker) TaskSet() *TaskSet {
	return h.parent
}
