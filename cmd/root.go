/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CfgFilePath string
var StateFilePath string

// Verbosity is a number from 0 to 5 indicating how chatty the program is
// 0 = no messages except errors and essential,
// all the way to 5 = lots and lots of messages
var Verbosity int

var Debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goskydarks",
	Short: "Collect dark frames using TheSkyX for camera control",
	Long: `Connect with the TCP server running inside the program TheSkyX, 
somewhere on your local network, and orchestrate it taking a set of 
dark or bias frames with a variety of specifications.  
This enables the automation of the process of collecting these 
calibration frames.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running root-level command")
		if Debug || Verbosity >= 4 {
			DisplayFlags()
		}
		err := validateGlobalConfig()
		if err != nil {
			fmt.Println("Error in global flag:", err)
			os.Exit(1)
		}
	},
}

func validateGlobalConfig() error {
	//	Verbosity: integer from 0 to 5
	verbosity := viper.GetInt("verbosity")
	if verbosity < 0 || verbosity > 5 {
		return errors.New(fmt.Sprintf("%d is an invalid verbosity level (must be 0 to 5)", verbosity))
	}

	//	Config file: empty string or path must exist
	//	Config file not implemented yet
	//configPath := viper.GetString("config")
	//if configPath != "" {
	//	_, err := os.Stat(configPath)
	//	if err != nil {
	//		return errors.New(fmt.Sprintf("config file \"%s\" does not exist", configPath))
	//	}
	//}

	//	State file: no validation - will depend on what we do with it and we'll
	//	detect any errors in path then

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	//Config file not implemented yet - for now, just invoke from command line and put invocation
	//	in a shell script to avoid the constant retyping
	//rootCmd.PersistentFlags().StringVarP(&CfgFilePath, "config", "c", "", "config yaml file")
	//err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	//if err != nil {
	//	fmt.Println("Error binding config flag:", err)
	//}

	rootCmd.PersistentFlags().StringVarP(&StateFilePath, "statefile", "f", "", "")
	err := viper.BindPFlag("statefile", rootCmd.PersistentFlags().Lookup("statefile"))
	if err != nil {
		fmt.Println("Error binding statefile flag:", err)
	}

	rootCmd.PersistentFlags().IntVarP(&Verbosity, "verbosity", "v", 1, "Number of messages. 0 (none) to 5 (lots)")
	err = viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	if err != nil {
		fmt.Println("Error binding verbosity flag:", err)
	}

	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
	err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		fmt.Println("Error binding debug flag:", err)
	}

}

func DisplayFlags() {
	fmt.Println("All program config settings:")
	for k, v := range viper.AllSettings() {
		fmt.Printf("   %s: %v\n", k, v)
	}
}
