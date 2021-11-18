package main

import (
	"net/http"

	"github.com/wosai/ultron/v2"
)

func main() {
	runner := ultron.NewLocalRunner()
	task := ultron.NewTask()
	bing := ultron.NewHTTPAttacker("bing")
	bing.Apply(
		ultron.WithPrepareFunc(func() (*http.Request, error) {
			return http.NewRequest(http.MethodGet, "https://bing.com", nil)
		}),
		ultron.WithCheckFuncs(ultron.CheckHTTPStatusCode),
	)
	baidu := ultron.NewHTTPAttacker("baidu")
	baidu.Apply(
		ultron.WithPrepareFunc(func() (*http.Request, error) {
			return http.NewRequest(http.MethodGet, "https://www.baidu.com", nil)
		}),
		ultron.WithCheckFuncs(ultron.CheckHTTPStatusCode),
	)

	task.Add(bing, 1)
	task.Add(baidu, 1)
	runner.Assign(task)

	if err := runner.Launch(); err != nil {
		panic(err)
	}

	// open http://localhost:2017
	block := make(chan struct{}, 1)
	<-block
}
