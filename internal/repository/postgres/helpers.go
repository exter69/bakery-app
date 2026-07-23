package postgres

import (
	"fmt"
	"math"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// dayIntToWeek maps Postgres day_of_week INT (0=Sunday) to domain.DayOfWeek.
var dayIntToWeek = map[int]domain.DayOfWeek{
	0: domain.Sunday,
	1: domain.Monday,
	2: domain.Tuesday,
	3: domain.Wednesday,
	4: domain.Thursday,
	5: domain.Friday,
	6: domain.Saturday,
}

// dayWeekToInt is the reverse mapping.
var dayWeekToInt = map[domain.DayOfWeek]int{
	domain.Sunday:    0,
	domain.Monday:    1,
	domain.Tuesday:   2,
	domain.Wednesday: 3,
	domain.Thursday:  4,
	domain.Friday:    5,
	domain.Saturday:  6,
}

// centsToDecimal converts an int64 cents amount to a float64 for DB storage.
func centsToDecimal(cents int64) float64 {
	return float64(cents) / 100.0
}

// decimalToCents converts a float64 decimal from the DB to int64 cents.
func decimalToCents(d float64) int64 {
	return int64(math.Round(d * 100))
}

// timeOfDayToTime converts a domain.TimeOfDay to a time.Time (date part is zero).
func timeOfDayToTime(tod domain.TimeOfDay) time.Time {
	return time.Date(0, 1, 1, tod.Hour, tod.Minute, 0, 0, time.UTC)
}

// timeToTimeOfDay converts a time.Time to domain.TimeOfDay (ignoring date).
func timeToTimeOfDay(t time.Time) domain.TimeOfDay {
	return domain.TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}
}

// paginationToOffset converts page/pageSize to SQL LIMIT/OFFSET values.
func paginationToOffset(params domain.PaginationParams) (limit, offset int) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	return pageSize, (page - 1) * pageSize
}

// orderSortColumn returns a safe SQL column name for order sorting.
func orderSortColumn(sortBy string) string {
	switch sortBy {
	case "scheduledTime":
		return "scheduled_start_time"
	case "createdAt":
		return "created_at"
	default:
		return "created_at"
	}
}

// sortDirection returns "ASC" or "DESC" from a filter value.
func sortDirection(dir string) string {
	if dir == "asc" {
		return "ASC"
	}
	return "DESC"
}

// buildWhereStatus appends a status filter clause. Returns updated query, args, argIndex.
func buildWhereStatus[T ~string](baseQuery string, args []any, argIdx int, status *T) (string, []any, int) {
	if status == nil {
		return baseQuery, args, argIdx
	}
	baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
	args = append(args, string(*status))
	argIdx++
	return baseQuery, args, argIdx
}

// nilIfEmpty returns nil if s is empty, otherwise a pointer to s.
// Used for nullable VARCHAR columns that should be NULL when unset.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
