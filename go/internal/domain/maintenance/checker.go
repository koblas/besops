package maintenance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Checker evaluates whether a monitor is currently within an active maintenance window.
type Checker struct {
	repo Repository
}

func NewChecker(repo Repository) *Checker {
	return &Checker{repo: repo}
}

// IsMonitorInMaintenance returns true if the given monitor has at least one
// active maintenance window that covers the current time.
func (c *Checker) IsMonitorInMaintenance(ctx context.Context, monitorID string) (bool, error) {
	maintenanceIDs, err := c.repo.GetMonitorMaintenanceIDs(ctx, monitorID)
	if err != nil {
		return false, fmt.Errorf("getting maintenance IDs for monitor %s: %w", monitorID, err)
	}

	if len(maintenanceIDs) == 0 {
		return false, nil
	}

	now := time.Now()

	for _, mID := range maintenanceIDs {
		m, err := c.repo.FindByID(ctx, mID)
		if err != nil {
			continue
		}

		if !m.Active {
			continue
		}

		if isActiveNow(m, now) {
			return true, nil
		}
	}

	return false, nil
}

func isActiveNow(m *Maintenance, now time.Time) bool {
	loc := loadTimezone(m.TimezoneOption)
	now = now.In(loc)

	switch m.Strategy {
	case "manual":
		return true
	case "single":
		return isInSingleWindow(m, now)
	case "recurring-weekday":
		return isInRecurringWeekday(m, now)
	case "recurring-day-of-month":
		return isInRecurringDayOfMonth(m, now)
	case "recurring-interval":
		return isInRecurringInterval(m, now)
	case "cron":
		return isInCronWindow(m, now)
	default:
		return false
	}
}

func loadTimezone(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

func isInSingleWindow(m *Maintenance, now time.Time) bool {
	if m.StartDate == nil || m.EndDate == nil {
		return false
	}
	return !now.Before(*m.StartDate) && !now.After(*m.EndDate)
}

func isInRecurringWeekday(m *Maintenance, now time.Time) bool {
	if m.StartDate != nil && now.Before(*m.StartDate) {
		return false
	}
	if m.EndDate != nil && now.After(*m.EndDate) {
		return false
	}

	weekdays := parseInt32List(m.Weekdays)
	todayWeekday := int32(now.Weekday()) //nolint:gosec // weekday is 0-6

	found := false
	for _, wd := range weekdays {
		if wd == todayWeekday {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	return isInTimeWindow(m.StartTime, m.EndTime, now)
}

func isInRecurringDayOfMonth(m *Maintenance, now time.Time) bool {
	days := parseInt32List(m.DaysOfMonth)
	todayDay := int32(now.Day()) //nolint:gosec // day is 1-31

	found := false
	for _, d := range days {
		if d == todayDay {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	return isInTimeWindow(m.StartTime, m.EndTime, now)
}

func isInRecurringInterval(m *Maintenance, now time.Time) bool {
	if m.StartDate == nil || m.IntervalDay <= 0 {
		return false
	}

	start := m.StartDate.In(now.Location())
	if now.Before(start) {
		return false
	}

	daysSinceStart := int(now.Sub(start).Hours() / 24)
	if daysSinceStart%m.IntervalDay != 0 {
		return false
	}

	return isInTimeWindow(m.StartTime, m.EndTime, now)
}

func isInCronWindow(m *Maintenance, now time.Time) bool {
	if m.CronExpression == "" || m.DurationMinutes <= 0 {
		return false
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(m.CronExpression)
	if err != nil {
		return false
	}

	duration := time.Duration(m.DurationMinutes) * time.Minute

	// Check if we're within a window that started recently.
	// Walk backwards from now to find the most recent fire time.
	candidate := now.Add(-duration)
	next := sched.Next(candidate)

	// If the most recent fire time + duration covers now, we're in the window.
	return !next.After(now) && next.Add(duration).After(now)
}

// isInTimeWindow checks if now's clock time is between startTime and endTime (HH:mm format).
func isInTimeWindow(startTime, endTime string, now time.Time) bool {
	if startTime == "" || endTime == "" {
		return false
	}

	startH, startM := parseHHMM(startTime)
	endH, endM := parseHHMM(endTime)

	nowMinutes := now.Hour()*60 + now.Minute()
	startMinutes := startH*60 + startM
	endMinutes := endH*60 + endM

	if endMinutes > startMinutes {
		return nowMinutes >= startMinutes && nowMinutes < endMinutes
	}
	// Wraps midnight (e.g. 22:00 - 06:00)
	return nowMinutes >= startMinutes || nowMinutes < endMinutes
}

func parseHHMM(s string) (int, int) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m
}
