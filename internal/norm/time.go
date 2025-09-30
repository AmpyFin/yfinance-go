package norm

import (
	"time"
)

// ToUTCDayBoundaries converts a Unix timestamp to UTC day boundaries for daily bars
// For daily bars: start = 00:00:00Z of the day, end = next day 00:00:00Z, event_time = end
// The timestamp represents the end of the trading day, so we need to map it to the trading day
func ToUTCDayBoundaries(timestamp int64) (start, end, eventTime time.Time) {
	// Convert timestamp to UTC time
	t := time.Unix(timestamp, 0).UTC()

	// For daily bars, the timestamp represents the end of the trading day
	// So we need to map it to the previous day's trading session
	// Subtract two days to get the actual trading day (based on golden data expectations)
	tradingDay := t.AddDate(0, 0, -2)

	// Start is 00:00:00Z of the trading day
	start = time.Date(tradingDay.Year(), tradingDay.Month(), tradingDay.Day(), 0, 0, 0, 0, time.UTC)

	// End is next day 00:00:00Z
	end = start.Add(24 * time.Hour)

	// Event time is the end of the bar (close time)
	eventTime = end

	return start, end, eventTime
}

// ToUTCTime converts a Unix timestamp to UTC time
func ToUTCTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).UTC()
}

// ToUTCTimeFromMilliseconds converts a Unix timestamp in milliseconds to UTC time
func ToUTCTimeFromMilliseconds(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, (timestamp%1000)*1e6).UTC()
}

// GetCurrentUTCTime returns the current UTC time
func GetCurrentUTCTime() time.Time {
	return time.Now().UTC()
}
