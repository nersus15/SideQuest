// Package schedule decides when a scheduled quest is due. It is pure logic
// so it can be tested without clocks or timers.
package schedule

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// Days a schedule runs on.
const (
	Daily    = "daily"
	Weekdays = "weekdays"
	Weekends = "weekends"
)

// MaxTimes is how many times a day can be scheduled.
const MaxTimes = 4

// Grace is how long after its time a slot can still fire, e.g. when the
// screen was busy at that moment. After that it is skipped, never piled up.
const Grace = 15 * time.Minute

// Schedule is when Boo offers a quest on its own.
type Schedule struct {
	Times  []string `json:"times"`  // "HH:MM", local time
	Days   string   `json:"days"`   // Daily, Weekdays or Weekends
	Notify bool     `json:"notify"` // also send a system notification
}

// ErrInvalid is returned for a malformed schedule.
var ErrInvalid = errors.New("invalid schedule")

// Normalize validates s and returns it with times sorted.
func Normalize(s Schedule) (Schedule, error) {
	if s.Days == "" {
		s.Days = Daily
	}
	if s.Days != Daily && s.Days != Weekdays && s.Days != Weekends {
		return Schedule{}, ErrInvalid
	}
	if len(s.Times) > MaxTimes {
		return Schedule{}, ErrInvalid
	}
	times := make([]string, 0, len(s.Times))
	for _, t := range s.Times {
		h, m, ok := parse(t)
		if !ok {
			return Schedule{}, ErrInvalid
		}
		norm := fmt.Sprintf("%02d:%02d", h, m)
		if slices.Contains(times, norm) {
			return Schedule{}, ErrInvalid
		}
		times = append(times, norm)
	}
	slices.Sort(times)
	s.Times = times
	return s, nil
}

func parse(t string) (h, m int, ok bool) {
	if len(t) != 5 || t[2] != ':' {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(t, "%02d:%02d", &h, &m); err != nil {
		return 0, 0, false
	}
	return h, m, h >= 0 && h < 24 && m >= 0 && m < 60
}

func (s Schedule) runsOn(d time.Weekday) bool {
	weekend := d == time.Saturday || d == time.Sunday
	switch s.Days {
	case Weekdays:
		return !weekend
	case Weekends:
		return weekend
	}
	return true
}

// Due returns the key of the slot that is due at now (its time has passed
// by less than Grace), e.g. "2026-10-21 10:00". Keys let callers fire each
// slot exactly once.
func (s Schedule) Due(now time.Time) (string, bool) {
	if !s.runsOn(now.Weekday()) {
		return "", false
	}
	for _, t := range s.Times {
		h, m, ok := parse(t)
		if !ok {
			continue
		}
		at := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
		if d := now.Sub(at); d >= 0 && d < Grace {
			return at.Format("2006-01-02 15:04"), true
		}
	}
	return "", false
}
