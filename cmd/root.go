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
calibration frames.`,
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

	if Settings.Debug || Settings.Verbosity > 1 {
		fmt.Printf("Read configuration from file: %s\n", viper.ConfigFileUsed())
	}

	if err := Settings.ValidateGlobals(); err != nil {
		fmt.Println("Error validating global settings:", err)
		os.Exit(1)
	}

	if Settings.Debug && len(Settings.BiasFrames) > 0 {
		//fmt.Printf("\nBias Frames: %#v\n\n", config.Settings.BiasFrames)
		biasList := Settings.GetBiasSets()
		for i, bias := range biasList {
			fmt.Printf("Bias set %d: %d frames binned at %d\n", i, bias.Frames, bias.Binning)
		}
	}

	if Settings.Debug && len(Settings.DarkFrames) > 0 {
		//fmt.Printf("\nDark Frames: %#v\n\n", config.Settings.DarkFrames)
		darkList := Settings.GetDarkSets()
		for i, dark := range darkList {
			fmt.Printf("Dark set %d: %d frames of %d seconds binned at %d\n",
				i, dark.Frames, dark.Seconds, dark.Binning)
		}
	}

	//if len(config.Settings.DarkFrames) > 0 {
	//	fmt.Printf("\nDark Frames: %#v\n\n", config.Settings.DarkFrames)
	//	for i := 0; i < len(config.Settings.DarkFrames); i++ {
	//		fmt.Printf("Dark Set %d %#v\n", i, config.Settings.DarkFrames[i])
	//	}
	//}

	//rootCmd.PersistentFlags().StringVarP(&StateFilePath, "statefile", "f", "", "")
	//err := viper.BindPFlag("statefile", rootCmd.PersistentFlags().Lookup("statefile"))
	//if err != nil {
	//	fmt.Println("Error binding statefile flag:", err)
	//}
	//
	//rootCmd.PersistentFlags().IntVarP(&Verbosity, "verbosity", "v", 1, "Number of messages. 0 (none) to 5 (lots)")
	//err = viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))
	//if err != nil {
	//	fmt.Println("Error binding verbosity flag:", err)
	//}
	//
	//rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
	//err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	//if err != nil {
	//	fmt.Println("Error binding debug flag:", err)
	//}

}

//func DisplayFlags() {
//	fmt.Println("All program config settings:")
//	for k, v := range viper.AllSettings() {
//		fmt.Printf("   %s: %v\n", k, v)
//	}
//}
