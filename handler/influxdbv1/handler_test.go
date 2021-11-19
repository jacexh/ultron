package influxdbv1

import (
	"sync"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
)

func TestBatchPoints(t *testing.T) {
	bp := &batchPointsBuffer{
		handler:  nil,
		interval: 200 * time.Millisecond,
		conf: influxdb.BatchPointsConfig{
			Precision: "ms",
			Database:  "ultron"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			point, err := influxdb.NewPoint(
				"result",
				map[string]string{ultron.KeyAttacker: "unittest"},
				map[string]interface{}{"response_time": 10 * time.Millisecond},
				time.Now(),
			)
			assert.Nil(t, err)
			bp.addPoint(point)
		}()
	}

	wg.Wait()
	assert.NotNil(t, bp.bp)
	assert.EqualValues(t, len(bp.bp.Points()), 10)
}
