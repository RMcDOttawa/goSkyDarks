package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type SettingsType struct {
	Verbosity    int
	Debug        bool
	StateFile    string //	Path to state file
	ShowSettings bool
	Cooling      CoolingConfig
	Start        StartConfig
	Server       ServerConfig
	BiasFrames   []string
	DarkFrames   []string
}

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

// Keys to retrieve settings from viper

const VerbositySetting = "verbosity"
const DebugSetting = "debug"
const StateFileSetting = "statefile"
const ShowSettingsSetting = "ShowSettings"
const UseCoolerSetting = "Cooling.UseCooler"
const CoolToSetting = "Cooling.CoolTo"
const CoolStartTolSetting = "Cooling.CoolStartTol"
const CoolWaitMinutesSetting = "Cooling.CoolWaitMinutes"
const AbortOnCoolingSetting = "Cooling.AbortOnCooling"
const CoolAbortTolSetting = "Cooling.CoolAbortTol"
const CoolerOffAtEndSetting = "Cooling.OffAtEnd"
const StartDelaySetting = "Start.Delay"
const StartDaySetting = "Start.Day"
const StartTimeSetting = "Start.Time"
const ServerAddressSetting = "Server.Address"
const ServerPortSetting = "Server.Port"
const BiasFramesSetting = "BiasFrames"
const DarkFramesSetting = "DarkFrames"

func ShowAllSettings() {
	fmt.Println("Validating and displaying all config settings:")

	//	Global settings
	fmt.Println("Global settings")
	fmt.Printf("   Show Settings: %t\n", viper.GetBool(ShowSettingsSetting))
	fmt.Printf("   Verbosity: %d\n", viper.GetInt(VerbositySetting))
	fmt.Printf("   Debug: %t\n", viper.GetBool(DebugSetting))
	fmt.Printf("   State File Path: %s\n", viper.GetString(StateFileSetting))

	//	Server settings
	fmt.Println("Server settings")
	fmt.Printf("   Address: %s\n", viper.GetString(ServerAddressSetting))
	fmt.Printf("   Port: %d\n", viper.GetInt(ServerPortSetting))

	//	Start time
	fmt.Println("Delayed Start settings")
	fmt.Printf("   Delay: %t\n", viper.GetBool(StartDelaySetting))
	fmt.Printf("   Day: %s\n", viper.GetString(StartDaySetting))
	fmt.Printf("   Time: %s\n", viper.GetString(StartTimeSetting))
	delay, start, err := ParseStart()
	if err != nil {
		fmt.Printf("Error parsing start settings: %s\n", err)
	}
	fmt.Printf("   Converted to: %t, %v\n", delay, start)

	//	Cooling info
	fmt.Println("Cooling settings")
	fmt.Printf("   Use cooler: %t\n", viper.GetBool(UseCoolerSetting))
	fmt.Printf("   Cool to: %g degrees\n", viper.GetFloat64(CoolToSetting))
	fmt.Printf("   Start tolerance: %g degrees\n", viper.GetFloat64(CoolStartTolSetting))
	fmt.Printf("   Wait maximum: %d minutes\n", viper.GetInt(CoolWaitMinutesSetting))
	fmt.Printf("   Abort if cooling outside tolerance: %t\n", viper.GetBool(AbortOnCoolingSetting))
	fmt.Printf("   Abort tolerance: %g degrees\n", viper.GetFloat64(CoolAbortTolSetting))
	fmt.Printf("   Turn off cooler at end of session: %t\n", viper.GetBool(CoolerOffAtEndSetting))

	//	Bias Frames
	fmt.Println("Bias Frames")
	for _, frameSetString := range viper.GetStringSlice(BiasFramesSetting) {
		count, binning, err := ParseBiasSet(frameSetString)
		if err != nil {
			fmt.Println("   Syntax error in set:", frameSetString)
		} else {
			fmt.Printf("   %d bias frames at %d x %d binning\n", count, binning, binning)
		}
	}

	//	Dark Frames
	fmt.Println("Dark Frames")
	for _, frameSetString := range viper.GetStringSlice(DarkFramesSetting) {
		count, exposure, binning, err := ParseDarkSet(frameSetString)
		if err != nil {
			fmt.Println("   Syntax error in set:", frameSetString)
		} else {
			fmt.Printf("   %d dark frames of %.2f seconds at %d x %d binning\n", count, exposure, binning, binning)
		}
	}
}

// ValidateGlobals validates any global settings
func ValidateGlobals() error {
	//	Verbosity must be between 0 and 5
	verbosity := viper.GetInt(VerbositySetting)
	if verbosity < 0 || verbosity > 5 {
		return errors.New(fmt.Sprintf("invalid verbosity level (%d); must be between 0 and 5", verbosity))
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

func ParseStart() (bool, time.Time, error) {
	startDelay := viper.GetBool(StartDelaySetting)
	startDay := viper.GetString(StartDaySetting)
	startTime := viper.GetString(StartTimeSetting)
	if startDelay == false {
		//	Valid result, no checking of date needed
		return false, time.Time{}, nil
	}
	if startTime == "" {
		return false, time.Time{}, errors.New("missing start time")
	}
	if startDay == "" {
		startDay = "today"
	}
	startDay = strings.ToLower(startDay)
	if startDay == "today" {
		//	Get today's date in yyyy-mm-dd format
		today := time.Now()
		startDay = today.Format("2006-01-02")
	}
	if startDay == "tomorrow" {
		today := time.Now()
		tomorrow := today.AddDate(0, 0, 1)
		startDay = tomorrow.Format("2006-01-02")
	}
	converted, err := time.ParseInLocation(
		"2006-01-02 15:04",
		startDay+" "+startTime,
		time.Local)
	if err != nil {
		return false, time.Time{}, err
	}
	return true, converted, nil
}
