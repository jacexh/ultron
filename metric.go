package ultron

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	metric struct {
		report statistics.SummaryReport
		mu     sync.Mutex
	}
)

var (
	_ prometheus.Collector = (*metric)(nil)
)

var (
	metricTags          = []string{KeyAttacker, KeyPlan}
	descTotalRequests   = prometheus.NewDesc("ultron_attacker_requests_total", "total requests number of this attacker", metricTags, nil)
	descTotalFailures   = prometheus.NewDesc("ultron_attacker_failures_total", "total failures number of this attacker", metricTags, nil)
	descMinResponseTime = prometheus.NewDesc("ultron_attacker_response_time_min", "the min response time for this attacker", metricTags, nil)
	descMaxResponseTime = prometheus.NewDesc("ultron_attacker_response_time_max", "the max response time for this attacker", metricTags, nil)
	descAvgResponseTime = prometheus.NewDesc("ultron_attacker_response_time_avg", "the avg response time for this attacker", metricTags, nil)
	descResponseTime    = prometheus.NewDesc("ultron_attacker_response_time", "the response time for this attacker", metricTags, nil)
	descFailureRatio    = prometheus.NewDesc("ultron_attacker_failure_ratio", "the failure ratio of this attacker", metricTags, nil)
	descCurrentTPS      = prometheus.NewDesc("ultron_attacker_tps_current", "current TPS of this attacker", metricTags, nil)
	descTotalTPS        = prometheus.NewDesc("ultron_attacker_tps_total", "total TPS of this attacker", metricTags, nil)
	descConcurrentUsers = prometheus.NewDesc("ultron_current_users", "the number of current users", []string{KeyPlan}, nil)
	descSlaves          = prometheus.NewDesc("ultron_slaves", "the number of salves", []string{KeyPlan}, nil)
)

func newMetric() *metric {
	return &metric{}
}

func (m *metric) Describe(ch chan<- *prometheus.Desc) {
	ch <- descTotalRequests
	ch <- descTotalFailures
	ch <- descMinResponseTime
	ch <- descMaxResponseTime
	ch <- descAvgResponseTime
	ch <- descResponseTime
	ch <- descFailureRatio
	ch <- descCurrentTPS
	ch <- descTotalTPS
	ch <- descConcurrentUsers
	ch <- descSlaves
}

func (m *metric) Collect(ch chan<- prometheus.Metric) {
	m.mu.Lock()
	report := m.report
	plan := m.report.Extras[KeyPlan]
	m.mu.Unlock()

	if report.FirstAttack.IsZero() { // 空的报告
		return
	}

	for _, report := range report.Reports {
		ch <- prometheus.MustNewConstMetric(descTotalRequests, prometheus.GaugeValue, float64(report.Requests), report.Name, plan)
		ch <- prometheus.MustNewConstMetric(descTotalFailures, prometheus.GaugeValue, float64(report.Failures), report.Name, plan)
		ch <- prometheus.MustNewConstMetric(descMinResponseTime, prometheus.GaugeValue, float64(report.Min.Milliseconds()), report.Name, plan)
		ch <- prometheus.MustNewConstMetric(descMaxResponseTime, prometheus.GaugeValue, float64(report.Max.Milliseconds()), report.Name, plan)
		ch <- prometheus.MustNewConstMetric(descAvgResponseTime, prometheus.GaugeValue, float64(report.Average.Milliseconds()), report.Name, plan)
		ch <- prometheus.MustNewConstSummary(descResponseTime, report.Requests, float64(report.Average.Milliseconds())*float64(report.Requests), map[float64]float64{
			0.00: float64(report.Min.Milliseconds()),
			0.50: float64(report.Median.Milliseconds()),
			0.60: float64(report.Distributions["0.60"].Milliseconds()),
			0.70: float64(report.Distributions["0.70"].Milliseconds()),
			0.80: float64(report.Distributions["0.80"].Milliseconds()),
			0.90: float64(report.Distributions["0.90"].Milliseconds()),
			0.95: float64(report.Distributions["0.95"].Milliseconds()),
			0.97: float64(report.Distributions["0.97"].Milliseconds()),
			0.98: float64(report.Distributions["0.98"].Milliseconds()),
			0.99: float64(report.Distributions["0.99"].Milliseconds()),
			1.00: float64(report.Distributions["1.00"].Milliseconds()),
		}, report.Name, plan)
		ch <- prometheus.MustNewConstMetric(descFailureRatio, prometheus.GaugeValue, report.FailRatio, report.Name, plan)
		if report.FullHistory {
			ch <- prometheus.MustNewConstMetric(descTotalTPS, prometheus.GaugeValue, report.TPS, report.Name, plan)
		} else {
			ch <- prometheus.MustNewConstMetric(descCurrentTPS, prometheus.GaugeValue, report.TPS, report.Name, plan)
		}
	}
}

func (m *metric) handleReport() ReportHandleFunc {
	return func(c context.Context, sr statistics.SummaryReport) {
		m.mu.Lock()
		defer m.mu.Unlock()

		m.report = sr
	}
}
