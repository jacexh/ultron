package influxdbv1

import (
	"context"
	"math/rand"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	InfluxDBV1Handler struct {
		client            influxdb.Client
		database          string
		measurementResult string
		measurementReport string
		buffer            *batchPointsBuffer
	}

	batchPointsBuffer struct {
		handler *InfluxDBV1Handler
		bp      influxdb.BatchPoints
		conf    influxdb.BatchPointsConfig
		ticker  *time.Ticker
		ch      chan *influxdb.Point
		size    uint32
		counter uint32
	}

	influxDBV1HandlerOption func(*InfluxDBV1Handler)
)

const (
	// DefaultInfluxDatabase influxdb database name
	DefaultInfluxDatabase = "ultron"
	// DefaultMeasurementResult the measurement to store successful request
	DefaultMeasurementResult = "result"
	// DefaultMeasurementReport the measurement to store report
	DefaultMeasurementReport = "report"
)

func newBatchPointBuffer(handler *InfluxDBV1Handler) *batchPointsBuffer {
	return &batchPointsBuffer{
		handler: handler,
		ch:      make(chan *influxdb.Point, 1024),
		size:    100,
	}
}

func (b *batchPointsBuffer) addPoint(p *influxdb.Point) {
	b.ch <- p
}

func (b *batchPointsBuffer) insertPointAndGetBatchPoint(p *influxdb.Point) (influxdb.BatchPoints, error) {
	if b.bp == nil {
		bp, err := influxdb.NewBatchPoints(b.conf)
		if err != nil {
			return nil, err
		}
		b.bp = bp
	}
	b.bp.AddPoint(p)
	return b.bp, nil
}

func (b *batchPointsBuffer) resetBufferAndWriteBatchPoint(bp influxdb.BatchPoints) {
	b.bp = nil

	go func(c influxdb.Client, bp influxdb.BatchPoints) {
		if c == nil {
			ultron.Logger.Warn("no influxdb client provided, you should call Apply(WithHTTPClient/WithUDPClient) first")
			return
		}
		err := c.Write(bp)
		if err != nil {
			ultron.Logger.Error("failed to write bach points", zap.Error(err))
		}
	}(b.handler.client, bp)
}

func (b *batchPointsBuffer) flushing() {
	b.ticker = time.NewTicker(500 * time.Millisecond)
	defer b.ticker.Stop()

flushing:
	for {
		select {
		case <-b.ticker.C:
			if b.bp != nil {
				b.counter = 0
				b.resetBufferAndWriteBatchPoint(b.bp)
			}

		case point := <-b.ch:
			bp, err := b.insertPointAndGetBatchPoint(point)
			if err != nil {
				ultron.Logger.Error("failed to get batchpoint", zap.Error(err))
				continue flushing
			}
			b.counter++
			if (b.counter % b.size) == 0 {
				b.resetBufferAndWriteBatchPoint(bp)
			}
		}
	}
}

// NewInfluxDBV1Handler 实例化NewInfluxDBHelper对象
func NewInfluxDBV1Handler() *InfluxDBV1Handler {
	handler := &InfluxDBV1Handler{
		database:          DefaultInfluxDatabase,
		measurementResult: DefaultMeasurementResult,
		measurementReport: DefaultMeasurementReport,
	}

	buf := newBatchPointBuffer(handler)
	handler.buffer = buf
	go buf.flushing()
	return handler
}

func (hdl *InfluxDBV1Handler) Apply(opts ...influxDBV1HandlerOption) {
	for _, opt := range opts {
		opt(hdl)
	}
}

// HandleResult samplingRate表示采样率
func (hdl *InfluxDBV1Handler) HandleResult(samplingRate float64) ultron.ResultHandleFunc {
	return func(c context.Context, ar statistics.AttackResult) {
		if rand.Float64() > samplingRate {
			return
		}

		if ar.IsFailure() { // TODO: 以后实现
			return
		}

		point, err := influxdb.NewPoint(
			hdl.measurementResult,
			map[string]string{ultron.KeyAttacker: ar.Name},
			map[string]interface{}{"response_time": ar.Duration.Milliseconds()},
			time.Now(),
		)
		if err != nil {
			ultron.Logger.Error("failed to create new point", zap.Error(err))
			return
		}
		hdl.buffer.addPoint(point)
	}
}

func (hdl *InfluxDBV1Handler) HandleReport() ultron.ReportHandleFunc {
	return func(c context.Context, sr statistics.SummaryReport) {
		if sr.FullHistory {
			return
		}

		now := time.Now()

		for key, report := range sr.Reports {
			tags := make(map[string]string)
			tags[ultron.KeyAttacker] = key
			for k, v := range sr.Extras {
				tags[k] = v
			}
			point, err := influxdb.NewPoint(
				hdl.measurementReport,
				tags,
				map[string]interface{}{
					"tps":           report.TPS,
					"successes":     int64(report.Requests),
					"failures":      int64(report.Failures),
					"failure_ratio": report.FailureRatio,
					"min":           report.Min.Milliseconds(),
					"max":           report.Max.Milliseconds(),
					"avg":           report.Average.Milliseconds(),
					"TP50":          report.Distributions["0.50"].Milliseconds(),
					"TP60":          report.Distributions["0.60"].Milliseconds(),
					"TP70":          report.Distributions["0.70"].Milliseconds(),
					"TP80":          report.Distributions["0.80"].Milliseconds(),
					"TP90":          report.Distributions["0.90"].Milliseconds(),
					"TP95":          report.Distributions["0.95"].Milliseconds(),
					"TP96":          report.Distributions["0.96"].Milliseconds(),
					"TP97":          report.Distributions["0.97"].Milliseconds(),
					"TP98":          report.Distributions["0.98"].Milliseconds(),
					"TP99":          report.Distributions["0.99"].Milliseconds(),
				},
				now,
			)
			if err != nil {
				ultron.Logger.Error("failed to create new point", zap.Error(err))
				return
			}
			hdl.buffer.addPoint(point)
		}
	}
}

func WithDatabase(db string) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		hdl.database = db
		hdl.buffer.conf.Database = db
	}
}

func WithMeasurementResult(n string) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		hdl.measurementResult = n
	}
}

func WithMeasurementReport(n string) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		hdl.measurementReport = n
	}
}

func WithHTTPClient(url, username, password string) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:     url,
			Username: username,
			Password: password,
		})
		if err != nil {
			panic(err)
		}
		hdl.client = client
	}
}

func WithUDPClient(url string, size int) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		client, err := influxdb.NewUDPClient(influxdb.UDPConfig{Addr: url, PayloadSize: size})
		if err != nil {
			panic(err)
		}
		hdl.client = client
	}
}

func WithBufferSize(size uint32) influxDBV1HandlerOption {
	return func(hdl *InfluxDBV1Handler) {
		if size == 0 {
			panic("bad buffer size")
		}
		hdl.buffer.size = size
	}
}
