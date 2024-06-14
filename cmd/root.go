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
//verbosity := viper.GetInt(VerbositySetting)
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
	fmt.Println("Root Init")
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	Settings = &config.SettingsType{}

	defineGlobalSettings()
	defineCaptureSettings()
	readConfigFile()

}

func defineGlobalSettings() {
	rootCmd.PersistentFlags().BoolVarP(&Settings.Debug, "debug", "", false, "Display debugging output in the console. (default: false)")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().IntVarP(&Settings.Verbosity, "verbosity", "v", 1, "Set the verbosity level from 0 to 5. (default: 0)")
	_ = viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))

	rootCmd.PersistentFlags().StringVarP(&Settings.StateFile, "statefile", "", "./stateFile.state", "State file to store session status")
	_ = viper.BindPFlag("statefile", rootCmd.PersistentFlags().Lookup("statefile"))

	rootCmd.PersistentFlags().BoolVarP(&Settings.ShowSettings, "showsettings", "", false, "show settings")
	_ = viper.BindPFlag("showsettings", rootCmd.PersistentFlags().Lookup("showsettings"))

}

func readConfigFile() {
	//	Read config settings from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//	No config file is not an error. Just mention it and then leave
			fmt.Println("No config.yml file found")
			return
		}
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

	if err := config.ValidateGlobals(); err != nil {
		fmt.Println("Error validating global settings:", err)
		os.Exit(1)
	}
}

func defineCaptureSettings() {
	captureCmd := findCommand(rootCmd, "capture")

	defineServerFlags(captureCmd)
	defineStartDelayFlags(captureCmd)
	defineCoolingFlags(captureCmd)
	defineFramesFlags(captureCmd)

}

func defineFramesFlags(_ *cobra.Command) {

	captureCmd.Flags().StringArrayVarP(&Settings.BiasFrames, "bias", "b", []string{}, "Bias frame \"count,binning\" - can repeat multiple times")
	_ = viper.BindPFlag(config.BiasFramesSetting, captureCmd.Flags().Lookup("bias"))

	captureCmd.Flags().StringArrayVarP(&Settings.DarkFrames, "dark", "d", []string{}, "Dark frame \"count,seconds,binning\" - can repeat multiple times")
	_ = viper.BindPFlag(config.DarkFramesSetting, captureCmd.Flags().Lookup("dark"))

}

func defineServerFlags(captureCmd *cobra.Command) {

	captureCmd.Flags().StringVarP(&Settings.Server.Address, "server", "", "localhost", "Server address")
	_ = viper.BindPFlag(config.ServerAddressSetting, captureCmd.Flags().Lookup("server"))

	captureCmd.Flags().IntVarP(&Settings.Server.Port, "port", "", 3040, "Server port number")
	_ = viper.BindPFlag(config.ServerPortSetting, captureCmd.Flags().Lookup("port"))

}

func defineStartDelayFlags(captureCmd *cobra.Command) {

	captureCmd.Flags().BoolVarP(&Settings.Start.Delay, "delaystart", "", false, "Delay start until later")
	_ = viper.BindPFlag(config.StartDelaySetting, captureCmd.Flags().Lookup("delaystart"))

	captureCmd.Flags().StringVarP(&Settings.Start.Day, "startday", "", "today", "Delay start until what day (today, tomorrow, or yyyy-mm-dd)")
	_ = viper.BindPFlag(config.StartDaySetting, captureCmd.Flags().Lookup("startday"))

	captureCmd.Flags().StringVarP(&Settings.Start.Time, "starttime", "", "", "Delay start until what time (\"HH:MM\" 24-hour format)")
	_ = viper.BindPFlag(config.StartTimeSetting, captureCmd.Flags().Lookup("starttime"))

}

func defineCoolingFlags(captureCmd *cobra.Command) {

	captureCmd.Flags().BoolVarP(&Settings.Cooling.UseCooler, "usecooler", "", false, "Use camera cooler")
	_ = viper.BindPFlag(config.UseCoolerSetting, captureCmd.Flags().Lookup("usecooler"))

	captureCmd.Flags().Float64VarP(&Settings.Cooling.CoolTo, "coolto", "t", 0.0, "Camera target temperature")
	_ = viper.BindPFlag(config.CoolToSetting, captureCmd.Flags().Lookup("coolto"))

	captureCmd.Flags().Float64VarP(&Settings.Cooling.CoolStartTol, "coolstarttol", "", 2.0, "Cooling start tolerance")
	_ = viper.BindPFlag(config.CoolStartTolSetting, captureCmd.Flags().Lookup("coolstarttol"))

	captureCmd.Flags().IntVarP(&Settings.Cooling.CoolWaitMinutes, "coolwaitminutes", "", 30, "Cooling maximum wait time")
	_ = viper.BindPFlag(config.CoolWaitMinutesSetting, captureCmd.Flags().Lookup("coolwaitminutes"))

	captureCmd.Flags().BoolVarP(&Settings.Cooling.AbortOnCooling, "abortoncooling", "", false, "Abort capture if cooling is outside tolerance")
	_ = viper.BindPFlag(config.AbortOnCoolingSetting, captureCmd.Flags().Lookup("abortoncooling"))

	captureCmd.Flags().Float64VarP(&Settings.Cooling.CoolAbortTol, "coolaborttol", "", 2.0, "Cooling abort tolerance")
	_ = viper.BindPFlag(config.CoolAbortTolSetting, captureCmd.Flags().Lookup("coolaborttol"))

	captureCmd.Flags().BoolVarP(&Settings.Cooling.OffAtEnd, "coolingoffafter", "", false, "Cooling off after capture complete")
	_ = viper.BindPFlag(config.CoolerOffAtEndSetting, captureCmd.Flags().Lookup("coolingoffafter"))
}

func findCommand(rootCmd *cobra.Command, name string) *cobra.Command {
	commands := rootCmd.Commands()
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
