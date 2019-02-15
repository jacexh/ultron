package ultron

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

type (
	mockFasthttpPrepare struct {
		mock.Mock
	}

	mockFasthttpClient struct {
		mock.Mock
	}
)

func (ep *mockFasthttpPrepare) handle(req *fasthttp.Request) error {
	args := ep.Called(req)
	return args.Error(0)
}

func (fc *mockFasthttpClient) Do(req *fasthttp.Request, res *fasthttp.Response) error {
	args := fc.Called(req, res)
	return args.Error(0)
}

func TestNewFastHTTPAttacker(t *testing.T) {
	attcker := NewFastHTTPAttacker("hello", nil)
	assert.Equal(t, attcker.Name(), "hello")
	assert.Nil(t, attcker.CheckChain)
	assert.Nil(t, attcker.Prepare)
}

func TestFastHTTPAttacker_Name(t *testing.T) {
	name := "hello"
	attacker := NewFastHTTPAttacker(name, nil)
	assert.Equal(t, attacker.Name(), name)
}

func TestFastHTTPAttacker_Fire_NoPrepare(t *testing.T) {
	assert.Panics(t,
		func() { NewFastHTTPAttacker("hello", nil).Fire() },
	)
}

func TestFastHTTPAttacker_Fire_PrepareError(t *testing.T) {
	p := new(mockFasthttpPrepare)
	p.On("handle", fasthttp.AcquireRequest()).Return(errors.New("bad prepare func"))
	attacker := NewFastHTTPAttacker("hello", p.handle)

	assert.EqualError(t, attacker.Fire(), "bad prepare func")
}

//func TestFastHTTPAttacker_Fire_RequestError(t *testing.T) {
//	p := new(mockFasthttpPrepare)
//	p.On("handle").Return(nil)
//	attacker := NewFastHTTPAttacker("hello", p.handle)
//
//	c := new(mockFasthttpClient)
//	attacker.Client = c
//}
