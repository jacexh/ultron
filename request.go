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
	// Request .
	Request interface {
		SetParent(*TaskSet)
		Name() string
		Fire() error
	}

	// FastHTTPRequest 结构体
	FastHTTPRequest struct {
		client     *fasthttp.Client
		name       string
		parent     *TaskSet
		Prepare    func() *fasthttp.Request
		CheckChain []func(*fasthttp.Response) error
	}

	// HTTPRequest net/http request
	HTTPRequest struct {
		client     *http.Client
		name       string
		parent     *TaskSet
		Prepare    func() *http.Request
		CheckChain []func(*http.Response, []byte) error
	}
)

var (
	// DefaultFastHTTTPClient 默认fasthttp客户端
	DefaultFastHTTTPClient = &fasthttp.Client{
		MaxConnsPerHost:     1000,
		MaxIdleConnDuration: time.Second * 30,
		ReadTimeout:         time.Second * 60,
		WriteTimeout:        time.Second * 30,
	}

	// DefaultHTTPClient default net/http client
	DefaultHTTPClient = &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 1000,
		},
	}
)

// NewFastHTTPRequest 创建fasthttp实例
func NewFastHTTPRequest(n string) *FastHTTPRequest {
	return &FastHTTPRequest{
		client: DefaultFastHTTTPClient,
		name:   n,
		CheckChain: []func(*fasthttp.Response) error{
			func(r *fasthttp.Response) error { return checkStatusCode(r.StatusCode()) },
		},
	}
}

// Name 获取http请求名称
func (f *FastHTTPRequest) Name() string {
	return f.name
}

// SetParent 设置task
func (f *FastHTTPRequest) SetParent(t *TaskSet) {
	f.parent = t
}

// Fire 发起请求
func (f *FastHTTPRequest) Fire() error {
	if f.Prepare == nil {
		panic(errors.New("please imple Prepare() method"))
	}
	response := fasthttp.AcquireResponse()
	request := f.Prepare()

	if err := f.client.Do(request, response); err != nil {
		return err
	}
	response.Body()

	for _, f := range f.CheckChain {
		err := f(response)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewHTTPRequest create new HTTPRequest instance
func NewHTTPRequest(n string) *HTTPRequest {
	return &HTTPRequest{
		client: DefaultHTTPClient,
		name:   n,
		CheckChain: []func(*http.Response, []byte) error{
			func(r *http.Response, b []byte) error { return checkStatusCode(r.StatusCode) },
		},
	}
}

// Name return the name of HTTPRequest
func (h *HTTPRequest) Name() string {
	return h.name
}

// Fire send to request and read response
func (h *HTTPRequest) Fire() error {
	if h.Prepare == nil {
		panic(errors.New("please impl Prepare()"))
	}
	resp, err := h.client.Do(h.Prepare())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	for _, check := range h.CheckChain {
		err := check(resp, body)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetParent set taskset
func (h *HTTPRequest) SetParent(t *TaskSet) {
	h.parent = t
}
