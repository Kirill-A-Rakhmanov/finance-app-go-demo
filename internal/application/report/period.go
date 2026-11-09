package report

import (
	"finance-app-go-demo/internal/application/query"
	"time"
)

const (
	MinBalanceReportDays = 2
	MaxDailyReportDays   = 30
	MaxWeeklyReportDays  = 365
	MaxReportPeriodYears = 100
)

func calendarDate(t *time.Time, l *time.Location) *time.Time {
	if t == nil {
		return nil
	}

	year, month, day := t.In(l).Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, l)

	return &date
}

func calendarDays(start, end time.Time) int {
	startYear, startMonth, startDay := start.Date()
	endYear, endMonth, endDay := end.Date()

	from := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)
	to := time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, time.UTC)

	return int(to.Sub(from).Hours()/24) + 1
}

func granularityForDays(days int) query.Granularity {
	switch {
	case days <= MaxDailyReportDays:
		return query.GranularityDay
	case days <= MaxWeeklyReportDays:
		return query.GranularityWeek
	default:
		return query.GranularityMonth
	}
}

func periodExceedsLimit(start, end time.Time) bool {
	return end.After(start.AddDate(MaxReportPeriodYears, 0, 0))
}
