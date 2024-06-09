/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
	delay, start, err := Settings.ParseStart()
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

	//	Bias Frames
	fmt.Println("Bias Frames")
	bf := Settings.GetBiasSets()
	if len(bf) == 0 {
		fmt.Println("   No bias sets")
	} else {
		for _, bias := range bf {
			fmt.Printf("    %d frames binned at %d\n", bias.Frames, bias.Binning)
		}
	}

	//	Dark Frames
	fmt.Println("Dark Frames")
	df := Settings.GetDarkSets()
	if len(df) == 0 {
		fmt.Println("   No dark sets")
	} else {
		for _, dark := range df {
			fmt.Printf("    %d frames of %d seconds binned at %d\n",
				dark.Frames, dark.Seconds, dark.Binning)
		}
	}
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
