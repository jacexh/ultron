package helper

import (
	"sync"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/qastub/ultron"
)

type (
	batchPointsBuffer struct {
		helper   *InfluxDBHelper
		bp       influx.BatchPoints
		interval time.Duration
		mu       sync.Mutex
		conf     influx.BatchPointsConfig
	}

	// InfluxDBHelper .
	InfluxDBHelper struct {
		client influx.Client
		conf   *InfluxDBHelperConfig
		buffer *batchPointsBuffer
	}

	// InfluxDBHelperConfig InfluxDBHelper配置
	InfluxDBHelperConfig struct {
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

func (b *batchPointsBuffer) addPoint(p *influx.Point) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.bp == nil {
		bp, err := influx.NewBatchPoints(b.conf)
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

		go func(c influx.Client, bp influx.BatchPoints) {
			err := c.Write(bp)
			if err != nil {
				ultron.Logger.Error("failed to write bach points: " + err.Error())
			}
		}(b.helper.client, b.bp)

		b.bp = nil
		b.mu.Unlock()
	}

}

func newInfluxDBHTTPClient(url, user, password string) (influx.Client, error) {
	return influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     url,
		Username: user,
		Password: password,
	})
}

func newInfluxDBUDPClient(url string) (influx.Client, error) {
	return influx.NewUDPClient(influx.UDPConfig{
		Addr: url,
	})
}

// NewInfluxDBHelperConfig 实例化InfluxDBHelpConfig默认配置
func NewInfluxDBHelperConfig() *InfluxDBHelperConfig {
	return &InfluxDBHelperConfig{
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

// NewInfluxDBHelper 实例化NewInfluxDBHelper对象
func NewInfluxDBHelper(conf *InfluxDBHelperConfig) (*InfluxDBHelper, error) {
	var err error
	var client influx.Client
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
		helper:   nil,
		interval: 200 * time.Millisecond,
		conf:     influx.BatchPointsConfig{Precision: "ms", Database: conf.Database},
	}
	helper := &InfluxDBHelper{client: client, conf: conf, buffer: buf}
	buf.helper = helper
	go buf.flushing()

	return helper, nil
}

// HandleResult 处理单次请求结果
func (i *InfluxDBHelper) HandleResult() ultron.ResultHandleFunc {
	return func(r *ultron.Result) {
		if r.Error != nil {
			// 暂时不处理失败的请求
			return
		}

		point, err := influx.NewPoint(
			i.conf.MeasurementSucc,
			map[string]string{"api": r.Name},
			map[string]interface{}{"response_time": int(r.Duration / 1e6)},
			time.Now(),
		)
		if err != nil {
			ultron.Logger.Error("failed to create new point: " + err.Error())
		} else {
			i.buffer.addPoint(point)
		}
	}
}

// HandleReport 处理聚合报告
func (i *InfluxDBHelper) HandleReport() ultron.ReportHandleFunc {
	return func(r ultron.Report) {
		for _, report := range r {
			// 不处理total
			if report.FullHistory {
				return
			}

			point, err := influx.NewPoint(
				i.conf.MeasurementAggregation,
				map[string]string{"api": report.Name},
				map[string]interface{}{
					"qps":        report.QPS,
					"success":    report.Requests,
					"failures":   report.Failures,
					"fail_ratio": report.FailRatio,
					"min":        report.Min,
					"max":        report.Max,
					"avg":        report.Average,
					"TP50":       report.Distributions["0.50"],
					"TP60":       report.Distributions["0.60"],
					"TP70":       report.Distributions["0.70"],
					"TP80":       report.Distributions["0.80"],
					"TP90":       report.Distributions["0.90"],
					"TP95":       report.Distributions["0.95"],
					"TP96":       report.Distributions["0.96"],
					"TP97":       report.Distributions["0.97"],
					"TP98":       report.Distributions["0.98"],
					"TP99":       report.Distributions["0.99"],
				},
				time.Now(),
			)
			if err != nil {
				ultron.Logger.Error("failed to create new point: " + err.Error())
			} else {
				i.buffer.addPoint(point)
			}
		}
	}
}
