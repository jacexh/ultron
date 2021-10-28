package main

import (
	"github.com/wosai/ultron/v2"
)

func main() {
	runner := ultron.BuildMasterRunner()
	_ = runner.Launch(ultron.RunnerConfig{})
}
