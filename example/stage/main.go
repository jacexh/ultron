package main

import (
	"net/http"
	"time"

	"github.com/qastub/ultron"
)

const (
	api = "http://10.0.0.30/benchmark"
)

func main() {
	attacker := ultron.NewHTTPAttacker("benchmark", func() (*http.Request, error) { return http.NewRequest(http.MethodGet, api, nil) })
	task := ultron.NewTask()
	stage1 := ultron.NewStageConfig(1 * time.Minute, 100, 225)
	stage2 := ultron.NewStageConfig(2*time.Minute, 300, 500)
	task.Add(attacker, 1)

	ultron.LocalRunner.Config.AppendStage(stage1).AppendStage(stage2)

	ultron.LocalRunner.WithTask(task)
	//fmt.Println("ultron.LocalRunner", ultron.LocalRunner)
	ultron.LocalRunner.Start()
}
