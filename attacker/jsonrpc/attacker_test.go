package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	rpcResult struct {
		Sn     string `json:"sn"`
		Amount int    `json:"amount"`
	}
)

func testServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		res := &Response{Version: "2.0", Result: []byte(`{"sn": "1000", "amount": 2}`)}
		data, _ := json.Marshal(res)
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	return ts
}

func TestAttacker(t *testing.T) {
	ts := testServer()
	defer ts.Close()
	attacker := NewJSONRPCAttacker(
		ts.URL,
		"query",
		WithPrepareFunc(func() interface{} {
			return []string{"w4414"}
		}),
		WithUnmarshalFunc(func(b []byte) (interface{}, error) {
			ret := new(rpcResult)
			err := json.Unmarshal(b, ret)
			return ret, err
		}),
		WithCheckFuncs(func(i interface{}) error {
			if ret, ok := i.(*rpcResult); ok {
				if ret.Amount != 2 {
					return errors.New("bad amount")
				}
			}
			return nil
		}),
	)
	err := attacker.Fire(context.Background())
	assert.Nil(t, err)
}
