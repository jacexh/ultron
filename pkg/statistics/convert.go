package statistics

import (
	"errors"
	"time"

	"github.com/wosai/ultron/pkg/statistics/proto"
)

func ConvertAsDTO(as *AttackStatistician) (*proto.AttackStatisticsDTO, error) {
	if as == nil {
		return nil, errors.New("failed to convert as dto: <nil>")
	}
	as.mu.Lock()
	defer as.mu.Unlock()

	dto := &proto.AttackStatisticsDTO{
		Name:                as.name,
		Requests:            as.requests,
		Failures:            as.failures,
		TotalResponseTime:   int64(as.totalResponseTime),
		MinResponseTime:     int64(as.minResponseTime),
		MaxResponseTime:     int64(as.maxResponseTime),
		RecentSuccessBucket: make(map[int64]int64),
		RecentFailureBucket: make(map[int64]int64),
		ResponseBucket:      make(map[int64]uint64),
		FailureBucket:       make(map[string]uint64),
		FirstAttack:         as.firstAttack.UnixNano(),
		LastAttack:          as.lastAttack.UnixNano(),
		Interval:            int64(as.interval),
	}

	for k, v := range as.recentSuccessBucket.container {
		dto.RecentSuccessBucket[k] = v
	}
	for k, v := range as.recentFailureBucket.container {
		dto.RecentFailureBucket[k] = v
	}
	for k, v := range as.responseBucket {
		dto.ResponseBucket[int64(k)] = v
	}
	for k, v := range as.failureBucket {
		dto.FailureBucket[k] = v
	}
	return dto, nil
}

func NewAttackStatisticianFromDTO(dto *proto.AttackStatisticsDTO) (*AttackStatistician, error) {
	if dto == nil {
		return nil, errors.New("failed to new AttackStatistician: <nil>")
	}
	as := NewAttackStatistician(dto.Name)
	as.requests = dto.Requests
	as.failures = dto.Failures
	as.totalResponseTime = time.Duration(dto.TotalResponseTime)
	as.minResponseTime = time.Duration(dto.MinResponseTime)
	as.maxResponseTime = time.Duration(dto.MaxResponseTime)
	for k, v := range dto.RecentSuccessBucket {
		as.recentSuccessBucket.accumulate(k, v)
	}
	for k, v := range dto.RecentFailureBucket {
		as.recentFailureBucket.accumulate(k, v)
	}
	for k, v := range dto.ResponseBucket {
		as.responseBucket[time.Duration(k)] = v
	}
	for k, v := range dto.FailureBucket {
		as.failureBucket[k] = v
	}
	as.firstAttack = time.Unix(0, dto.FirstAttack)
	as.lastAttack = time.Unix(0, dto.LastAttack)

	return as, nil
}
