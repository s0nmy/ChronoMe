package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEntryValidateTimeRange(t *testing.T) {
	start := time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name    string
		end     *time.Time
		invalid bool
	}{
		{name: "running"},
		{name: "later", end: timePointer(start.Add(time.Hour))},
		{name: "equal", end: timePointer(start), invalid: true},
		{name: "earlier", end: timePointer(start.Add(-time.Hour)), invalid: true},
		{name: "equal in another timezone", end: timePointer(start.In(time.FixedZone("JST", 9*60*60))), invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := Entry{Title: "Focus", StartedAt: start, EndedAt: tc.end, Ratio: 1}
			if tc.invalid {
				require.EqualError(t, entry.Validate(), "ended_at must be after started_at")
			} else {
				require.NoError(t, entry.Validate())
			}
		})
	}
}

func TestEntryUpdateDurationReplacesPreviousValue(t *testing.T) {
	start := time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		end  *time.Time
		now  time.Time
		want int64
	}{
		{name: "finished", end: timePointer(start.Add(30 * time.Minute)), now: start.Add(2 * time.Hour), want: 1800},
		{name: "running", now: start.Add(15 * time.Minute), want: 900},
		{name: "equal end", end: timePointer(start), now: start.Add(time.Hour)},
		{name: "earlier end", end: timePointer(start.Add(-time.Hour)), now: start.Add(time.Hour)},
		{name: "running at start", now: start},
		{name: "running before start", now: start.Add(-time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := Entry{StartedAt: start, EndedAt: tc.end, DurationSec: 3600}
			entry.UpdateDuration(tc.now)
			require.Equal(t, tc.want, entry.DurationSec)
		})
	}
}

func timePointer(value time.Time) *time.Time { return &value }
