/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/dchest/validator"
	"github.com/spf13/viper"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var CaptureConfig struct {
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
	delayStart      bool      //	Calculated from presence of startat string
	startTime       time.Time // Date and time to start the capture run
	biasFrames      []BiasSpec
	darkFrames      []DarkSpec
}

type BiasSpec struct {
	number  int
	binning int
}

type DarkSpec struct {
	number  int
	seconds float32
	binning int
}

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Use TheSkyX to capture calibration frames",
	Long: `Uses TheSkyX to capture dark and bias frames as specified in the command flags or configuration file.  
If the state file indicates that a previous run was terminated but unfinished, capture will pick up from where the previous run left off.  
Use the RESET command to prevent this and start over.

To delay the start until later, use --startat <day>,<time> or just --startat <time> (which assumes today)
<day> can be "today", "tomorrow", or a date in yyyy-mm-dd form
`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateSettings()

		if Debug || Verbosity >= 1 {
			fmt.Println("Capture command entered")
			if Debug || Verbosity >= 4 {
				DisplayFlags()
				if CaptureConfig.delayStart {
					fmt.Println("Delaying start to", CaptureConfig.startTime)
				}
			}
		}

	},
}

func ValidateSettings() {
	err := validateConfig()
	if err != nil {
		fmt.Println("Error in setting:", err)
		os.Exit(1)
	}

	CaptureConfig.delayStart, CaptureConfig.startTime, err = ParseStartTime(viper.GetString("startat"))
	if err != nil {
		fmt.Println("Error in setting:", err)
		os.Exit(1)
	}

	CaptureConfig.biasFrames, err = ParseBiasStrings(CaptureConfig.BiasStrings)
	if err != nil {
		fmt.Println("Error in specified bias string:", err)
		os.Exit(1)
	}

	CaptureConfig.darkFrames, err = ParseDarkStrings(CaptureConfig.DarkStrings)
	if err != nil {
		fmt.Println("Error in specified dark string:", err)
		os.Exit(1)
	}

	if len(CaptureConfig.biasFrames) == 0 && len(CaptureConfig.darkFrames) == 0 {
		fmt.Println("At least one bias or dark frame set must be specified")
		os.Exit(1)
	}
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
			number:  frameCount,
			binning: binning,
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
			number:  frameCount,
			seconds: float32(exposure),
			binning: binning,
		}
		outputSpecs = append(outputSpecs, darkItem)
	}
	//fmt.Println("Returning specs:", outputSpecs)
	return outputSpecs, nil
}

func validateConfig() error {
	//	Server address must be a valid IP address or domain name
	err := validateServer(viper.GetString("server"))
	if err != nil {
		return err
	}

	//	Port number in range
	port := viper.GetInt("port")
	if port < 0 || port > 65535 {
		return errors.New("port must be between 0 and 65535")
	}

	//	Cool To value in range -200 to +200
	coolTo := viper.GetFloat64("coolto")
	const minCoolTo = -200.0
	const maxCoolTo = 200.0
	if coolTo < minCoolTo || coolTo > maxCoolTo {
		return errors.New(fmt.Sprintf("coolto must be between %g and %g (degrees)", minCoolTo, maxCoolTo))
	}

	//	Cool Start Tolerance value in range 0 to 50
	coolToStartTol := viper.GetFloat64("coolstarttol")
	const minStartTol = 0.0
	const maxStartTol = 50.0
	if coolToStartTol < minStartTol || coolToStartTol > maxStartTol {
		return errors.New(fmt.Sprintf("coolstarttol must be between %g and %g (degrees)", minStartTol, maxStartTol))
	}

	//	Cool Abort Tolerance value in range 0 to 50
	coolAbortTol := viper.GetFloat64("coolaborttol")
	const minAbortTol = 0.0
	const maxAbortTol = 50.0
	if coolAbortTol < minAbortTol || coolAbortTol > maxAbortTol {
		return errors.New(fmt.Sprintf("coolaborttol must be between %g and %g (degrees)", minAbortTol, maxAbortTol))
	}

	//	Cooling wait minutes in range 1 to (24 hours)
	coolingWaitMinutes := viper.GetInt("coolwait")
	const minCoolWait = 1
	const maxCoolWait = 24 * 60
	if coolingWaitMinutes < minCoolWait || coolingWaitMinutes > maxCoolWait {
		return errors.New(fmt.Sprintf("coolwait must be between %d and %d (minutes)", minCoolWait, maxCoolWait))
	}

	return nil
}

func validateServer(addressString string) error {
	tryIp := net.ParseIP(addressString)
	if tryIp != nil {
		//	Successful parse, so it's a valid IP.  Return nil for no error
		return nil
	}
	//	Not a valid IP.  See if it's a (syntactically) valid domain name.
	addressString = strings.ToLower(addressString)
	if addressString == "localhost" {
		return nil
	}
	//	We won't try to actually connect - leave that for the capture process
	if validator.IsValidDomain(addressString) {
		return nil
	}
	return errors.New(fmt.Sprintf("Invalid server address: %s", addressString))
}

func init() {
	rootCmd.AddCommand(captureCmd)

	captureCmd.PersistentFlags().StringVarP(&CaptureConfig.ServerAddress, "server", "s", "localhost", "Address of TheSkyX server")
	err := viper.BindPFlag("server", captureCmd.PersistentFlags().Lookup("server"))
	if err != nil {
		fmt.Println("Error binding server flag:", err)
	}

	captureCmd.PersistentFlags().IntVarP(&CaptureConfig.ServerPort, "port", "p", 3040, "Port number of TheSkyX server")
	err = viper.BindPFlag("port", captureCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		fmt.Println("Error binding port flag:", err)
	}

	captureCmd.PersistentFlags().BoolVarP(&CaptureConfig.UseCooler, "cool", "", false, "Use camera cooler")
	err = viper.BindPFlag("cool", captureCmd.PersistentFlags().Lookup("cool"))
	if err != nil {
		fmt.Println("Error binding cool flag:", err)
	}

	captureCmd.PersistentFlags().Float32VarP(&CaptureConfig.CoolTo, "coolto", "", -10.0, "Cool to target temperature")
	err = viper.BindPFlag("coolto", captureCmd.PersistentFlags().Lookup("coolto"))
	if err != nil {
		fmt.Println("Error binding coolto flag:", err)
	}

	captureCmd.PersistentFlags().Float32VarP(&CaptureConfig.CoolStartTol, "coolstarttol", "", 1.0, "Cooling start tolerance")
	err = viper.BindPFlag("coolstarttol", captureCmd.PersistentFlags().Lookup("coolstarttol"))
	if err != nil {
		fmt.Println("Error binding coolstarttol flag:", err)
	}

	captureCmd.PersistentFlags().IntVarP(&CaptureConfig.CoolWaitMinutes, "coolwait", "", 30, "Maximum minutes to reach temperature")
	err = viper.BindPFlag("coolwait", captureCmd.PersistentFlags().Lookup("coolwait"))
	if err != nil {
		fmt.Println("Error binding coolwait flag:", err)
	}

	captureCmd.PersistentFlags().BoolVarP(&CaptureConfig.CoolAbort, "coolabort", "", false, "Abort capture if cooling leaves range")
	err = viper.BindPFlag("coolabort", captureCmd.PersistentFlags().Lookup("coolabort"))
	if err != nil {
		fmt.Println("Error binding coolabort flag:", err)
	}

	captureCmd.PersistentFlags().Float32VarP(&CaptureConfig.CoolAbortTol, "coolaborttol", "", 3.0, "Cooling capture abort tolerance")
	err = viper.BindPFlag("coolaborttol", captureCmd.PersistentFlags().Lookup("coolaborttol"))
	if err != nil {
		fmt.Println("Error binding coolaborttol flag:", err)
	}

	captureCmd.PersistentFlags().StringVarP(&CaptureConfig.StartAtString, "startat", "", "", "Time to start capture")
	err = viper.BindPFlag("startat", captureCmd.PersistentFlags().Lookup("startat"))
	if err != nil {
		fmt.Println("Error binding startat flag:", err)
	}

	CaptureConfig.BiasStrings = captureCmd.PersistentFlags().StringArray("bias", []string{}, "Bias frame (count,binning) - can repeat multiple times")
	err = viper.BindPFlag("bias", captureCmd.PersistentFlags().Lookup("bias"))
	if err != nil {
		fmt.Println("Error binding bias flag:", err)
	}

	CaptureConfig.DarkStrings = captureCmd.PersistentFlags().StringArray("dark", []string{}, "Dark frame (count,seconds,binning) - can repeat multiple times")
	err = viper.BindPFlag("dark", captureCmd.PersistentFlags().Lookup("dark"))
	if err != nil {
		fmt.Println("Error binding dark flag:", err)
	}

}
