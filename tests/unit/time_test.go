package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeWindowValidation(t *testing.T) {
	// Test 1d window validation: end = start + 24h, event_time = end
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	validEnd := start.Add(24 * time.Hour)
	validEvent := validEnd

	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		event     time.Time
		expectErr bool
	}{
		{
			name:      "valid daily window",
			start:     start,
			end:       validEnd,
			event:     validEvent,
			expectErr: false,
		},
		{
			name:      "invalid end time - too short",
			start:     start,
			end:       start.Add(23 * time.Hour),
			event:     validEvent,
			expectErr: true,
		},
		{
			name:      "invalid end time - too long",
			start:     start,
			end:       start.Add(25 * time.Hour),
			event:     validEvent,
			expectErr: true,
		},
		{
			name:      "invalid event time - before end",
			start:     start,
			end:       validEnd,
			event:     start.Add(12 * time.Hour),
			expectErr: true,
		},
		{
			name:      "invalid event time - after end",
			start:     start,
			end:       validEnd,
			event:     validEnd.Add(time.Hour),
			expectErr: true,
		},
		{
			name:      "DST boundary - spring forward",
			start:     time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), // DST starts
			end:       time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			event:     time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			expectErr: false,
		},
		{
			name:      "DST boundary - fall back",
			start:     time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC), // DST ends
			end:       time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			event:     time.Date(2024, 11, 3, 0, 0, 0, 0, time.UTC).Add(24 * time.Hour),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeWindow(tt.start, tt.end, tt.event)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeWindowEdgeCases(t *testing.T) {
	t.Run("exactly 24 hours", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.Add(24 * time.Hour)
		event := end

		err := validateTimeWindow(start, end, event)
		assert.NoError(t, err)
	})

	t.Run("24 hours minus 1 nanosecond", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.Add(24*time.Hour - time.Nanosecond)
		event := end

		err := validateTimeWindow(start, end, event)
		assert.Error(t, err)
	})

	t.Run("24 hours plus 1 nanosecond", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := start.Add(24*time.Hour + time.Nanosecond)
		event := end

		err := validateTimeWindow(start, end, event)
		assert.Error(t, err)
	})
}

func TestTimeZoneHandling(t *testing.T) {
	// All times should be in UTC to avoid DST issues
	tests := []struct {
		name      string
		timezone  string
		expectErr bool
	}{
		{
			name:      "UTC timezone",
			timezone:  "UTC",
			expectErr: false,
		},
		{
			name:      "EST timezone",
			timezone:  "America/New_York",
			expectErr: true, // Should use UTC only
		},
		{
			name:      "PST timezone",
			timezone:  "America/Los_Angeles",
			expectErr: true, // Should use UTC only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			require.NoError(t, err)

			start := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
			end := start.Add(24 * time.Hour)
			event := end

			err = validateTimeWindow(start, end, event)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// validateTimeWindow validates that a time window represents exactly 24 hours
// and that the event time equals the end time
func validateTimeWindow(start, end, event time.Time) error {
	// Check that end = start + 24h (exactly)
	expectedEnd := start.Add(24 * time.Hour)
	if !end.Equal(expectedEnd) {
		return fmt.Errorf("end time must be exactly 24 hours after start: got %v, expected %v", end, expectedEnd)
	}

	// Check that event_time = end
	if !event.Equal(end) {
		return fmt.Errorf("event time must equal end time: got %v, expected %v", event, end)
	}

	// Check that all times are in UTC
	if start.Location() != time.UTC {
		return fmt.Errorf("start time must be in UTC, got %v", start.Location())
	}
	if end.Location() != time.UTC {
		return fmt.Errorf("end time must be in UTC, got %v", end.Location())
	}
	if event.Location() != time.UTC {
		return fmt.Errorf("event time must be in UTC, got %v", event.Location())
	}

	return nil
}
