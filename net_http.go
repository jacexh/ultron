package ultron

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	// DefaultHTTPAttacker default net/http client
	DefaultHTTPAttacker = &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 1000,
		},
	}
)

type (
	// HTTPRequest net/http request
	HTTPRequest struct {
		client     *http.Client
		name       string
		parent     *TaskSet
		Prepare    func() *http.Request
		CheckChain []func(*http.Response, []byte) error
	}
)

// NewHTTPRequest create new HTTPRequest instance
func NewHTTPRequest(n string) *HTTPRequest {
	return &HTTPRequest{
		client:     DefaultHTTPAttacker,
		name:       n,
		CheckChain: []func(*http.Response, []byte) error{CheckStatusCode},
	}
}

// Name return the name of HTTPRequest
func (h *HTTPRequest) Name() string {
	return h.name
}

// Fire send to request and read response
func (h *HTTPRequest) Fire() error {
	if h.Prepare == nil {
		panic(errors.New("please impl Prepare()"))
	}
	resp, err := h.client.Do(h.Prepare())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	for _, check := range h.CheckChain {
		err := check(resp, body)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetTaskSet set taskset
func (h *HTTPRequest) SetTaskSet(t *TaskSet) {
	h.parent = t
}

// CheckStatusCode checker status code
func CheckStatusCode(r *http.Response, body []byte) error {
	if r.StatusCode >= http.StatusBadRequest {
		return errors.New("bad status code")
	}
	return nil
}
