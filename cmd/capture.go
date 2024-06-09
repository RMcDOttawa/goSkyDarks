/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

//var captureConfig config.CaptureSpecs

//var localServerAddress string
//var localServerPort int
//var localUseCooler bool
//var localCoolTo float32
//var localCoolStartTol float32
//var localCoolWaitMinutes int
//var localCoolAbort bool
//var localCoolAbortTol float32
//var localStartAtString string
//var localBiasStrings *[]string
//var localDarkStrings *[]string

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
		//captureConfig.SetServerAddress(localServerAddress)
		//captureConfig.SetServerPort(localServerPort)
		//captureConfig.SetUseCooler(localUseCooler)
		//captureConfig.SetCoolTo(localCoolTo)
		//captureConfig.SetCoolStartTol(localCoolStartTol)
		//captureConfig.SetCoolWaitMinutes(localCoolWaitMinutes)
		//captureConfig.SetCoolAbort(localCoolAbort)
		//captureConfig.SetCoolAbortTol(localCoolAbortTol)
		//captureConfig.SetStartAtString(localStartAtString)
		//captureConfig.SetBiasStrings(*localBiasStrings)
		//captureConfig.SetDarkStrings(*localDarkStrings)

		//ValidateSettings(captureConfig)

		//if Debug || Verbosity >= 1 {
		//	fmt.Println("Capture command entered")
		//	fmt.Println("Capture command entered")
		//	if Debug || Verbosity >= 4 {
		//		DisplayFlags()
		//	}
		//}

		//	Create the capture session
		//session, err := session.NewSession()
		//if err != nil {
		//	_, _ = fmt.Fprintln(os.Stderr, err)
		//	return
		//}

		//	get ready for capture by waiting for the start time (if requested)
		//	and cooling the camera (if requested)
		//err = session.PrepareForCapture(captureConfig)
		//if err != nil {
		//	_, _ = fmt.Fprintln(os.Stderr, err)
		//	return
		//}

		//	Do the captures until finished or aborted

	},
}

//func ValidateSettings(config config.CaptureSpecs) {
//	fmt.Println("ValidateSettings")
//	fmt.Printf("captureConfig: %#v\n", config)
//	err := validateConfig(config)
//	if err != nil {
//		fmt.Println("Error in setting:", err)
//		os.Exit(1)
//	}
//
//	delayStart, startTime, err := config.ParseStartTime(config.GetStartAtString())
//	if err != nil {
//		fmt.Println("Error in setting:", err)
//		os.Exit(1)
//	}
//	config.SetDelayStart(delayStart)
//	config.SetStartTime(startTime)
//
//	biasFrames, err := config.ParseBiasStrings(config.GetBiasStrings())
//	if err != nil {
//		fmt.Println("Error in specified bias string:", err)
//		os.Exit(1)
//	}
//	config.SetBiasFrames(biasFrames)
//
//	darkFrames, err := config.ParseDarkStrings(config.GetDarkStrings())
//	if err != nil {
//		fmt.Println("Error in specified dark string:", err)
//		os.Exit(1)
//	}
//	config.SetDarkFrames(darkFrames)
//
//	if len(config.GetBiasFrames()) == 0 && len(config.GetDarkFrames()) == 0 {
//		fmt.Println("At least one bias or dark frame set must be specified")
//		os.Exit(1)
//	}
//}

//func validateConfig(config config.CaptureSpecs) error {
//	//	Server address must be a valid IP address or domain name
//	err := validateServer(config.GetServerAddress())
//	if err != nil {
//		return err
//	}
//
//	//	Port number in range
//	port := config.GetServerPort("port")
//	if port < 0 || port > 65535 {
//		return errors.New("port must be between 0 and 65535")
//	}
//
//	//	Cool To value in range -200 to +200
//	coolTo := config.GetCoolTo()
//	const minCoolTo = -200.0
//	const maxCoolTo = 200.0
//	if coolTo < minCoolTo || coolTo > maxCoolTo {
//		return errors.New(fmt.Sprintf("coolto must be between %g and %g (degrees)", minCoolTo, maxCoolTo))
//	}
//
//	//	Cool Start Tolerance value in range 0 to 50
//	coolToStartTol := config.GetCoolStartTol()
//	const minStartTol = 0.0
//	const maxStartTol = 50.0
//	if coolToStartTol < minStartTol || coolToStartTol > maxStartTol {
//		return errors.New(fmt.Sprintf("coolstarttol must be between %g and %g (degrees)", minStartTol, maxStartTol))
//	}
//
//	//	Cool Abort Tolerance value in range 0 to 50
//	coolAbortTol := config.GetCoolAbortTol()
//	const minAbortTol = 0.0
//	const maxAbortTol = 50.0
//	if coolAbortTol < minAbortTol || coolAbortTol > maxAbortTol {
//		return errors.New(fmt.Sprintf("coolaborttol must be between %g and %g (degrees)", minAbortTol, maxAbortTol))
//	}
//
//	//	Cooling wait minutes in range 1 to (24 hours)
//	coolingWaitMinutes := config.GetCoolWaitMinutes()
//	const minCoolWait = 1
//	const maxCoolWait = 24 * 60
//	if coolingWaitMinutes < minCoolWait || coolingWaitMinutes > maxCoolWait {
//		return errors.New(fmt.Sprintf("coolwait must be between %d and %d (minutes)", minCoolWait, maxCoolWait))
//	}
//
//	return nil
//}

//func validateServer(addressString string) error {
//	fmt.Println("validateServer:", addressString)
//	tryIp := net.ParseIP(addressString)
//	if tryIp != nil {
//		//	Successful parse, so it's a valid IP.  Return nil for no error
//		return nil
//	}
//	//	Not a valid IP.  See if it's a (syntactically) valid domain name.
//	addressString = strings.ToLower(addressString)
//	if addressString == "localhost" {
//		return nil
//	}
//	//	We won't try to actually connect - leave that for the capture process
//	if validator.IsValidDomain(addressString) {
//		return nil
//	}
//	return errors.New(fmt.Sprintf("Invalid server address: %s", addressString))
//}

func init() {
	rootCmd.AddCommand(captureCmd)

	//captureCmd.PersistentFlags().StringVarP(&localServerAddress, "server", "s", "localhost", "Address of TheSkyX server")
	//err := viper.BindPFlag("server", captureCmd.PersistentFlags().Lookup("server"))
	//if err != nil {
	//	fmt.Println("Error binding server flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().IntVarP(&localServerPort, "port", "p", 3040, "Port number of TheSkyX server")
	//err = viper.BindPFlag("port", captureCmd.PersistentFlags().Lookup("port"))
	//if err != nil {
	//	fmt.Println("Error binding port flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().BoolVarP(&localUseCooler, "cool", "", false, "Use camera cooler")
	//err = viper.BindPFlag("cool", captureCmd.PersistentFlags().Lookup("cool"))
	//if err != nil {
	//	fmt.Println("Error binding cool flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().Float32VarP(&localCoolTo, "coolto", "", -10.0, "Cool to target temperature")
	//err = viper.BindPFlag("coolto", captureCmd.PersistentFlags().Lookup("coolto"))
	//if err != nil {
	//	fmt.Println("Error binding coolto flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().Float32VarP(&localCoolStartTol, "coolstarttol", "", 1.0, "Cooling start tolerance")
	//err = viper.BindPFlag("coolstarttol", captureCmd.PersistentFlags().Lookup("coolstarttol"))
	//if err != nil {
	//	fmt.Println("Error binding coolstarttol flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().IntVarP(&localCoolWaitMinutes, "coolwait", "", 30, "Maximum minutes to reach temperature")
	//err = viper.BindPFlag("coolwait", captureCmd.PersistentFlags().Lookup("coolwait"))
	//if err != nil {
	//	fmt.Println("Error binding coolwait flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().BoolVarP(&localCoolAbort, "coolabort", "", false, "Abort capture if cooling leaves range")
	//err = viper.BindPFlag("coolabort", captureCmd.PersistentFlags().Lookup("coolabort"))
	//if err != nil {
	//	fmt.Println("Error binding coolabort flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().Float32VarP(&localCoolAbortTol, "coolaborttol", "", 3.0, "Cooling capture abort tolerance")
	//err = viper.BindPFlag("coolaborttol", captureCmd.PersistentFlags().Lookup("coolaborttol"))
	//if err != nil {
	//	fmt.Println("Error binding coolaborttol flag:", err)
	//}
	//
	//captureCmd.PersistentFlags().StringVarP(&localStartAtString, "startat", "", "", "Time to start capture")
	//err = viper.BindPFlag("startat", captureCmd.PersistentFlags().Lookup("startat"))
	//if err != nil {
	//	fmt.Println("Error binding startat flag:", err)
	//}
	//
	//localBiasStrings = captureCmd.PersistentFlags().StringArray("bias", []string{}, "Bias frame (count,binning) - can repeat multiple times")
	//err = viper.BindPFlag("bias", captureCmd.PersistentFlags().Lookup("bias"))
	//if err != nil {
	//	fmt.Println("Error binding bias flag:", err)
	//}
	//
	//localDarkStrings = captureCmd.PersistentFlags().StringArray("dark", []string{}, "Dark frame (count,seconds,binning) - can repeat multiple times")
	//err = viper.BindPFlag("dark", captureCmd.PersistentFlags().Lookup("dark"))
	//if err != nil {
	//	fmt.Println("Error binding dark flag:", err)
	//}

}
