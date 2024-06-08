/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/dchest/validator"
	"github.com/spf13/viper"
	"goskydarks/session"
	"goskydarks/specs"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var CaptureConfig specs.CaptureSpecs

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
			}
		}

		//	Create the capture session
		session, err := session.NewSession()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	get ready for capture by waiting for the start time (if requested)
		//	and cooling the camera (if requested)
		err = session.PrepareForCapture(CaptureConfig)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	Do the captures until finished or aborted

	},
}

func ValidateSettings() {
	err := validateConfig()
	if err != nil {
		fmt.Println("Error in setting:", err)
		os.Exit(1)
	}

	CaptureConfig.DelayStart, CaptureConfig.StartTime, err = specs.ParseStartTime(viper.GetString("startat"))
	if err != nil {
		fmt.Println("Error in setting:", err)
		os.Exit(1)
	}

	CaptureConfig.BiasFrames, err = specs.ParseBiasStrings(CaptureConfig.BiasStrings)
	if err != nil {
		fmt.Println("Error in specified bias string:", err)
		os.Exit(1)
	}

	CaptureConfig.DarkFrames, err = specs.ParseDarkStrings(CaptureConfig.DarkStrings)
	if err != nil {
		fmt.Println("Error in specified dark string:", err)
		os.Exit(1)
	}

	if len(CaptureConfig.BiasFrames) == 0 && len(CaptureConfig.DarkFrames) == 0 {
		fmt.Println("At least one bias or dark frame set must be specified")
		os.Exit(1)
	}
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
