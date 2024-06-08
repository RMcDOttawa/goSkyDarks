package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestParseStartTime(t *testing.T) {
	now := time.Now()
	todayString := now.Format("2006-01-02")

	//	******* Valid cases *******

	//	Null string
	useStart, startTime, err := ParseStartTime("")
	require.Nil(t, err)
	require.False(t, useStart, "empty string should parse as non-deferred start")

	//	12:35
	useStart, startTime, err = ParseStartTime("12:35")
	require.Nil(t, err)
	require.True(t, useStart, "12:35 should be parsed as a deferred start")
	require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, "12:35 didn't parse to correct time")
	require.Equal(t, todayString, startTime.Format("2006-01-02"), "12:35 didn't parse to today's date")

	//	today,22:15
	useStart, startTime, err = ParseStartTime("today,12:35")
	require.Nil(t, err)
	require.True(t, useStart, "today,12:35 should be parsed as a deferred start")
	require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, "today,12:35 didn't parse to correct time")
	require.Equal(t, todayString, startTime.Format("2006-01-02"), "today,12:35 didn't parse to today's date")

	//	tomorrow,08:00
	tomorrow := now.AddDate(0, 0, 1)
	tomorrowString := tomorrow.Format("2006-01-02")
	useStart, startTime, err = ParseStartTime("tomorrow,12:35")
	require.Nil(t, err)
	require.True(t, useStart, "tomorrow,12:35 should be parsed as a deferred start")
	require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, "tomorrow,12:35 didn't parse to correct time")
	require.Equal(t, tomorrowString, startTime.Format("2006-01-02"), "tomorrowString,12:35 didn't parse to tomorrow's date")

	//	2155-07-01,09:00
	const arbitraryFutureDate = "2155-07-01"
	useStart, startTime, err = ParseStartTime(fmt.Sprintf("%s,12:35", arbitraryFutureDate))
	require.Nil(t, err)
	require.True(t, useStart, fmt.Sprintf("%s,12:35 should be parsed as a deferred start", arbitraryFutureDate))
	require.True(t, startTime.Hour() == 12 && startTime.Minute() == 35, fmt.Sprintf("%s,12:35 didn't parse to correct time", arbitraryFutureDate))
	require.Equal(t, arbitraryFutureDate, startTime.Format("2006-01-02"), fmt.Sprintf("%s,12:35 didn't parse to correct date", arbitraryFutureDate))

	//	******* Invalid cases that should produce an error *******

	//	bad-keyword
	useStart, startTime, err = ParseStartTime("nonsense")
	require.NotNil(t, err)
	require.False(t, useStart, "invalid keyword should not be parsed as a deferred start")

	//	too many keywords
	useStart, startTime, err = ParseStartTime("today,12:35,surplus")
	require.NotNil(t, err)
	require.False(t, useStart, "string with surplus keywords should not be parsed as a deferred start")

	//	2024-17-33,10:00
	useStart, startTime, err = ParseStartTime("2024-17-33,10:00")
	require.NotNil(t, err)
	require.False(t, useStart, "invalid date should not be parsed as a deferred start")

	//	today,47:00
	useStart, startTime, err = ParseStartTime("today,47:00")
	require.NotNil(t, err)
	require.False(t, useStart, "invalid time should not be parsed as a deferred start")

	//	today,12:99
	useStart, startTime, err = ParseStartTime("today,12:99")
	require.NotNil(t, err)
	require.False(t, useStart, "invalid time should not be parsed as a deferred start")

	//	today,nonsense
	useStart, startTime, err = ParseStartTime("today,nonsense")
	require.NotNil(t, err)
	require.False(t, useStart, "invalid time should not be parsed as a deferred start")
}

func TestParseBiasStrings(t *testing.T) {
	//	Empty list
	biasSpecs, err := ParseBiasStrings(&[]string{})
	require.Nil(t, err, "Empty bias string should be considered valid")
	require.Equal(t, 0, len(biasSpecs), "Empty bias string should produce empty bias list")

	//	List of valid strings
	validInputs := []string{"10,1", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&validInputs)
	require.Nil(t, err, "List of valid bias strings should produce no error")
	require.Equal(t, len(validInputs), len(biasSpecs), "Bias list should be same length as input")
	require.True(t, reflect.DeepEqual(biasSpecs[0], BiasSpec{10, 1}), "First bias spec parsed incorrectly")
	require.True(t, reflect.DeepEqual(biasSpecs[1], BiasSpec{20, 2}), "2nd bias spec parsed incorrectly")
	require.True(t, reflect.DeepEqual(biasSpecs[2], BiasSpec{30, 3}), "3rd bias spec parsed incorrectly")

	//	Bad item count
	wrongCountInput := []string{"10,1,3", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&wrongCountInput)
	require.NotNil(t, err, "Bias item with wrong number of elements should fail")

	//	Invalid frame count: syntax
	badCountSyntaxInput := []string{"16wombat78,6", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&badCountSyntaxInput)
	require.NotNil(t, err, "Bias item with bad count syntax should fail")

	//	Invalid frame count: range
	badCountValueInput := []string{"0,6", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&badCountValueInput)
	require.NotNil(t, err, "Bias item with invalid count should fail")

	//	Invalid binning: syntax
	badBinningSyntaxInput := []string{"16,platypus", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&badBinningSyntaxInput)
	require.NotNil(t, err, "Bias item with bad binning syntax should fail")

	//	Invalid binning: range
	badBinningRangeInput := []string{"16,44", "20,2", "30,3"}
	biasSpecs, err = ParseBiasStrings(&badBinningRangeInput)
	require.NotNil(t, err, "Bias item with bad binning range should fail")

}

func TestParseDarkStrings(t *testing.T) {
	//	Empty list
	darkSpecs, err := ParseDarkStrings(&[]string{})
	require.Nil(t, err, "Empty dark string should be considered valid")
	require.Equal(t, 0, len(darkSpecs), "Empty dark string should produce empty bias list")

	//	List of valid strings
	validInputs := []string{"10,11,1", "20,22,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&validInputs)
	//fmt.Println("Parsed valid specs:", darkSpecs)
	require.Nil(t, err, "List of valid dark strings should produce no error")
	require.Equal(t, len(validInputs), len(darkSpecs), "Bias list should be same length as input")
	require.True(t, reflect.DeepEqual(darkSpecs[0], DarkSpec{10, 11.0, 1}), "First dark spec parsed incorrectly")
	require.True(t, reflect.DeepEqual(darkSpecs[1], DarkSpec{20, 22.0, 2}), "2nd dark spec parsed incorrectly")
	require.True(t, reflect.DeepEqual(darkSpecs[2], DarkSpec{30, 33.0, 3}), "3rd dark spec parsed incorrectly")

	//	Bad item count
	wrongCountInput := []string{"10,20,1,9", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&wrongCountInput)
	require.NotNil(t, err, "Dark item with wrong number of elements should fail")

	//	Invalid frame count: syntax
	badCountSyntaxInput := []string{"wombat,20,1", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badCountSyntaxInput)
	require.NotNil(t, err, "Dark item with bad count syntax should fail")

	////	Invalid frame count: range
	badCountValueInput := []string{"0,20,1", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badCountValueInput)
	require.NotNil(t, err, "Dark item with invalid count should fail")

	//	Invalid exposure seconds: syntax
	badExposureSyntaxInput := []string{"10,dingo,1", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badExposureSyntaxInput)
	//fmt.Println("bad exposure syntax error:", err)
	require.NotNil(t, err, "Dark item with bad exposure syntax should fail")

	//	Invalid exposure seconds: range
	badExposureValueInput := []string{"10,0,1", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badExposureValueInput)
	//fmt.Println("bad exposure value error:", err)
	require.NotNil(t, err, "Dark item with invalid exposure should fail")

	//	Invalid binning: syntax
	badBinningSyntaxInput := []string{"10,10,wagga", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badBinningSyntaxInput)
	//fmt.Println("bad binning syntax error:", err)
	require.NotNil(t, err, "Dark item with bad binning syntax should fail")

	////	Invalid binning: range
	badBinningValueInput := []string{"10,10,44", "20,21,2", "30,33,3"}
	darkSpecs, err = ParseDarkStrings(&badBinningValueInput)
	//fmt.Println("bad binning value error:", err)
	require.NotNil(t, err, "Dark item with bad binning value should fail")

}
