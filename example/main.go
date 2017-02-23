package main

import (
	"github.com/jacexh/ultron"
	"go.uber.org/zap"
)

func main() {
	ultron.Logger.Info("danger", zap.String("foo", "bar"))
}
