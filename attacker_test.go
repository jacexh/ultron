package ultron

import (
	"errors"
	"io"
	"io/ioutil"
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

	mockHTTPClient struct {
		mock.Mock
		client *http.Client
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

func (mh *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := mh.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNewFastHTTPAttacker(t *testing.T) {
	attacker := NewFastHTTPAttacker("hello", nil)
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
	attacker := NewFastHTTPAttacker("hello", p.fastHTTPPrepare)

	assert.EqualError(t, attacker.Fire(), "bad prepare func")
}

func TestFastHTTPAttacker_FireError(t *testing.T) {
	client := new(mockFastClient)
	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(nil)
	attacker := NewFastHTTPAttacker(
		"foobar",
		func(request *fasthttp.Request) error { return nil },
	)
	attacker.Client = client
	assert.Nil(t, attacker.Fire())
}

func TestFastHTTPAttacker_FireOK(t *testing.T) {
	client := new(mockFastClient)
	attacker := NewFastHTTPAttacker(
		"foobar",
		func(request *fasthttp.Request) error { return nil },
	)
	attacker.Client = client
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
		func(request *fasthttp.Request) error { return nil },
		checker.fastHTTPChecker,
	)
	attacker.Client = client
	assert.EqualError(t, attacker.Fire(), "check failure")
}

func TestFastHTTPAttacker_CheckFuncNil(t *testing.T) {
	client := new(mockFastClient)

	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(nil)
	attacker := NewFastHTTPAttacker(
		"foobar",
		func(request *fasthttp.Request) error { return nil },
		nil, nil,
	)
	attacker.Client = client
	assert.Nil(t, attacker.Fire())
}

func TestFastHTTPAttacker_NoCheckFunc(t *testing.T) {
	client := new(mockFastClient)

	client.On("Do", &fasthttp.Request{}, &fasthttp.Response{}).Return(nil)
	attacker := NewFastHTTPAttacker(
		"foobar",
		func(request *fasthttp.Request) error { return nil },
	)
	attacker.Client = client
	assert.Nil(t, attacker.Fire())
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

func TestHTTPAttacker_FireOK(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("httpPrepare").Return(&http.Request{}, nil)

	res, _ := http.Get("https://www.baidu.com")

	checker := new(mockHTTPChecker)
	checker.On("check")

	client := new(mockHTTPClient)
	client.On("Do", &http.Request{}).Return(res, nil)
	attacker := NewHTTPAttacker("foobar", p.httpPrepare)
	attacker.Client = client
	assert.Nil(t, attacker.Fire())
}

func TestHTTPAttacker_FireError(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("httpPrepare").Return(&http.Request{}, nil)

	client := new(mockHTTPClient)
	client.On("Do", &http.Request{}).Return(&http.Response{}, errors.New("fetch failure"))
	attacker := NewHTTPAttacker("foobar", p.httpPrepare)
	attacker.Client = client
	assert.EqualError(t, attacker.Fire(), "fetch failure")
}

func TestHTTPAttacker_CheckOK(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("httpPrepare").Return(&http.Request{}, nil)

	res, _ := http.Get("https://www.baidu.com")
	io.Copy(ioutil.Discard, res.Body) // 读取body，使得Fire()内的body值为空...

	checker := new(mockHTTPChecker)
	checker.On("httpChecker", res, []byte{}).Return(nil)

	client := new(mockHTTPClient)
	client.On("Do", &http.Request{}).Return(res, nil)
	attacker := NewHTTPAttacker(
		"foobar",
		p.httpPrepare,
		checker.httpChecker)
	attacker.Client = client
	assert.Nil(t, attacker.Fire())
}

func TestHTTPAttacker_CheckError(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("httpPrepare").Return(&http.Request{}, nil)

	res, _ := http.Get("https://www.baidu.com")
	io.Copy(ioutil.Discard, res.Body) // 读取body，使得Fire()内的body值为空...

	checker := new(mockHTTPChecker)
	checker.On("httpChecker", res, []byte{}).Return(errors.New("check failure"))

	client := new(mockHTTPClient)
	client.On("Do", &http.Request{}).Return(res, nil)
	attacker := NewHTTPAttacker(
		"foobar",
		p.httpPrepare,
		checker.httpChecker)
	attacker.Client = client
	assert.EqualError(t, attacker.Fire(), "check failure")
}
