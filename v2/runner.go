package ultron

type (
	// Runner 定义压测执行器
	Runner interface {
		Run(*AttackerSuite) error
		Shutdown() error
	}
)
