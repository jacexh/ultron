package statistics

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/assert"
)

func TestFindResponseBucket(t *testing.T) {
	t1 := 61 * time.Millisecond
	assert.EqualValues(t, t1, findResponseBucket(t1))

	t2 := 121 * time.Millisecond
	assert.EqualValues(t, 120*time.Millisecond, findResponseBucket(t2))

	t3 := 1111 * time.Millisecond
	assert.EqualValues(t, 1100*time.Millisecond, findResponseBucket(t3))

	t4 := 3 * time.Second * 333 * time.Microsecond
	assert.NotEqualValues(t, 3300*time.Microsecond, findResponseBucket(t4))
}

func TestAttackStatistician_Record(t *testing.T) {
	as := NewAttackStatistician("foobar")
	as.Record(AttackResult{
		Name:     "foobar",
		Duration: 200 * time.Millisecond,
	})
	as.Record(AttackResult{
		Name:     "foobar",
		Duration: 300 * time.Millisecond,
		Error:    errors.New("unknown"),
	})
	report := as.Report(false)
	assert.EqualValues(t, report.Requests, 1)
	assert.EqualValues(t, report.Failures, 1)
	assert.EqualValues(t, report.FailRation, .5)
}

func BenchmarkAttackResultAggregator_RecordSuccess(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			agg.recordSuccess(AttackResult{
				Name:     "benchmark",
				Duration: 111 * time.Millisecond,
			})
		}
	})
}

func BenchmarkStatistician_SyncRecord(b *testing.B) {
	s := NewStatisticianGroup()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Record(AttackResult{
				Name:     "benchmar",
				Duration: 111 * time.Millisecond,
			})
		}
	})
}

func BenchmarkAttackResultAggregator_Percent(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.9)
	}
}

func BenchmarkAttackResultAggregator_Percent2(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.5, .6, .7, .8, .9, .95, .99)
	}
}

func TestAttackResultAggregator_Percent(t *testing.T) {
	agg := NewAttackStatistician("benchmark")
	for i := 1; i <= 11; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(i) * time.Millisecond,
		})
	}

	assert.EqualValues(t, agg.percentile(.0)[0], 1*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.1)[0], 1*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.5)[0], 6*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.9)[0], 10*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.97)[0], 11*time.Millisecond)
}

func TestAttackResultAggregator_merge(t *testing.T) {
	a1 := NewAttackStatistician("test")
	for i := 0; i < 10; i++ {
		a1.Record(AttackResult{Name: "test", Duration: time.Duration(i) * time.Millisecond})
	}
	a2 := NewAttackStatistician("test")
	for j := 10; j < 20; j++ {
		a2.Record(AttackResult{Name: "test", Duration: time.Duration(j) * time.Millisecond})
	}

	a1.merge(a2)
	report := a1.Report(true)

	assert.EqualValues(t, report.Average, 9500*time.Microsecond)
	assert.EqualValues(t, report.Median, 9*time.Millisecond)
	assert.EqualValues(t, report.Requests, 20)
}

// todo: remove this case
func TestAttackResultAggregator_Report(t *testing.T) {
	s := NewStatisticianGroup()
	for i := 0; i < 400*400; i++ {
		s.Record(AttackResult{Name: "/api/foobar", Duration: time.Duration(rand.Int63n(1200)+1) * time.Millisecond})
	}
	report := s.Report(true)
	data := [][]string{
		{
			report.Reports["/api/foobar"].Name,
			report.Reports["/api/foobar"].Min.String(),
			report.Reports["/api/foobar"].Distributions["0.50"].String(),
			report.Reports["/api/foobar"].Distributions["0.60"].String(),
			report.Reports["/api/foobar"].Distributions["0.70"].String(),
			report.Reports["/api/foobar"].Distributions["0.80"].String(),
			report.Reports["/api/foobar"].Distributions["0.90"].String(),
			report.Reports["/api/foobar"].Distributions["0.95"].String(),
			report.Reports["/api/foobar"].Distributions["0.97"].String(),
			report.Reports["/api/foobar"].Distributions["0.98"].String(),
			report.Reports["/api/foobar"].Distributions["0.99"].String(),
			report.Reports["/api/foobar"].Max.String(),
			report.Reports["/api/foobar"].Average.String(),
			strconv.FormatUint(report.Reports["/api/foobar"].Requests, 10),
			strconv.FormatUint(report.Reports["/api/foobar"].Failures, 10),
			strconv.FormatFloat(report.Reports["/api/foobar"].TPS, 'f', 2, 64)},
	}
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Attacker", "Min", "P50", "P60", "P70", "P80", "P90", "P95", "P97", "P98", "P99", "Max", "Avg", "Requests", "Failures", "TPS"}

	table.SetHeader(header)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
	)

	footer := make([]string, 16)
	if report.FullHistory {
		footer[11] = "Full History"
	}
	footer[12] = "Total"
	footer[13] = strconv.FormatUint(report.TotalRequests, 10)
	footer[14] = strconv.FormatUint(report.TotalFailures, 10)
	footer[15] = strconv.FormatFloat(report.TotalTPS, 'f', 2, 64)

	table.SetFooter(footer)
	table.SetFooterColor(
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlueColor},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
	)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.AppendBulk(data)
	table.Render()

	fmt.Fprint(os.Stdout, "\n")
}
