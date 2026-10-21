package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestNormalize(t *testing.T) {
	s, err := Normalize(Schedule{Times: []string{"15:00", "09:30"}})
	if err != nil || s.Days != Daily || s.Times[0] != "09:30" || s.Times[1] != "15:00" {
		t.Fatalf("got %+v %v", s, err)
	}
	bad := []Schedule{
		{Times: []string{"25:00"}},
		{Times: []string{"9:30"}},
		{Times: []string{"10:00", "10:00"}},
		{Times: []string{"01:00", "02:00", "03:00", "04:00", "05:00"}},
		{Days: "mondays"},
		{Times: []string{"aa:bb"}},
	}
	for i, b := range bad {
		if _, err := Normalize(b); !errors.Is(err, ErrInvalid) {
			t.Errorf("case %d should be invalid, got %v", i, err)
		}
	}
}

func TestDue(t *testing.T) {
	wed := time.Date(2026, 10, 21, 0, 0, 0, 0, time.Local) // a Wednesday
	s := Schedule{Times: []string{"10:00", "15:30"}, Days: Daily}
	at := func(d time.Time, hm string) time.Time {
		h, m, _ := parse(hm)
		return d.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
	}

	if _, ok := s.Due(at(wed, "09:59")); ok {
		t.Error("not due before its time")
	}
	if key, ok := s.Due(at(wed, "10:00")); !ok || key != "2026-10-21 10:00" {
		t.Errorf("due at its time, got %q %v", key, ok)
	}
	if key, ok := s.Due(at(wed, "10:14")); !ok || key != "2026-10-21 10:00" {
		t.Errorf("still due within grace, got %q %v", key, ok)
	}
	if _, ok := s.Due(at(wed, "10:15")); ok {
		t.Error("skipped after the grace period, never piled up")
	}

	sat := wed.AddDate(0, 0, 3)
	weekdays := Schedule{Times: []string{"10:00"}, Days: Weekdays}
	if _, ok := weekdays.Due(at(sat, "10:01")); ok {
		t.Error("weekday schedule should not run on Saturday")
	}
	weekends := Schedule{Times: []string{"10:00"}, Days: Weekends}
	if _, ok := weekends.Due(at(wed, "10:01")); ok {
		t.Error("weekend schedule should not run on Wednesday")
	}
	if _, ok := weekends.Due(at(sat, "10:01")); !ok {
		t.Error("weekend schedule should run on Saturday")
	}
}
