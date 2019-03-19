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
)

func (ep *mockPrepareFunc) fastHTTPPrepare(req *fasthttp.Request) error {
	args := ep.Called(req)
	return args.Error(0)
}

func (ep *mockPrepareFunc) httpPrepare() (*http.Request, error) {
	args := ep.Called()
	return args.Get(0).(*http.Request), args.Error(1)
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
	attacker := NewFastHTTPAttacker(name, nil)
	assert.Equal(t, attacker.Name(), name)
}

func TestFastHTTPAttacker_Fire_NoPrepare(t *testing.T) {
	assert.Panics(t,
		func() { NewFastHTTPAttacker("hello", nil).Fire() },
	)
}

func TestFastHTTPAttacker_Fire_PrepareError(t *testing.T) {
	p := new(mockPrepareFunc)
	p.On("fastHTTPPrepare", fasthttp.AcquireRequest()).Return(errors.New("bad prepare func"))
	attacker := NewFastHTTPAttacker("hello", p.fastHTTPPrepare)

	assert.EqualError(t, attacker.Fire(), "bad prepare func")
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
}

func TestHTTPAttacker_Fire_NoPrepare(t *testing.T) {
	assert.Panics(t, func() { NewHTTPAttacker("hello", nil).Fire() })
}

func TestHTTPAttacker_Fire_PrepareError(t *testing.T) {
	p := new(mockPrepareFunc)
	req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
	p.On("httpPrepare").Return(req, errors.New("bad prepare func"))
	attacker := NewHTTPAttacker("hello", p.httpPrepare)
	assert.EqualError(t, attacker.Fire(), "bad prepare func")
}
