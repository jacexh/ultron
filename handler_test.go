package ultron

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

func TestTerminalTabl(t *testing.T) {
	sg := statistics.NewStatisticianGroup()
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 1 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 2 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 3 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 4 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 5 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 6 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 7 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 8 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 9 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 9 * time.Millisecond, Error: errors.New("unknown error")})

	report := sg.Report(true)
	printReportToConsole(os.Stdout)(context.TODO(), report)
}

func TestPrintJsonReport(t *testing.T) {
	sg := statistics.NewStatisticianGroup()
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 1 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 2 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 3 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 4 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 5 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 6 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 7 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 8 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 9 * time.Millisecond})
	sg.Record(statistics.AttackResult{Name: "unittest", Duration: 9 * time.Millisecond, Error: errors.New("unknown error")})

	report := sg.Report(true)
	printJsonReport(os.Stdout)(context.TODO(), report)
}
