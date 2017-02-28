package ultron

import (
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	// DefaultFastHTTTPClient 默认fasthttp客户端
	DefaultFastHTTTPClient = &fasthttp.Client{
		MaxConnsPerHost:     1000,
		MaxIdleConnDuration: time.Second * 30,
		ReadTimeout:         time.Second * 60,
		WriteTimeout:        time.Second * 30,
	}
)

type (
	// FastHTTPRequest 结构体
	FastHTTPRequest struct {
		client     *fasthttp.Client
		name       string
		parent     *TaskSet
		Prepare    func() *fasthttp.Request
		CheckChain []func(*fasthttp.Response) error
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

// SetTaskSet 设置task
func (f *FastHTTPRequest) SetTaskSet(t *TaskSet) {
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
