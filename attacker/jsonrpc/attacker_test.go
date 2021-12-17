package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	rpcResult struct {
		Sn     string `json:"sn"`
		Amount int    `json:"amount"`
	}
)

func TestAttacker(t *testing.T) {
	attacker := NewJSONRPCAttacker(
		"https://api.foobar.com",
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
				if ret.Amount <= 0 {
					return errors.New("bad amount")
				}
			}
			return nil
		}),
	)
	err := attacker.Fire(context.Background())
	assert.Nil(t, err)
}
