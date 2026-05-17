package maintenance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsInSingleWindow(t *testing.T) {
	now := time.Date(2026, 5, 17, 14, 30, 0, 0, time.UTC)
	start := time.Date(2026, 5, 17, 14, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)

	m := &Maintenance{Strategy: "single", StartDate: &start, EndDate: &end}
	assert.True(t, isActiveNow(m, now))

	before := time.Date(2026, 5, 17, 13, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, before))

	after := time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, after))
}

func TestIsInRecurringWeekday(t *testing.T) {
	// May 15 2026 = Friday (5)
	now := time.Date(2026, 5, 15, 10, 30, 0, 0, time.UTC)
	m := &Maintenance{
		Strategy:  "recurring-weekday",
		Weekdays:  "[1,3,5]",
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	assert.True(t, isActiveNow(m, now))

	// May 16 = Saturday (6), not in weekdays list
	saturday := time.Date(2026, 5, 16, 10, 30, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, saturday))

	// Right day, wrong time
	lateOnFriday := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, lateOnFriday))
}

func TestIsInRecurringWeekdayWithDateRange(t *testing.T) {
	effectiveStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	effectiveEnd := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	m := &Maintenance{
		Strategy:  "recurring-weekday",
		Weekdays:  "[5]",
		StartTime: "10:00",
		EndTime:   "11:00",
		StartDate: &effectiveStart,
		EndDate:   &effectiveEnd,
	}

	inRange := time.Date(2026, 5, 15, 10, 30, 0, 0, time.UTC) // Friday (5) in range
	assert.True(t, isActiveNow(m, inRange))

	outOfRange := time.Date(2026, 5, 22, 10, 30, 0, 0, time.UTC) // Friday (5) out of range
	assert.False(t, isActiveNow(m, outOfRange))
}

func TestIsInRecurringDayOfMonth(t *testing.T) {
	m := &Maintenance{
		Strategy:    "recurring-day-of-month",
		DaysOfMonth: "[1,15]",
		StartTime:   "02:00",
		EndTime:     "04:00",
	}

	inWindow := time.Date(2026, 5, 15, 3, 0, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, inWindow))

	wrongDay := time.Date(2026, 5, 14, 3, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, wrongDay))

	wrongTime := time.Date(2026, 5, 15, 5, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, wrongTime))
}

func TestIsInRecurringInterval(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	m := &Maintenance{
		Strategy:    "recurring-interval",
		IntervalDay: 7,
		StartDate:   &start,
		StartTime:   "02:00",
		EndTime:     "04:00",
	}

	// Day 7 after start (May 8)
	inWindow := time.Date(2026, 5, 8, 3, 0, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, inWindow))

	// Day 14 after start (May 15)
	alsoInWindow := time.Date(2026, 5, 15, 3, 0, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, alsoInWindow))

	// Day 6 after start (wrong interval day)
	wrongDay := time.Date(2026, 5, 7, 3, 0, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, wrongDay))
}

func TestIsInCronWindow(t *testing.T) {
	// Every Sunday at 02:00 for 60 minutes
	m := &Maintenance{
		Strategy:        "cron",
		CronExpression:  "0 2 * * 0",
		DurationMinutes: 60,
	}

	// Sunday May 17 2026 is actually a Sunday
	inWindow := time.Date(2026, 5, 17, 2, 30, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, inWindow))

	afterWindow := time.Date(2026, 5, 17, 3, 30, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, afterWindow))

	wrongDay := time.Date(2026, 5, 16, 2, 30, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, wrongDay))
}

func TestIsInTimeWindowMidnightWrap(t *testing.T) {
	// Window from 22:00 to 06:00 (wraps midnight)
	now2300 := time.Date(2026, 5, 17, 23, 0, 0, 0, time.UTC)
	assert.True(t, isInTimeWindow("22:00", "06:00", now2300))

	now0300 := time.Date(2026, 5, 17, 3, 0, 0, 0, time.UTC)
	assert.True(t, isInTimeWindow("22:00", "06:00", now0300))

	now1200 := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	assert.False(t, isInTimeWindow("22:00", "06:00", now1200))
}

func TestManualStrategyAlwaysActive(t *testing.T) {
	m := &Maintenance{Strategy: "manual", Active: true}
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, now))
}

func TestTimezoneHandling(t *testing.T) {
	m := &Maintenance{
		Strategy:       "recurring-weekday",
		Weekdays:       "[5]",
		StartTime:      "10:00",
		EndTime:        "11:00",
		TimezoneOption: "America/New_York",
	}

	// May 15 2026 = Friday. 14:30 UTC = 10:30 EDT
	utcTime := time.Date(2026, 5, 15, 14, 30, 0, 0, time.UTC)
	assert.True(t, isActiveNow(m, utcTime))

	// 10:30 UTC = 06:30 EDT (too early)
	earlyUTC := time.Date(2026, 5, 15, 10, 30, 0, 0, time.UTC)
	assert.False(t, isActiveNow(m, earlyUTC))
}
