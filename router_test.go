package ultron

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRouter(t *testing.T) {
	runner := newMasterRunner()
	handler := buildHTTPRouter(runner)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/metrics.json")
	assert.Nil(t, err)
	data, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	d := make([]interface{}, 0)
	err = json.Unmarshal(data, &d)
	assert.Nil(t, err)

	reader := bytes.NewBuffer([]byte("foobar"))
	res, err = http.Post(ts.URL+"/api/v1/plan", "", reader)
	assert.Nil(t, err)
	ret := new(restResponse)
	err = json.NewDecoder(res.Body).Decode(ret)
	assert.Nil(t, err)
	assert.True(t, ret.ErrorMessage != "")
}
