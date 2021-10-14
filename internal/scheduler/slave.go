package scheduler

// Slave 节点接口
type Slave interface {
	ID() string
	ExecuteStage(planID string, stageSeq int, conf StageConfiguration) error
}
