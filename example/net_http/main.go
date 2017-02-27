package main

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jacexh/ultron"
)

type (
	BaiduAttacker struct {
		name   string
		client *http.Client
	}
)

func newBaiduAttacher(n string) *BaiduAttacker {
	return &BaiduAttacker{
		name: n,
		client: &http.Client{
			Timeout: time.Second * 5,
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 1000,
			},
		},
	}
}

func (b *BaiduAttacker) Name() string {
	return b.name
}

func (b *BaiduAttacker) Fire() (time.Duration, error) {
	request, err := http.NewRequest(http.MethodGet, "http://proxy.sz.wosai-inc.com/", nil)
	if err != nil {
		return ultron.ZeroDuration, err
	}
	response, err := b.client.Do(request)
	if err != nil {
		return ultron.ZeroDuration, err
	}
	defer response.Body.Close()
	_, err = ioutil.ReadAll(response.Body)
	return ultron.ZeroDuration, err
}

func main() {
	taskSet := ultron.NewTaskSet()
	taskSet.Concurrency = 100
	taskSet.MinWait = ultron.ZeroDuration
	taskSet.MaxWait = ultron.ZeroDuration
	taskSet.Add(newBaiduAttacher("index"), 1)
	ultron.CoreRunner.WithTaskSet(taskSet).Run()
}
