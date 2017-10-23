package main

import (
	"net/http"

	"github.com/jacexh/ultron"
	"google.golang.org/grpc"
)

func main() {
	task := ultron.NewTask()
	baidu := ultron.NewHTTPAttacker("nginx", func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodGet, "http://www.baidu.com/", nil)
		if err != nil {
			return nil, err
		}
		return req, nil
	})
	task.Add(baidu, 1)

	ultron.SlaveRunner.Connect("127.0.0.1:9500", grpc.WithInsecure())
	ultron.SlaveRunner.WithTask(task)
	ultron.SlaveRunner.Start()
}
