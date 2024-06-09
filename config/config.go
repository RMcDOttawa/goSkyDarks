package config

import (
	"errors"
	"fmt"
)

type SettingsType struct {
	Verbosity  int
	Debug      bool
	StateFile  string //	Path to state file
	Cooling    CoolingConfig
	Start      StartConfig
	Server     ServerConfig
	BiasFrames []BiasSetElementMap
	DarkFrames []DarkSetElementMap
}

type BiasSetElementMap map[string]BiasSetMap
type BiasSetMap map[string]int

type DarkSetElementMap map[string]DarkSetMap
type DarkSetMap map[string]int

// CoolingConfig is configuration about use the cameras cooler
type CoolingConfig struct {
	UseCooler       bool    //	Camera has cooler and we'll use it
	CoolTo          float32 //	Target temperature
	CoolStartTol    float32 //	Target plus-or-minus this
	CoolWaitMinutes int     //	How long to wait for target (minutes)
	AbortOnCooling  bool    //	Abort collection if temp rises
	CoolAbortTol    float32 //	Amount of temp rise before abort
}

// StartConfig is configuration about delayed start to the collection
type StartConfig struct {
	Delay bool
	Day   string
	Time  string
}

// ServerConfig is configuration to reach the TheSkyX server
type ServerConfig struct {
	Address string // IP, domain name, or localhost
	Port    int    // TCP port number
}

// BiasSet is the specification for one set of identical bias frames
type BiasSet struct {
	Frames  int
	Binning int
}

// DarkSet is the specification for one set of identical bias frames
type DarkSet struct {
	Frames  int
	Seconds int
	Binning int
}

// GetBiasSets retrieves the list of bias frames requested, converting them from the
// awkward internal map format to a simple slice of BiasSet structs
func (config *SettingsType) GetBiasSets() []BiasSet {
	result := make([]BiasSet, 0, len(config.BiasFrames))
	//fmt.Printf("ConvertBiasMap: %#v\n", theMap)
	for _, mapEl := range config.BiasFrames {
		//fmt.Println("  Handling", key, mapEl)
		frames := mapEl["biasset"]["frames"]
		binning := mapEl["biasset"]["binning"]
		result = append(result, BiasSet{
			Frames:  frames,
			Binning: binning,
		})
	}
	return result
}

// GetDarkSets retrieves the list of dark frames requested, converting them from the
// awkward internal map format to a simple slice of DarkSet structs
func (config *SettingsType) GetDarkSets() []DarkSet {
	result := make([]DarkSet, 0, len(config.DarkFrames))
	//fmt.Printf("ConvertBiasMap: %#v\n", theMap)
	for _, mapEl := range config.DarkFrames {
		//fmt.Println("  Handling", key, mapEl)
		frames := mapEl["darkset"]["frames"]
		seconds := mapEl["darkset"]["seconds"]
		binning := mapEl["darkset"]["binning"]
		result = append(result, DarkSet{
			Frames:  frames,
			Seconds: seconds,
			Binning: binning,
		})
	}
	return result
}

// ValidateGlobals validates any global settings
func (config *SettingsType) ValidateGlobals() error {
	//	Verbosity must be between 0 and 5
	if config.Verbosity < 0 || config.Verbosity > 5 {
		return errors.New(fmt.Sprintf("invalid verbosity level (%d); must be between 0 and 5", config.Verbosity))
	}
	return nil
}

//// ParseStartTime Parses the start time string and returns whether we are delaying start, and to when
//// Start time could be:
////
////	Empty string, in which case we are not delaying start
////	A time in hh:mm (24-hour) format, meaning we start at that time today
////	A date indicator, comma, time.
////		Date indicator is one of:
////			today
////			tomorrow
////			date in yyyy-mm-dd format
//func ParseStartTime(startAtString string) (bool, time.Time, error) {
//	//fmt.Println("Parsing starttime:", startAtString)
//	if startAtString == "" {
//		return false, time.Time{}, nil
//	}
//	startAtString = strings.ToLower(startAtString)
//	parts := strings.Split(startAtString, ",")
//	var dayString string
//	var timeString string
//	if len(parts) == 1 {
//		dayString = "today"
//		timeString = parts[0]
//	} else if len(parts) == 2 {
//		dayString = parts[0]
//		timeString = parts[1]
//	} else {
//		return false, time.Time{}, errors.New("invalid start time")
//	}
//	if dayString == "today" {
//		//	Get today's date in yyyy-mm-dd format
//		today := time.Now()
//		dayString = today.Format("2006-01-02")
//	}
//	if dayString == "tomorrow" {
//		today := time.Now()
//		tomorrow := today.AddDate(0, 0, 1)
//		dayString = tomorrow.Format("2006-01-02")
//	}
//
//	converted, err := time.Parse("2006-01-02 15:04", dayString+" "+timeString)
//	if err != nil {
//		return false, time.Time{}, err
//	}
//	return true, converted, nil
//}
//
//// ParseBiasStrings Parses the bias frame strings slice, and returns a list of bias frame config (if valid)
//// Each bias frame string is a pair of comma-separated integers
////
////	The first is the number of frames, so > 0
////	The second is the binning level, which must be an integer from 1 to 8
//func ParseBiasStrings(biasStrings []string) ([]BiasSpec, error) {
//	fmt.Println("ParseBiasStrings:", biasStrings)
//	outputSpecs := make([]BiasSpec, 0, len(biasStrings))
//	for _, biasString := range biasStrings {
//		//fmt.Println("   Parsing", biasString)
//		parts := strings.Split(biasString, ",")
//		if len(parts) != 2 {
//			return outputSpecs, errors.New(fmt.Sprintf("invalid bias specification \"%s\", format should be count,binning", biasString))
//		}
//		frameCount, err := strconv.Atoi(parts[0])
//		if err != nil {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in bias frame spec", parts[0]))
//		}
//		if frameCount < 1 {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in bias frame spec: must be > 0", parts[0]))
//		}
//		binning, err := strconv.Atoi(parts[1])
//		if err != nil {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in bias frame spec", parts[1]))
//		}
//		if binning < 1 || binning > 8 {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in bias frame spec: must be 1 - 8", parts[1]))
//		}
//		biasItem := BiasSpec{
//			number:  frameCount,
//			binning: binning,
//		}
//		outputSpecs = append(outputSpecs, biasItem)
//	}
//	//fmt.Println("Returning config:", outputSpecs)
//	return outputSpecs, nil
//}
//
//// ParseDarkStrings Parses the dark frame strings slice, and returns a list of dark frame config (if valid)
//// Each dark frame string is a triplet of comma-separated integers
////
////	The first is the number of frames, so > 0
////	The second is the exposure length in seconds, so integer > 0
////	The third is the binning level, which must be an integer from 1 to 8
//func ParseDarkStrings(darkStrings []string) ([]DarkSpec, error) {
//	fmt.Println("ParseDarkStrings", darkStrings)
//	outputSpecs := make([]DarkSpec, 0, len(darkStrings))
//	for _, darkString := range darkStrings {
//		//fmt.Println("   Parsing", darkString)
//		parts := strings.Split(darkString, ",")
//		if len(parts) != 3 {
//			return outputSpecs, errors.New(fmt.Sprintf("invalid dark specification \"%s\", format should be count,seconds,binning", darkString))
//		}
//		frameCount, err := strconv.Atoi(parts[0])
//		if err != nil {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in dark frame spec", parts[0]))
//		}
//		if frameCount < 1 {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid count \"%s\", in dark frame spec: must be > 0", parts[0]))
//		}
//		exposure, err := strconv.ParseFloat(parts[1], 32)
//		if err != nil {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid exposure \"%s\", in dark frame spec", parts[1]))
//		}
//		if exposure <= 0.0 {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid exposure \"%s\", in dark frame spec: must be > 0", parts[1]))
//		}
//		binning, err := strconv.Atoi(parts[2])
//		if err != nil {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in dark frame spec", parts[2]))
//		}
//		if binning < 1 || binning > 8 {
//			return outputSpecs, errors.New(fmt.Sprintf("Invalid binning \"%s\", in dark frame spec: must be 1 - 8", parts[2]))
//		}
//		darkItem := DarkSpec{
//			number:  frameCount,
//			seconds: float32(exposure),
//			binning: binning,
//		}
//		outputSpecs = append(outputSpecs, darkItem)
//	}
//	//fmt.Println("Returning config:", outputSpecs)
//	return outputSpecs, nil
//}