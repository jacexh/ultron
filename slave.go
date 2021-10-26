package ultron

type (
	// SlaveAgent 定义master侧的slave对象
	SlaveAgent interface {
		ID() string
		Extras() map[string]string
	}
)
