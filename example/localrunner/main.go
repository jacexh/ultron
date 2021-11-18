package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/wosai/ultron/v2"
)

type (
	fakeAttacker struct {
		name string
	}
)

func (f *fakeAttacker) Name() string {
	return f.name
}

func (f *fakeAttacker) Fire(_ context.Context) error {
	i := rand.Float32()
	if i < 0.05 {
		return errors.New("unknown error")
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func main() {
	runner := ultron.NewLocalRunner()
	task := ultron.NewTask()
	task.Add(&fakeAttacker{name: "foobar"}, 1)
	task.Add(&fakeAttacker{name: "ultron-test"}, 1)
	runner.Assign(task)

	if err := runner.Launch(); err != nil {
		panic(err)
	}

	// open http://localhost:2017
	block := make(chan struct{}, 1)
	<-block
}
