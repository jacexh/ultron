package eventbus

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

func PrintReportToConsole(output io.Writer) statistics.ReportHandleFunc {
	return func(ctx context.Context, report statistics.SummaryReport) {
		table := tablewriter.NewWriter(output)
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

		for _, rpt := range report.Reports {
			cells := []string{
				rpt.Name,
				rpt.Min.String(),
				rpt.Distributions["0.50"].String(),
				rpt.Distributions["0.60"].String(),
				rpt.Distributions["0.70"].String(),
				rpt.Distributions["0.80"].String(),
				rpt.Distributions["0.90"].String(),
				rpt.Distributions["0.95"].String(),
				rpt.Distributions["0.97"].String(),
				rpt.Distributions["0.98"].String(),
				rpt.Distributions["0.99"].String(),
				rpt.Max.String(),
				rpt.Average.String(),
				strconv.FormatUint(rpt.Requests, 10),
				strconv.FormatUint(rpt.Failures, 10),
				strconv.FormatFloat(rpt.TPS, 'f', 2, 64),
			}
			table.Append(cells)
		}
		table.Render()
		fmt.Fprintln(output, "")
	}
}
