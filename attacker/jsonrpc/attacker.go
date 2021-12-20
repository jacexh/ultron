package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2"
)

type (
	// JSONRPCAttacker jsonrpc 2.0 over http
	JSONRPCAttacker struct {
		id            int32
		client        *http.Client
		endpoint      string
		method        string
		prepareFunc   PrepareFunc
		checkFuncs    []CheckFunc
		unmarshalFunc UnmarshalFunc
	}

	// Request is http schema for jsonrpc 2.0
	Request struct {
		ID      *int32      `json:"id,omitempty"`
		Version string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}

	// Response is reply schema for jsonrpc 2.0
	Response struct {
		Version string          `json:"jsonrpc"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *ErrorDetails   `json:"error,omitempty"`
	}

	// ErrorDetails is error schema for jsonrpc 2.0
	ErrorDetails struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}

	// PrepareFunc return the params of jsonrpc method
	PrepareFunc func(context.Context) interface{}

	// UnmarshalFunc how to convert response.result to golang struct
	UnmarshalFunc func([]byte) (interface{}, error)

	// CheckFunc check if response.result meet expectation
	CheckFunc func(context.Context, interface{}) error

	// Option ...
	Option func(*JSONRPCAttacker)

	bytesPool struct {
		pool sync.Pool
	}
)

var (
	bp = newBytesPool()

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

	_ ultron.Attacker = (*JSONRPCAttacker)(nil)
)

func newBytesPool() *bytesPool {
	return &bytesPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func (bp *bytesPool) get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *bytesPool) put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}

func (res *Response) Err() error {
	if res.Error == nil {
		return nil
	}
	return res.Error
}

// Error  impl error interface
func (e *ErrorDetails) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// NewJSONRPCAttacker init a jsonrpc attacker
func NewJSONRPCAttacker(endpoint, method string, opts ...Option) *JSONRPCAttacker {
	attacker := &JSONRPCAttacker{
		client:     defaultHTTPClient,
		method:     method,
		endpoint:   endpoint,
		checkFuncs: make([]CheckFunc, 0),
	}

	attacker.Apply(opts...)
	return attacker
}

func (j *JSONRPCAttacker) Name() string {
	return j.method
}

func (j *JSONRPCAttacker) Fire(ctx context.Context) error {
	if j.prepareFunc == nil {
		panic("call Apply(WithPrepareFunc()) first")
	}

	defer ultron.ClearContext(ctx)

	v := atomic.AddInt32(&j.id, 1)
	payload := &Request{
		ID:      &v,
		Version: "2.0",
		Method:  j.method,
		Params:  j.prepareFunc(ctx),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	buf := bp.get()
	defer bp.put(buf)
	if _, err = buf.Write(data); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, j.endpoint, buf)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "github.com/wosai/ultron")
	res, err := j.client.Do(req)
	if err != nil {
		return err
	}

	resObj := new(Response)
	err = json.NewDecoder(res.Body).Decode(resObj)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if err = resObj.Err(); err != nil {
		return err
	}

	if j.unmarshalFunc != nil {
		result, err := j.unmarshalFunc(resObj.Result)
		if err != nil {
			return err
		}
		for _, checker := range j.checkFuncs {
			if err := checker(ctx, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func (j *JSONRPCAttacker) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(j)
	}
}

func WithPrepareFunc(fn PrepareFunc) Option {
	return func(j *JSONRPCAttacker) {
		if fn != nil {
			j.prepareFunc = fn
		}
	}
}

func WithCheckFuncs(checks ...CheckFunc) Option {
	return func(j *JSONRPCAttacker) {
		for _, check := range checks {
			if check == nil {
				panic("invalid check function")
			}
			j.checkFuncs = append(j.checkFuncs, check)
		}
	}
}

func WithClient(client *http.Client) Option {
	return func(j *JSONRPCAttacker) {
		if client != nil {
			j.client = client
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(j *JSONRPCAttacker) {
		if j.client != nil {
			j.client.Timeout = d
		}
	}
}

func WithUnmarshalFunc(fn UnmarshalFunc) Option {
	return func(j *JSONRPCAttacker) {
		if fn != nil {
			j.unmarshalFunc = fn
		}
	}
}
