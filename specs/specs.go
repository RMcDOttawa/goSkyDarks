package specs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CaptureSpecs struct {
	ServerAddress   string
	ServerPort      int
	UseCooler       bool
	CoolTo          float32 // TEC target temperature
	CoolStartTol    float32 // Cooling tolerance to start capture (get this close)
	CoolWaitMinutes int     // How long willing to wait to reach target temp
	CoolAbort       bool    // Abort capture if cooling leaves range
	CoolAbortTol    float32 //	Drift from target temp to abort capture
	StartAtString   string  // When to start run.  Parseable time string
	BiasStrings     *[]string
	DarkStrings     *[]string
	DelayStart      bool      //	Calculated from presence of startat string
	StartTime       time.Time // Date and time to start the capture run
	BiasFrames      []BiasSpec
	DarkFrames      []DarkSpec
}

type BiasSpec struct {
	Number  int
	Binning int
}

type DarkSpec struct {
	Number  int
	Seconds float32
	Binning int
}

// ParseStartTime Parses the start time string and returns whether we are delaying start, and to when
// Start time could be:
//
//	Empty string, in which case we are not delaying start
//	A time in hh:mm (24-hour) format, meaning we start at that time today
//	A date indicator, comma, time.
//		Date indicator is one of:
//			today
//			tomorrow
//			date in yyyy-mm-dd format
func ParseStartTime(startAtString string) (bool, time.Time, error) {
	//fmt.Println("Parsing starttime:", startAtString)
	if startAtString == "" {
		return false, time.Time{}, nil
	}
	startAtString = strings.ToLower(startAtString)
	parts := strings.Split(startAtString, ",")
	var dayString string
	var timeString string
	if len(parts) == 1 {
		dayString = "today"
		timeString = parts[0]
	} else if len(parts) == 2 {
		dayString = parts[0]
		timeString = parts[1]
	} else {
		return false, time.Time{}, errors.New("invalid start time")
	}
	if dayString == "today" {
		//	Get today's date in yyyy-mm-dd format
		today := time.Now()
		dayString = today.Format("2006-01-02")
	}
	if dayString == "tomorrow" {
		today := time.Now()
		tomorrow := today.AddDate(0, 0, 1)
		dayString = tomorrow.Format("2006-01-02")
	}

	converted, err := time.Parse("2006-01-02 15:04", dayString+" "+timeString)
	if err != nil {
		return false, time.Time{}, err
	}
	return true, converted, nil
}

// ParseBiasStrings Parses the bias frame strings slice, and returns a list of bias frame specs (if valid)
// Each bias frame string is a pair of comma-separated integers
//
//	The first is the number of frames, so > 0
//	The second is the binning level, which must be an integer from 1 to 8
func ParseBiasStrings(biasStrings *[]string) ([]BiasSpec, error) {
	//fmt.Println("ParseBiasStrings")
	outputSpecs := make([]BiasSpec, 0, len(*biasStrings))
	for _, biasString := range *biasStrings {
		//fmt.Println("   Parsing", biasString)
		parts := strings.Split(biasString, ",")
		if len(parts) != 2 {
			return outputSpecs, errors.New(fmt.Sprintf("invalid bias specification \"%s\", format should be count,binning", biasString))
		}
		frameCount, err := strconv.Atoi(parts[0])
		if err != nil {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in bias frame spec", parts[0]))
		}
		if frameCount < 1 {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in bias frame spec: must be > 0", parts[0]))
		}
		binning, err := strconv.Atoi(parts[1])
		if err != nil {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in bias frame spec", parts[1]))
		}
		if binning < 1 || binning > 8 {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in bias frame spec: must be 1 - 8", parts[1]))
		}
		biasItem := BiasSpec{
			Number:  frameCount,
			Binning: binning,
		}
		outputSpecs = append(outputSpecs, biasItem)
	}
	//fmt.Println("Returning specs:", outputSpecs)
	return outputSpecs, nil
}

// ParseDarkStrings Parses the dark frame strings slice, and returns a list of dark frame specs (if valid)
// Each dark frame string is a triplet of comma-separated integers
//
//	The first is the number of frames, so > 0
//	The second is the exposure length in seconds, so integer > 0
//	The third is the binning level, which must be an integer from 1 to 8
func ParseDarkStrings(darkStrings *[]string) ([]DarkSpec, error) {
	//fmt.Println("ParseDarkStrings", *darkStrings)
	outputSpecs := make([]DarkSpec, 0, len(*darkStrings))
	for _, darkString := range *darkStrings {
		//fmt.Println("   Parsing", darkString)
		parts := strings.Split(darkString, ",")
		if len(parts) != 3 {
			return outputSpecs, errors.New(fmt.Sprintf("invalid dark specification \"%s\", format should be count,seconds,binning", darkString))
		}
		frameCount, err := strconv.Atoi(parts[0])
		if err != nil {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in dark frame spec", parts[0]))
		}
		if frameCount < 1 {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in dark frame spec: must be > 0", parts[0]))
		}
		exposure, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid exposure \"%s\", in dark frame spec", parts[1]))
		}
		if exposure <= 0.0 {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid exposure \"%s\", in dark frame spec: must be > 0", parts[1]))
		}
		binning, err := strconv.Atoi(parts[2])
		if err != nil {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in dark frame spec", parts[2]))
		}
		if binning < 1 || binning > 8 {
			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in dark frame spec: must be 1 - 8", parts[2]))
		}
		darkItem := DarkSpec{
			Number:  frameCount,
			Seconds: float32(exposure),
			Binning: binning,
		}
		outputSpecs = append(outputSpecs, darkItem)
	}
	//fmt.Println("Returning specs:", outputSpecs)
	return outputSpecs, nil
}
