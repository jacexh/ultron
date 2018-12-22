package main

import (
	"net/http"

	"github.com/qastub/ultron"
)

const (
	api = "http://10.0.0.30/benchmark"
)

func main() {
	attacker := ultron.NewHTTPAttacker("benchmark", func() (*http.Request, error) { return http.NewRequest(http.MethodGet, api, nil) })
	task := ultron.NewTask()
	task.Add(attacker, 1)

	ultron.LocalRunner.Config.Concurrence = 100
	ultron.LocalRunner.Config.HatchRate = 10
	ultron.LocalRunner.Config.MinWait = ultron.ZeroDuration
	ultron.LocalRunner.Config.MaxWait = ultron.ZeroDuration

	ultron.LocalRunner.WithTask(task)
	ultron.LocalRunner.Start()
}
