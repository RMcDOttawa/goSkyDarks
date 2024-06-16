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
	"os"
)

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Use TheSkyX to capture calibration frames",
	Long: `Uses TheSkyX to capture dark and bias frames as specified in the configuration file.  
If the state file indicates that a previous run was terminated but unfinished, capture will pick up from where the previous run left off.  
Use the RESET command to prevent this and start over.

Note the config file allows the capture to be deferred until later - e.g. after dark when it is cooler.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool(config.ShowSettingsSetting) {
			config.ShowAllSettings()
		}
		//	State file is mandatory when doing a capture
		if viper.GetString(config.StateFileSetting) == "" {
			_, _ = fmt.Fprintln(os.Stderr, "State file is required for capture")
			return
		}

		consistentizeCooling(cmd)

		//	Get bias and dark frame specs
		biasFrames := viper.GetStringSlice(config.BiasFramesSetting)
		darkFrames := viper.GetStringSlice(config.DarkFramesSetting)
		if err := validateBiasFrames(biasFrames); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if err := validateDarkFrames(darkFrames); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if len(biasFrames) == 0 && len(darkFrames) == 0 {
			fmt.Println("Nothing to capture - specify bias or dark frames")
			return
		}

		//	Create the capture session
		session, err := session.NewSession()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		defer func() {
			//fmt.Println("Closing Session")
			_ = session.Close()
		}()

		//	Delay start
		delay, targetTime, err := config.ParseStart()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if delay {
			err = session.DelayStart(targetTime)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return
			}
		}

		//	Establish server connection
		err = session.ConnectToServer()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	Cool the camera
		err = session.CoolForStart()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	Do the captures until done, interrupted, or cooling aborts
		err = session.CaptureFrames(areDarksFirst(cmd), biasFrames, darkFrames)

		//	Stop cooling
		err = session.StopCooling()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

	},
}

//	User may use the --coolto flag thinking that is sufficient to turn on cooling
//	(it isn't - also need the useCooling flag).  If --coolto flag is explicitly used
//	then we'll set --useCooling on.  We'll warn them if this was a change.

func consistentizeCooling(cmd *cobra.Command) {
	if config.FlagExplicitlySet(cmd, "coolto") {
		if !viper.GetBool(config.UseCoolerSetting) {
			fmt.Println("--coolto used without --usecooler. Turning --usecooler on too.")
			viper.Set(config.UseCoolerSetting, true)
		}
	}
}

func validateDarkFrames(frameStrings []string) error {
	for _, frameString := range frameStrings {
		_, _, _, err := config.ParseDarkSet(frameString)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateBiasFrames(frameStrings []string) error {
	for _, frameString := range frameStrings {
		_, _, err := config.ParseBiasSet(frameString)
		if err != nil {
			return err
		}
	}
	return nil
}

// func
func init() {
	rootCmd.AddCommand(captureCmd)

}

// Determine which set of frames to do first.  We return this result by returning a boolean
// - true if darks are to be done first
// - false if bias frames are to be done first
// If one of (bias, dark) is explicitly set in the cli, use that (darks if both are specified)
// Otherwise, use the setting in the config file
func areDarksFirst(cmd *cobra.Command) bool {
	if config.FlagExplicitlySet(cmd, "darkfirst") {
		return true
	}
	if config.FlagExplicitlySet(cmd, "biasfirst") {
		return false
	}
	if viper.GetBool(config.DarkFirstSetting) {
		return true
	}
	if viper.GetBool(config.BiasFirstSetting) {
		return false
	}
	return true
}
