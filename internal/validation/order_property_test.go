package validation

import (
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"pgregory.net/rapid"
)

// **Validates: Requirements 3.2, 3.3, 5.2, 6.7**

// genTimeOfDay generates a valid TimeOfDay with hour in [0,23] and minute in [0,59].
func genTimeOfDay() *rapid.Generator[domain.TimeOfDay] {
	return rapid.Custom(func(t *rapid.T) domain.TimeOfDay {
		return domain.TimeOfDay{
			Hour:   rapid.IntRange(0, 23).Draw(t, "hour"),
			Minute: rapid.IntRange(0, 59).Draw(t, "minute"),
		}
	})
}

// genTimeOfDayInRange generates a TimeOfDay that is >= lower and <= upper.
func genTimeOfDayInRange(lower, upper domain.TimeOfDay) *rapid.Generator[domain.TimeOfDay] {
	return rapid.Custom(func(t *rapid.T) domain.TimeOfDay {
		// Convert to total minutes for easy range generation
		lowerMins := lower.Hour*60 + lower.Minute
		upperMins := upper.Hour*60 + upper.Minute
		mins := rapid.IntRange(lowerMins, upperMins).Draw(t, "minutes")
		return domain.TimeOfDay{
			Hour:   mins / 60,
			Minute: mins % 60,
		}
	})
}

// genTimeOfDayBefore generates a TimeOfDay strictly before the given time.
func genTimeOfDayBefore(upper domain.TimeOfDay) *rapid.Generator[domain.TimeOfDay] {
	return rapid.Custom(func(t *rapid.T) domain.TimeOfDay {
		upperMins := upper.Hour*60 + upper.Minute
		if upperMins == 0 {
			// Can't go before 00:00, just return 00:00 (edge case handled by caller)
			return domain.TimeOfDay{Hour: 0, Minute: 0}
		}
		mins := rapid.IntRange(0, upperMins-1).Draw(t, "minutes")
		return domain.TimeOfDay{
			Hour:   mins / 60,
			Minute: mins % 60,
		}
	})
}

// genTimeOfDayAfter generates a TimeOfDay strictly after the given time.
func genTimeOfDayAfter(lower domain.TimeOfDay) *rapid.Generator[domain.TimeOfDay] {
	return rapid.Custom(func(t *rapid.T) domain.TimeOfDay {
		lowerMins := lower.Hour*60 + lower.Minute
		maxMins := 23*60 + 59
		if lowerMins >= maxMins {
			// Can't go after 23:59, return 23:59
			return domain.TimeOfDay{Hour: 23, Minute: 59}
		}
		mins := rapid.IntRange(lowerMins+1, maxMins).Draw(t, "minutes")
		return domain.TimeOfDay{
			Hour:   mins / 60,
			Minute: mins % 60,
		}
	})
}

// genDayOfWeek generates a random day of the week.
func genDayOfWeek() *rapid.Generator[domain.DayOfWeek] {
	return rapid.Custom(func(t *rapid.T) domain.DayOfWeek {
		days := domain.AllDaysOfWeek()
		idx := rapid.IntRange(0, len(days)-1).Draw(t, "dayIdx")
		return days[idx]
	})
}

// genOpenSchedule generates a DaySchedule where the bakery is open, with valid openTime < closeTime.
func genOpenSchedule(day domain.DayOfWeek) *rapid.Generator[domain.DaySchedule] {
	return rapid.Custom(func(t *rapid.T) domain.DaySchedule {
		// Generate openTime with room for closeTime after it
		// openTime can be at most 23:58 so closeTime can be at least 23:59
		openMins := rapid.IntRange(0, 23*60+58).Draw(t, "openMins")
		closeMins := rapid.IntRange(openMins+1, 23*60+59).Draw(t, "closeMins")
		return domain.DaySchedule{
			Day:       day,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: openMins / 60, Minute: openMins % 60},
			CloseTime: domain.TimeOfDay{Hour: closeMins / 60, Minute: closeMins % 60},
		}
	})
}

func TestProperty_ScheduleValidity_AcceptsWithinHours(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random day and an open schedule for that day
		day := genDayOfWeek().Draw(t, "day")
		sched := genOpenSchedule(day).Draw(t, "schedule")

		// Generate a time slot within operating hours
		startTime := genTimeOfDayInRange(sched.OpenTime, sched.CloseTime).Draw(t, "startTime")
		endTime := genTimeOfDayInRange(startTime, sched.CloseTime).Draw(t, "endTime")
		slot := domain.TimeSlot{StartTime: startTime, EndTime: endTime}

		schedule := []domain.DaySchedule{sched}
		result := ValidateSchedule(day, slot, schedule)

		if result.HasErrors() {
			t.Fatalf("expected no errors for valid schedule, got: %v", result.Errors)
		}
	})
}

func TestProperty_ScheduleValidity_RejectsClosedDay(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random day with bakery closed
		day := genDayOfWeek().Draw(t, "day")
		closedSchedule := domain.DaySchedule{
			Day:       day,
			IsOpen:    false,
			OpenTime:  domain.TimeOfDay{Hour: 0, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 0, Minute: 0},
		}

		// Generate any random time slot
		startTime := genTimeOfDay().Draw(t, "startTime")
		endTime := genTimeOfDay().Draw(t, "endTime")
		slot := domain.TimeSlot{StartTime: startTime, EndTime: endTime}

		schedule := []domain.DaySchedule{closedSchedule}
		result := ValidateSchedule(day, slot, schedule)

		if !result.HasErrors() {
			t.Fatal("expected error when bakery is closed, got none")
		}

		// Verify error mentions "closed"
		found := false
		for _, err := range result.Errors {
			if err.Field == "scheduledDay" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected scheduledDay error, got: %v", result.Errors)
		}
	})
}

func TestProperty_ScheduleValidity_RejectsStartBeforeOpening(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a day with open schedule where openTime > 00:00 so we can generate a startTime before it
		day := genDayOfWeek().Draw(t, "day")
		// Ensure openTime is at least 00:01 so there's room for a startTime before it
		openMins := rapid.IntRange(1, 23*60+58).Draw(t, "openMins")
		closeMins := rapid.IntRange(openMins+1, 23*60+59).Draw(t, "closeMins")
		sched := domain.DaySchedule{
			Day:       day,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: openMins / 60, Minute: openMins % 60},
			CloseTime: domain.TimeOfDay{Hour: closeMins / 60, Minute: closeMins % 60},
		}

		// Generate startTime strictly before openTime
		startTime := genTimeOfDayBefore(sched.OpenTime).Draw(t, "startTime")
		// endTime within operating hours to isolate the startTime error
		endTime := genTimeOfDayInRange(sched.OpenTime, sched.CloseTime).Draw(t, "endTime")
		slot := domain.TimeSlot{StartTime: startTime, EndTime: endTime}

		schedule := []domain.DaySchedule{sched}
		result := ValidateSchedule(day, slot, schedule)

		if !result.HasErrors() {
			t.Fatalf("expected error for startTime before opening, got none (start=%s, open=%s)", startTime, sched.OpenTime)
		}

		// Verify a startTime error is present
		found := false
		for _, err := range result.Errors {
			if err.Field == "scheduledTime.startTime" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected scheduledTime.startTime error, got: %v", result.Errors)
		}
	})
}

func TestProperty_ScheduleValidity_RejectsEndAfterClosing(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a day with open schedule where closeTime < 23:59 so there's room for an endTime after it
		day := genDayOfWeek().Draw(t, "day")
		openMins := rapid.IntRange(0, 23*60+57).Draw(t, "openMins")
		closeMins := rapid.IntRange(openMins+1, 23*60+58).Draw(t, "closeMins")
		sched := domain.DaySchedule{
			Day:       day,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: openMins / 60, Minute: openMins % 60},
			CloseTime: domain.TimeOfDay{Hour: closeMins / 60, Minute: closeMins % 60},
		}

		// Generate startTime within operating hours
		startTime := genTimeOfDayInRange(sched.OpenTime, sched.CloseTime).Draw(t, "startTime")
		// Generate endTime strictly after closeTime
		endTime := genTimeOfDayAfter(sched.CloseTime).Draw(t, "endTime")
		slot := domain.TimeSlot{StartTime: startTime, EndTime: endTime}

		schedule := []domain.DaySchedule{sched}
		result := ValidateSchedule(day, slot, schedule)

		if !result.HasErrors() {
			t.Fatalf("expected error for endTime after closing, got none (end=%s, close=%s)", endTime, sched.CloseTime)
		}

		// Verify an endTime error is present
		found := false
		for _, err := range result.Errors {
			if err.Field == "scheduledTime.endTime" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected scheduledTime.endTime error, got: %v", result.Errors)
		}
	})
}
