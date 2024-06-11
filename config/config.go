package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
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
type BiasSetMap map[string]float64

type DarkSetElementMap map[string]DarkSetMap
type DarkSetMap map[string]float64

// CoolingConfig is configuration about use the cameras cooler
type CoolingConfig struct {
	UseCooler       bool    //	Camera has cooler and we'll use it
	CoolTo          float64 //	Target temperature
	CoolStartTol    float64 //	Target plus-or-minus this
	CoolWaitMinutes int     //	How long to wait for target (minutes)
	AbortOnCooling  bool    //	Abort collection if temp rises
	CoolAbortTol    float64 //	Amount of temp rise before abort
	OffAtEnd        bool    //	Turn off cooler at end of session
}

// StartConfig is configuration about delayed start to the collection
type StartConfig struct {
	Delay bool   //	Should start be delayed?
	Day   string //	Day to start, yyyy-mm-dd or "today" or "tomorrow"
	Time  string //	Time to start, HH:MM 24-hour format
}

// ServerConfig is configuration to reach the TheSkyX server
type ServerConfig struct {
	Address string // IP, domain name, or localhost
	Port    int    // TCP port number
}

// BiasSet is the specification for one set of identical bias frames
type BiasSet struct {
	Frames  int //	Number of frames in this set
	Binning int //	Binning factor
}

// DarkSet is the specification for one set of identical bias frames
type DarkSet struct {
	Frames  int     //	Number of frames in this set
	Seconds float64 //	Exposure time in seconds
	Binning int     //	Binning factor
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
			Frames:  int(frames),
			Binning: int(binning),
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
			Frames:  int(frames),
			Seconds: seconds,
			Binning: int(binning),
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

//	ParseStart parses the string start time settings received from the
//	config file and returns whether a delay is wanted, and the start time
//	converted to a real Time object
//	day and time are checked only if delay=true
//	The day string can be one of:
//		Empty or missing, indicating "today"
//		The word "today" or "tomorrow"
//		A date in yyyy-mm-dd format
//	The time string can be
//		a time in 24-hour HH:MM format
//		it can be empty if delay=false, otherwise it is required

func (config *SettingsType) ParseStart() (bool, time.Time, error) {
	if config.Start.Delay == false {
		//	Valid result, no checking of date needed
		return false, time.Time{}, nil
	}
	if config.Start.Time == "" {
		return false, time.Time{}, errors.New("missing start time")
	}
	if config.Start.Day == "" {
		config.Start.Day = "today"
	}
	config.Start.Day = strings.ToLower(config.Start.Day)
	if config.Start.Day == "today" {
		//	Get today's date in yyyy-mm-dd format
		today := time.Now()
		config.Start.Day = today.Format("2006-01-02")
	}
	if config.Start.Day == "tomorrow" {
		today := time.Now()
		tomorrow := today.AddDate(0, 0, 1)
		config.Start.Day = tomorrow.Format("2006-01-02")
	}
	converted, err := time.ParseInLocation(
		"2006-01-02 15:04",
		config.Start.Day+" "+config.Start.Time,
		time.Local)
	if err != nil {
		return false, time.Time{}, err
	}
	return true, converted, nil
}
