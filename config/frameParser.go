package config

import (
	"errors"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

// Parse string in the form a,b,c into 3 numbers.
// a    number of exposures.  An integer > 0
// b    exposure time, a float > 0
// c    binning, an integer > 0 (surprising if it wasn't small, like from 1 to 4)
func ParseDarkSet(darkSet string) (int, float64, int, error) {
	parts := strings.Split(darkSet, ",")
	if len(parts) != 3 {
		return 0, 0.0, 0, errors.New("dark set must have 3 parts: count,exposure,time")
	}
	// Parse the count
	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0.0, 0, errors.New("Error in dark set count: " + err.Error())
	}
	if count < 1 {
		return 0, 0.0, 0, errors.New("dark set count must be > 0")
	}

	// Parse the exposure time
	exposure, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0.0, 0, errors.New("error in dark set exposure time: " + err.Error())
	}
	if exposure <= 0 {
		return 0, 0.0, 0, errors.New("dark set exposure time must be > 0")
	}

	// Parse the binning
	binning, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return 0, 0.0, 0, errors.New("error in dark set binning: " + err.Error())
	}
	if binning < 1 {
		return 0, 0.0, 0, errors.New("dark set binning must be > 0")
	}

	return count, exposure, binning, nil
}

//		Parse string in the form a,b into 2 numbers.
//		a    number of exposures.  An integer > 0
//	    b    binning, an integer > 0 (surprising if it wasn't small, like from 1 to 4)
func ParseBiasSet(darkSet string) (int, int, error) {
	parts := strings.Split(darkSet, ",")
	if len(parts) != 2 {
		return 0, 0, errors.New("Bias set must have 2 parts: count,time")
	}

	// Parse the count
	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, errors.New("Error in bias set count: " + err.Error())
	}
	if count < 1 {
		return 0, 0, errors.New("bias set count must be > 0")
	}

	// Parse the binning
	binning, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, errors.New("error in bias set binning: " + err.Error())
	}
	if binning < 1 {
		return 0, 0, errors.New("dark set binning must be > 0")
	}

	return count, binning, nil
}

// Determine if the named flag was explicitly set in the command line
func FlagExplicitlySet(cmd *cobra.Command, flagName string) bool {
	lookup := cmd.Flags().Lookup(flagName)
	if lookup == nil {
		return false
	}
	return lookup.Changed
}
