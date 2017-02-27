package ultron

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

var DefaultHTTPAttacker = &http.Client{
	Timeout: time.Second * 60,
	Transport: &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 1000,
	},
}

type (
	HTTPRequest struct {
		client  *http.Client
		name    string
		Prepare func() *http.Request
		Checker []func(*http.Response, []byte) error
	}
)

func NewHTTPRequest(n string) *HTTPRequest {
	return &HTTPRequest{
		client:  DefaultHTTPAttacker,
		name:    n,
		Checker: []func(*http.Response, []byte) error{CheckStatusCode},
	}
}

func (h *HTTPRequest) Name() string {
	return h.name
}

func (h *HTTPRequest) Fire() (time.Duration, error) {
	if h.Prepare == nil {
		panic("please impl Prepare()")
	}
	resp, err := h.client.Do(h.Prepare())
	if err != nil {
		return ZeroDuration, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ZeroDuration, err
	}
	for _, c := range h.Checker {
		err := c(resp, body)
		if err != nil {
			return ZeroDuration, err
		}
	}
	return ZeroDuration, err
}

func CheckStatusCode(r *http.Response, body []byte) error {
	if r.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code")
	}
	return nil
}
