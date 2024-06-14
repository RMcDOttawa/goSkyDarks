/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"goskydarks/config"
	"goskydarks/session"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the config file",
	Long:  `Validates the config file and displays all the settings.  No capture is performed.`,
	Run:   runValidateCommand,
}

func runValidateCommand(_ *cobra.Command, _ []string) {
	fmt.Println("Validating and displaying all config settings:")

	//	Global settings
	fmt.Println("Global settings")
	fmt.Printf("   Verbosity: %d\n", Settings.Verbosity)
	fmt.Printf("   Debug: %t\n", Settings.Debug)
	fmt.Printf("   State File Path: %s\n", Settings.StateFile)

	//	Server settings
	fmt.Println("Server settings")
	fmt.Printf("   Address: %s\n", Settings.Server.Address)
	fmt.Printf("   Port: %d\n", Settings.Server.Port)

	//	Start time
	fmt.Println("Delayed Start settings")
	fmt.Printf("   Delay: %t\n", Settings.Start.Delay)
	fmt.Printf("   Day: %s\n", Settings.Start.Day)
	fmt.Printf("   Time: %s\n", Settings.Start.Time)
	delay, start, err := config.ParseStart()
	if err != nil {
		fmt.Printf("Error parsing start settings: %s\n", err)
	}
	fmt.Printf("   Converted to: %t, %v\n", delay, start)

	//	Cooling info
	fmt.Println("Cooling settings")
	fmt.Printf("   Use cooler: %t\n", Settings.Cooling.UseCooler)
	fmt.Printf("   Cool to: %g degrees\n", Settings.Cooling.CoolTo)
	fmt.Printf("   Start tolerance: %g degrees\n", Settings.Cooling.CoolStartTol)
	fmt.Printf("   Wait maximum: %d minutes\n", Settings.Cooling.CoolWaitMinutes)
	fmt.Printf("   Abort if cooling outside tolerance: %t\n", Settings.Cooling.AbortOnCooling)
	fmt.Printf("   Abort tolerance: %g degrees\n", Settings.Cooling.CoolAbortTol)
	fmt.Printf("   Turn off cooler at end of session: %t\n", Settings.Cooling.OffAtEnd)

	//	Bias Frames
	fmt.Println("Bias Frames")
	for _, frameSetString := range viper.GetStringSlice(config.BiasFramesSetting) {
		count, binning, err := session.ParseBiasSet(frameSetString)
		if err != nil {
			fmt.Println("   Syntax error in set:", frameSetString)
		} else {
			fmt.Printf("   %d bias frames at %d x %d binning\n", count, binning, binning)
		}
	}

	//	Dark Frames
	fmt.Println("Dark Frames")
	for _, frameSetString := range viper.GetStringSlice(config.DarkFramesSetting) {
		count, exposure, binning, err := session.ParseDarkSet(frameSetString)
		if err != nil {
			fmt.Println("   Syntax error in set:", frameSetString)
		} else {
			fmt.Printf("   %d dark frames of %.2f seconds at %d x %d binning\n", count, exposure, binning, binning)
		}
	}
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
