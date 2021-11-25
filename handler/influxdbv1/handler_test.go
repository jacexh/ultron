//go:build !race

package influxdbv1

import (
	"sync"
	"testing"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
)

func TestBatchPoints(t *testing.T) {
	bp := newBatchPointBuffer(NewInfluxDBV1Handler())
	go bp.flushing()

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
	<-time.After(100 * time.Millisecond) // 不稳定
	assert.NotNil(t, bp.bp)
	assert.EqualValues(t, len(bp.bp.Points()), 10)
}

func TestBatchPoints_Flush(t *testing.T) {
	bp := newBatchPointBuffer(NewInfluxDBV1Handler())
	go bp.flushing()

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
	<-time.After(510 * time.Millisecond) // 不稳定
	assert.Nil(t, bp.bp)                 // 已经重置
}

func TestWithDatabase(t *testing.T) {
	handler := NewInfluxDBV1Handler()
	handler.Apply(WithDatabase("ultron-test"))
	assert.EqualValues(t, handler.database, "ultron-test")
	assert.EqualValues(t, handler.buffer.conf.Database, "ultron-test")
}
