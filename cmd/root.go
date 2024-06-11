/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"goskydarks/config"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Settings *config.SettingsType

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goskydarks",
	Short: "Collect dark frames using TheSkyX for camera control",
	Long: `Connect with the TCP server running inside the program TheSkyX, 
somewhere on your local network, and orchestrate it taking a set of 
dark or bias frames with a variety of specifications.  
This enables the automation of the process of collecting these 
calibration frames.

Note that TheSkyX doesn't offer any way to receive the collected frames over the
network. They will be saved on the computer where TheSkyX is running, in the location
specified in the "autosave location" in the application.  You can configure theSky to save
to a network drive if you like, but that is up to you. This program only causes the images
to be captured, it does not deal with where they are stored.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running root-level command")
		//if Debug || Verbosity >= 4 {
		//	DisplayFlags()
		//}
		//err := validateGlobalConfig()
		//if err != nil {
		//	fmt.Println("Error in global flag:", err)
		//	os.Exit(1)
		//}
	},
}

//func validateGlobalConfig() error {
//	Verbosity: integer from 0 to 5
//verbosity := viper.GetInt("verbosity")
//if verbosity < 0 || verbosity > 5 {
//	return errors.New(fmt.Sprintf("%d is an invalid verbosity level (must be 0 to 5)", verbosity))
//}

//	State file: no validation - will depend on what we do with it and we'll
//	detect any errors in path then

//return nil
//}

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

	Settings = &config.SettingsType{}

	//	Read config settings from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}
	if err := viper.UnmarshalExact(Settings); err != nil {
		fmt.Println("Unmarshal err:", err)
		os.Exit(1)
	}

	if Settings.Debug || Settings.Verbosity > 2 {
		fmt.Printf("Read configuration from file: %s\n", viper.ConfigFileUsed())
	}

	if err := Settings.ValidateGlobals(); err != nil {
		fmt.Println("Error validating global settings:", err)
		os.Exit(1)
	}

}

//func DisplayFlags() {
//	fmt.Println("All program config settings:")
//	for k, v := range viper.AllSettings() {
//		fmt.Printf("   %s: %v\n", k, v)
//	}
//}
