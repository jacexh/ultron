package stage

import "time"

type (
	StageOption struct {
		Durtion     time.Duration // 阶段持续时间，不严格控制
		Requests    uint64        // 阶段请求总数，不严格控制
		Concurrence uint32        // 阶段目标并发数
		HatchRate   int32         // 进入该阶段时，每秒增压、降压数目。
		MinWait     time.Duration // 最小等待时间
		MaxWait     time.Duration // 最大等待时间
	}
)
