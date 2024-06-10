/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
		//	Get bias and dark frame specs
		biasSets := Settings.GetBiasSets()
		darkSets := Settings.GetDarkSets()
		if len(biasSets) == 0 && len(darkSets) == 0 {
			fmt.Println("Nothing to capture - specify bias or dark frames")
			return
		}

		//	Create the capture session
		session, err := session.NewSession(*Settings)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		defer func() {
			//fmt.Println("Closing Session")
			_ = session.Close()
		}()

		//	Delay start
		delay, targetTime, err := Settings.ParseStart()
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
		err = session.ConnectToServer(Settings.Server)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	Cool the camera
		err = session.CoolForStart(Settings.Cooling)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

		//	Do the captures until done, interrupted, or cooling aborts
		err = session.CaptureFrames(biasSets, darkSets)

		//	Stop cooling
		err = session.StopCooling(Settings.Cooling)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}

	},
}

// func
func init() {
	rootCmd.AddCommand(captureCmd)

}
