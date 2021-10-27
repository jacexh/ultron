package main

import (
	"github.com/wosai/ultron/v2"
	_ "github.com/wosai/ultron/v2/internal/master"
)

func main() {
	runner := ultron.BuildMasterRunner()
	_ = runner.Launch(ultron.RunnerConfig{})
}
