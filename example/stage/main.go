package main

import (
	"net/http"
	"time"

	"github.com/qastub/ultron"
)

const (
	//api = "http://10.0.0.30/benchmark"
	api = "http://www.baidu.com"
)

func main() {
	attacker := ultron.NewHTTPAttacker("benchmark", func() (*http.Request, error) { return http.NewRequest(http.MethodGet, api, nil) })
	task := ultron.NewTask()
	stage1 := ultron.NewStageConfig(5 * time.Minute, 1000, 225)
	stage2 := ultron.NewStageConfig(1 * time.Hour, 12300, 500)
	task.Add(attacker, 1)

	//ultron.LocalRunner.Config.Concurrence = 100
	//ultron.LocalRunner.Config.HatchRate = 10
	//ultron.LocalRunner.Config.MinWait = ultron.ZeroDuration
	//ultron.LocalRunner.Config.MaxWait = ultron.ZeroDuration

	ultron.LocalRunner.Config.AppendStage(stage1).AppendStage(stage2)

	ultron.LocalRunner.WithTask(task)
	ultron.LocalRunner.Start()
}
