package ultron

import (
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

type (
	mockPrepareFunc struct {
		mock.Mock
	}

	mockHTTPChecker struct {
		mock.Mock
	}

	mockFastClient struct {
		mock.Mock
	}
)

func (ep *mockPrepareFunc) fastHTTPPrepare(req *fasthttp.Request) error {
	args := ep.Called(req)
	return args.Error(0)
}

func (ep *mockPrepareFunc) httpPrepare() (req *http.Request, err error) {
	args := ep.Called()
	return args.Get(0).(*http.Request), args.Error(1)
}

func (hc *mockHTTPChecker) httpChecker(res *http.Response, d []byte) error {
	args := hc.Called(res, d)
	return args.Error(0)
}

func (hc *mockHTTPChecker) fastHTTPChecker(res *fasthttp.Response) error {
	args := hc.Called(res)
	return args.Error(0)
}

func (mc *mockFastClient) Do(req *fasthttp.Request, res *fasthttp.Response) error {
	args := mc.Called(req, res)
	return args.Error(0)
}

func (mc *mockFastClient) DoDeadline(req *fasthttp.Request, res *fasthttp.Response, t time.Time) error {
	args := mc.Called(req, res, t)
	return args.Error(0)
}

func (mc *mockFastClient) DoTimeout(req *fasthttp.Request, res *fasthttp.Response, d time.Duration) error {
	args := mc.Called(req, res, d)
	return args.Error(0)
}

func TestNewFastHTTPAttacker(t *testing.T) {
	attacker := NewFastHTTPAttacker("hello", nil, nil)
	assert.NotNil(t, attacker)
	assert.Equal(t, attacker.Name(), "hello")
	assert.Nil(t, attacker.CheckChain)
	assert.Nil(t, attacker.Prepare)
}

func TestFastHTTPAttacker_Name(t *testing.T) {
	name := "hello"
	attacker := NewFastHTTPAttacker(name, nil, nil)
	assert.Equal(t, attacker.Name(), name)
}

func TestFastHTTPAttacker_Fire_NoPrepare(t *testing.T) {
	assert.Panics(t,
		func() { NewFastHTTPAttacker("hello", nil, nil).Fire() },
	)
}

func TestFastHTTPAttacker_Fire_PrepareError(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("fastHTTPPrepare", fasthttp.AcquireRequest()).Return(errors.New("bad prepare func"))
	attacker := NewFastHTTPAttacker("hello", nil, p.fastHTTPPrepare)

	assert.EqualError(t, attacker.Fire(), "bad prepare func")
}

func TestFastHTTPAttacker_Fire(t *testing.T) {
	client := new(mockFastClient)
	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(nil)
	attacker := NewFastHTTPAttacker(
		"foobar",
		client,
		func(request *fasthttp.Request) error { return nil },
	)
	assert.Nil(t, attacker.Fire())

	client = new(mockFastClient)
	attacker = NewFastHTTPAttacker(
		"foobar",
		client,
		func(request *fasthttp.Request) error { return nil },
	)
	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(errors.New("fire failure"))
	assert.EqualError(t, attacker.Fire(), "fire failure")
}

func TestFastHTTPAttacker_CheckError(t *testing.T) {
	client := new(mockFastClient)
	checker := new(mockHTTPChecker)

	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(nil)
	checker.On("fastHTTPChecker", &fasthttp.Response{}).Return(errors.New("check failure"))
	attacker := NewFastHTTPAttacker(
		"foobar",
		client,
		func(request *fasthttp.Request) error { return nil },
		checker.fastHTTPChecker,
	)
	assert.EqualError(t, attacker.Fire(), "check failure")
}

func TestNewHTTPAttacker(t *testing.T) {
	name := "hello world"
	attacker := NewHTTPAttacker(name, nil)
	assert.NotNil(t, attacker)
	assert.Equal(t, name, attacker.Name())
	assert.Nil(t, attacker.Prepare)
	assert.Nil(t, attacker.CheckChain)
}

func TestHTTPAttacker_Name(t *testing.T) {
	name := strconv.FormatInt(time.Now().UnixNano(), 10)
	attacker := NewHTTPAttacker(name, nil)
	assert.NotNil(t, attacker)
	assert.Equal(t, name, attacker.Name())
	assert.Equal(t, attacker.Client, DefaultHTTPClient)
}

func TestHTTPAttacker_Fire_NoPrepare(t *testing.T) {
	assert.Panics(t, func() { NewHTTPAttacker("hello", nil).Fire() })
}

func TestHTTPAttacker_Fire_PrepareError(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("httpPrepare").Return(new(http.Request), errors.New("bad prepare func"))
	attacker := NewHTTPAttacker("hello", p.httpPrepare)
	assert.EqualError(t, attacker.Fire(), "bad prepare func")
}
