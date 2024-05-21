package main

import (
	"context"
	"net/http"

	"github.com/wosai/ultron/v2"
	"google.golang.org/grpc"
)

func main() {
	slave := ultron.NewSlaveRunner()
	task := ultron.NewTask()
	attacker := ultron.NewHTTPAttacker("google")
	attacker.Apply(ultron.WithPrepareFunc(func(context.Context) (*http.Request, error) {
		return http.NewRequest(http.MethodGet, "https://www.google.com", nil)
	}))
	task.Add(attacker, 1)
	slave.Assign(task)

	if err := slave.Connect("127.0.0.1:2021", grpc.WithInsecure()); err != nil {
		panic(err)
	}

	select {}
}
