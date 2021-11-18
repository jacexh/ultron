package influxdbv1

import (
	"context"
	"math/rand"
	"sync"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	InfluxDBV1Handler struct {
		client influxdb.Client
		conf   *InfluxDBV1HandlerConfig
		buffer *batchPointsBuffer
	}

	batchPointsBuffer struct {
		handler  *InfluxDBV1Handler
		bp       influxdb.BatchPoints
		interval time.Duration
		mu       sync.Mutex
		conf     influxdb.BatchPointsConfig
	}

	// InfluxDBV1HandlerConfig InfluxDBHelper配置
	InfluxDBV1HandlerConfig struct {
		URL                    string
		UDP                    bool
		User                   string
		Password               string
		Database               string
		MeasurementSucc        string
		MeasurementFail        string
		MeasurementAggregation string
	}
)

const (
	// DefaultInfluxDBURL influxdb address
	DefaultInfluxDBURL = "127.0.0.1:8089"
	// DefaultInfluxDBName influxdb database name
	DefaultInfluxDBName = "ultron"
	// DefaultMeasurementSucc the measurement to store successful request
	DefaultMeasurementSucc = "success"
	// DefaultMeasurementFail the measurement to store failed request
	DefaultMeasurementFail = "failures"
	// DefaultMeasurementAggregation the measurement to store report
	DefaultMeasurementAggregation = "report"
)

func (b *batchPointsBuffer) addPoint(p *influxdb.Point) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.bp == nil {
		bp, err := influxdb.NewBatchPoints(b.conf)
		if err != nil {
			ultron.Logger.Error("failed to call NewBatchPoints: " + err.Error())
			return
		}
		b.bp = bp
	}
	b.bp.AddPoint(p)
}

func (b *batchPointsBuffer) flushing() {
	for {
		time.Sleep(b.interval)

		b.mu.Lock()
		if b.bp == nil {
			b.mu.Unlock()
			continue
		}

		go func(c influxdb.Client, bp influxdb.BatchPoints) {
			err := c.Write(bp)
			if err != nil {
				ultron.Logger.Error("failed to write bach points", zap.Error(err))
			}
		}(b.handler.client, b.bp)
		b.bp = nil
		b.mu.Unlock()
	}
}

func newInfluxDBHTTPClient(url, user, password string) (influxdb.Client, error) {
	return influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     url,
		Username: user,
		Password: password,
	})
}

func newInfluxDBUDPClient(url string) (influxdb.Client, error) {
	return influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: url,
	})
}

// NewInfluxDBV1HandlerConfig 实例化InfluxDBHelpConfig默认配置
func NewInfluxDBV1HandlerConfig() *InfluxDBV1HandlerConfig {
	return &InfluxDBV1HandlerConfig{
		URL:                    DefaultInfluxDBURL,
		UDP:                    false,
		User:                   "",
		Password:               "",
		Database:               DefaultInfluxDBName,
		MeasurementSucc:        DefaultMeasurementSucc,
		MeasurementFail:        DefaultMeasurementFail,
		MeasurementAggregation: DefaultMeasurementAggregation,
	}
}

// NewInfluxDBV1Handler 实例化NewInfluxDBHelper对象
func NewInfluxDBV1Handler(conf *InfluxDBV1HandlerConfig) (*InfluxDBV1Handler, error) {
	var err error
	var client influxdb.Client
	if conf.UDP {
		client, err = newInfluxDBUDPClient(conf.URL)
	} else {
		client, err = newInfluxDBHTTPClient(conf.URL, conf.User, conf.Password)
	}

	if err != nil {
		ultron.Logger.Error("failed to init influxdb client: " + err.Error())
		return nil, err
	}

	buf := &batchPointsBuffer{
		handler:  nil,
		interval: 200 * time.Millisecond,
		conf:     influxdb.BatchPointsConfig{Precision: "ms", Database: conf.Database},
	}
	handler := &InfluxDBV1Handler{client: client, conf: conf, buffer: buf}
	buf.handler = handler
	go buf.flushing()

	return handler, nil
}

// HandleResult samplingRate表示采样率
func (hdl *InfluxDBV1Handler) HandleResult(samplingRate float64) ultron.ResultHandleFunc {
	return func(c context.Context, ar statistics.AttackResult) {
		if rand.Float64() > samplingRate {
			return
		}

		if ar.IsFailure() { // todo: 以后实现
			return
		}

		point, err := influxdb.NewPoint(
			hdl.conf.MeasurementSucc,
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
				hdl.conf.MeasurementAggregation,
				tags,
				map[string]interface{}{
					"tps":        report.TPS,
					"success":    int64(report.Requests),
					"failures":   int64(report.Failures),
					"fail_ratio": report.FailRatio,
					"min":        report.Min.Milliseconds(),
					"max":        report.Max.Milliseconds(),
					"avg":        report.Average.Milliseconds(),
					"TP50":       report.Distributions["0.50"].Milliseconds(),
					"TP60":       report.Distributions["0.60"].Milliseconds(),
					"TP70":       report.Distributions["0.70"].Milliseconds(),
					"TP80":       report.Distributions["0.80"].Milliseconds(),
					"TP90":       report.Distributions["0.90"].Milliseconds(),
					"TP95":       report.Distributions["0.95"].Milliseconds(),
					"TP96":       report.Distributions["0.96"].Milliseconds(),
					"TP97":       report.Distributions["0.97"].Milliseconds(),
					"TP98":       report.Distributions["0.98"].Milliseconds(),
					"TP99":       report.Distributions["0.99"].Milliseconds(),
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
