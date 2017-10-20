package ultron

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// Attacker 定义一个事务、请求，需要确保该对象是Goroutine-safe
	Attacker interface {
		Name() string
		Fire() error
	}

	// HTTPPrepareFunc 构造http.Request函数，需要调用方定义，由HTTPAttacker来发送
	HTTPPrepareFunc func() (*http.Request, error)
	// HTTPResponseCheck http.Response校验函数，可由调用方自定义，如果返回error，则视为请求失败
	HTTPResponseCheck func(resp *http.Response, body []byte) error

	// HTTPAttacker http协议的Attacker实现
	HTTPAttacker struct {
		Client     *http.Client
		Prepare    HTTPPrepareFunc
		name       string
		CheckChain []HTTPResponseCheck
	}
)

var (
	// DefaultHTTPClient 默认http.Client
	// Todo: 后续做优化
	DefaultHTTPClient = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: false,
		},
	}
)

// NewHTTPAttacker 创建一个新的HTTPAttacker对象
func NewHTTPAttacker(n string, p HTTPPrepareFunc, check ...HTTPResponseCheck) *HTTPAttacker {
	return &HTTPAttacker{
		Client:     DefaultHTTPClient,
		Prepare:    p,
		name:       n,
		CheckChain: check,
	}
}

// Name 返回HTTPAttacker的名称
func (ha *HTTPAttacker) Name() string {
	return ha.name
}

// Fire HTTPAttacker发起一次请求
func (ha *HTTPAttacker) Fire() error {
	if ha.Prepare == nil {
		panic("please implement Prepare() first")
	}

	req, err := ha.Prepare()
	if err != nil {
		Logger.Error("occur error on creating new http.Request object", zap.Error(err))
		return err
	}

	resp, err := ha.Client.Do(req)
	if err != nil {
		Logger.Error("occur error on sending http request", zap.Error(err))
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("occur error on receiving http response", zap.Error(err))
		return err
	}
	resp.Body.Close()

	for _, check := range ha.CheckChain {
		if err = check(resp, body); err != nil {
			return err
		}
	}
	return nil
}

// CheckHTTPStatusCode 检查状态码是否>=400, 如果是则视为请求失败
func CheckHTTPStatusCode(resp *http.Response, body []byte) error {
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return nil
}
