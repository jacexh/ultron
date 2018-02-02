package helper

import (
	"bytes"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/jacexh/ultron"
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type (
	JSONRPCRequest struct {
		ID      int64       `json:"id"`
		Version string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
	}

	JSONRPCNotification struct {
		Version string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
	}

	JSONRPCResponse struct {
		ID     int                 `json:"id,omitempty"`
		Result jsoniter.RawMessage `json:"result,omitempty"`
		Error  *JSONRPCError       `json:"error,omitempty"`
	}

	JSONRPCError struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		Data    jsoniter.RawMessage `json:"data,omitempty"`
	}
)

var (
	AutoIncreaseJSONRPCID       = false
	jsonrpcIDCounter      int64 = 0
)

func NewJSONRPCRequest(method string, obj interface{}, args ...interface{}) *JSONRPCRequest {
	r := &JSONRPCRequest{Version: "2.0", Method: method}
	if AutoIncreaseJSONRPCID {
		r.ID = atomic.AddInt64(&jsonrpcIDCounter, 1)
	} else {
		r.ID = jsonrpcIDCounter
	}

	if obj != nil {
		r.Params = obj
	} else {
		r.Params = args
	}
	return r
}

func (req *JSONRPCRequest) ToHTTPRequest(url string) (*http.Request, error) {
	data, err := ultron.J.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	return r, nil

}

func (req *JSONRPCRequest) ToFastHTTPRequest(url string, r *fasthttp.Request) error {
	data, err := ultron.J.Marshal(req)
	if err != nil {
		return err
	}

	r.SetRequestURI(url)
	r.Header.SetMethod(http.MethodPost)
	r.Header.Set("Content-Type", "application/json")
	r.SetBody(data)
	return nil
}

func (err *JSONRPCError) Error() string {
	return fmt.Sprintf("<code>: %d  <message>: %s", err.Code, err.Message)
}

func (res *JSONRPCResponse) HasError() bool {
	if res.Error == nil {
		return false
	}
	return true
}

func (res *JSONRPCResponse) GetError() error {
	return res.Error
}
