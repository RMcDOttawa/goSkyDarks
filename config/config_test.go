package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParseStartTime(t *testing.T) {
	now := time.Now()
	todayString := now.Format("2006-01-02")

	//	******* Valid cases *******

	//	No delaypkg
	t.Run("delaypkg false", func(t *testing.T) {
		useStart, _, err := ParseStartTime(false, "", "")
		require.Nil(t, err)
		require.False(t, useStart, "delaypkg=false should parse as non-deferred start")
	})

	//	12:35
	t.Run("just 12:35", func(t *testing.T) {
		useStart, startTime, err := ParseStartTime(true, "", "12:35")
		require.Nil(t, err)
		require.True(t, useStart, "12:35 should be parsed as a deferred start")
		require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, "12:35 didn't parse to correct time")
		require.Equal(t, todayString, startTime.Format("2006-01-02"), "12:35 didn't parse to today's date")
	})

	//	today,22:15
	t.Run("today 12:35", func(t *testing.T) {
		useStart, startTime, err := ParseStartTime(true, "today", "12:35")
		require.Nil(t, err)
		require.True(t, useStart, "today,12:35 should be parsed as a deferred start")
		require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, "today,12:35 didn't parse to correct time")
		require.Equal(t, todayString, startTime.Format("2006-01-02"), "today,12:35 didn't parse to today's date")
	})

	//	tomorrow,08:00
	t.Run("tomorrow 08:00", func(t *testing.T) {
		tomorrow := now.AddDate(0, 0, 1)
		tomorrowString := tomorrow.Format("2006-01-02")
		useStart, startTime, err := ParseStartTime(true, "tomorrow", "08:00")
		require.Nil(t, err)
		require.True(t, useStart, "tomorrow 08:005 should be parsed as a deferred start")
		require.True(t, startTime.Hour() == 8 && startTime.Minute() == 0, "tomorrow 08:00 didn't parse to correct time")
		require.Equal(t, tomorrowString, startTime.Format("2006-01-02"), "tomorrow 08:00 didn't parse to today's date")
	})

	//	2155-07-01,09:00
	t.Run("future date 08:00", func(t *testing.T) {
		const arbitraryFutureDate = "2155-07-01"
		useStart, startTime, err := ParseStartTime(true, arbitraryFutureDate, "08:00")
		require.Nil(t, err)
		require.True(t, useStart, fmt.Sprintf("%s 08:00 should be parsed as a deferred start", arbitraryFutureDate))
		require.True(t, startTime.Hour() == 8 && startTime.Minute() == 0, fmt.Sprintf("%s 08:00 didn't parse to correct time", arbitraryFutureDate))
		require.Equal(t, arbitraryFutureDate, startTime.Format("2006-01-02"), fmt.Sprintf("%s 08:00 didn't parse to correct date", arbitraryFutureDate))
	})

	//	******* Invalid cases that should produce an error *******

	//	missing start time
	t.Run("missing start time", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "", "")
		require.NotNil(t, err, "missing start time should have produced an error")
		require.ErrorContains(t, err, "missing start time")
		require.False(t, useStart, "missing start time should be parsed as not deferred start")
	})

	//	bad-keyword
	t.Run("bad day keyword", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "nonsense", "08:30")
		require.NotNil(t, err, "Bad date keyword should have produced error")
		require.ErrorContains(t, err, "cannot parse")
		require.False(t, useStart, "invalid keyword should not be parsed as a deferred start")
	})

	//	Bad date: 2024-17-33,10:00
	t.Run("invalid date", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "2024-27-33", "10:00")
		require.NotNil(t, err, "invalid date should have produced error")
		require.ErrorContains(t, err, "month out of range")
		require.False(t, useStart, "invalid date should not be parsed as a deferred start")
	})

	//	Bad time: today,47:00
	t.Run("invalid time hour", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "today", "47:00")
		require.NotNil(t, err, "invalid time (hour) should have produced error")
		require.ErrorContains(t, err, "hour out of range")
		require.False(t, useStart, "invalid time (hour) should not be parsed as a deferred start")
	})

	t.Run("invalid time minute", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "today", "17:99")
		require.NotNil(t, err, "invalid time (minute) should have produced error")
		require.ErrorContains(t, err, "minute out of range")
		require.False(t, useStart, "invalid time (minute) should not be parsed as a deferred start")
	})

	t.Run("invalid time junk string", func(t *testing.T) {
		useStart, _, err := ParseStartTime(true, "today", "nonsense")
		require.NotNil(t, err, "invalid time should have produced error")
		require.ErrorContains(t, err, "cannot parse")
		require.False(t, useStart, "invalid time should not be parsed as a deferred start")
	})
}

func ParseStartTime(delay bool, day string, time string) (bool, time.Time, error) {
	viper.Set(StartDelaySetting, delay)
	viper.Set(StartDaySetting, day)
	viper.Set(StartTimeSetting, time)
	return ParseStart()
}
