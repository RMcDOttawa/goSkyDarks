package session

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"goskydarks/config"
	"goskydarks/delay"
	"goskydarks/theSkyX"
	"math"
	"time"
)

// Session struct implements the session service, used for overall session control
// such as start time or resuming from saved state
type Session struct {
	delayService     delay.DelayService //	Used to delay start; replace with mock for testing
	theSkyService    theSkyX.TheSkyService
	stateFileService StateFileService
	isConnected      bool
}

func NewSession() (*Session, error) {
	concreteDelayService := delay.NewDelayService()
	tsxService := theSkyX.NewTheSkyService(
		concreteDelayService,
	)
	stateFileService := NewStateFileService(viper.GetString(config.StateFileSetting), viper.GetFloat64(config.CoolToSetting))
	session := &Session{
		delayService:     concreteDelayService,
		theSkyService:    tsxService,
		stateFileService: stateFileService,
	}
	return session, nil
}

// CapturePlan is a struct that holds the plan for capturing frames, including what is needed
// and what has already been done.  It is used to track progress and to determine what
// needs to be done next.  It is saved to a file so that a session can be resumed if interrupted.
// Since capturing a frame includes waiting while it is downloaded, we also pre-measure the
// download time for each binning factor used so this can be taken into account in the delay waiting
// for the exposure to complete.  (TheSkyX doesn't provide a download complete notification)
// The download time is a linear function of the file size, which is a linear function of the binning factor,
// so we will just keep a measure for each binning level
type CapturePlan struct {
	DarksRequired []string
	BiasRequired  []string
	DarksDone     map[string]int
	BiasDone      map[string]int
	DownloadTimes map[int]float64 // seconds, indexed by binning
}

// SetDelayService allows delay service to be replaced with a mock for testing
func (s *Session) SetDelayService(delayService delay.DelayService) {
	s.delayService = delayService
}

// SetTheSkyService allows theSky service to be replaced with a mock for testing
func (s *Session) SetTheSkyService(theSkyService theSkyX.TheSkyService) {
	s.theSkyService = theSkyService
}

// SetStateFileService allows state file service to be replaced with a mock for testing
func (s *Session) SetStateFileService(theStateFileService StateFileService) {
	s.stateFileService = theStateFileService
}

// DelayStart optionally waits until a specified time before proceeding
// This can be used to initiate a session early in the day but have collection wait until
// later - perhaps when it is dark, or cooler
func (s *Session) DelayStart(startTime time.Time) error {
	fmt.Println("DelayStart to:", startTime)
	return s.delayService.DelayUntil(startTime)
}

// ConnectToServer opens the connection to the high-level communication service, keeping
// it open for subsequent use
func (s *Session) ConnectToServer() error {
	if s.isConnected {
		fmt.Println("Session already connected")
		return nil
	}

	if err := s.theSkyService.Connect(viper.GetString(config.ServerAddressSetting),
		viper.GetInt(config.ServerPortSetting)); err != nil {
		fmt.Println("Error in Session ConnectToServer:", err)
		return err
	}

	// TheSky sometimes returns nonsense as its first transaction.  e.g. sometimes the first temperature
	// read is -100, which is unlikely.  So we read and ignore the first temperature
	_, _ = s.theSkyService.GetCameraTemperature()
	//fmt.Println("Ignoring first temperature read:", ignoreTemp)

	s.isConnected = true
	return nil
}

// Close finishes the session, including closing the communication service if it is open
func (s *Session) Close() error {
	if !s.isConnected {
		fmt.Println("Session already disconnected")
		return nil
	}

	if err := s.theSkyService.Close(); err != nil {
		fmt.Println("Error in Session, closing theSky service:", err)
		return err
	}
	s.isConnected = false
	return nil
}

// CoolForStart turns on the camera cooler, if requested, and waits up to a maximum
// amount of time for the camera to reach the specified target temperature
func (s *Session) CoolForStart() error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	if verbosity > 2 || debug {
		fmt.Printf("Session/CoolForStart entered\n")
	}
	//	Is the session open?
	if !s.isConnected {
		return errors.New("session not connected")
	}
	//	See if we are being asked to cool the camera at all
	if !viper.GetBool(config.UseCoolerSetting) {
		if verbosity > 2 || debug {
			fmt.Println("UseCooling is not on, so nothing to do")
		}
		return nil
	}
	//	Cooling is requested.
	//	Start the cooler and set the target temperature

	coolTo := viper.GetFloat64(config.CoolToSetting)
	if err := s.theSkyService.StartCooling(coolTo); err != nil {
		fmt.Println("Error in Session CoolForStart, starting cooler:", err)
		return err
	}

	//	Wait until target temperature reached or time-out (too warm, can't cool that far)

	if err := s.waitForTargetTemperature(coolTo,
		viper.GetFloat64(config.CoolStartTolSetting),
		viper.GetInt(config.CoolWaitMinutesSetting)); err != nil {
		fmt.Println("Error in Session CoolForStart, waiting for cooler:", err)
		return err
	}
	if verbosity > 2 || debug {
		fmt.Printf("Session/CoolForStart exits\n")
	}
	return nil
}

//	 waitForTargetTemperature waits until the camera has cooled to the target temperature
//	 TheSkyX doesn't provide notifications, so we need to poll.  We will check the camera temperature
//		every minute.  We end the loop in one of two ways:
//		1. Success: the camera temperature is within the given tolerance of the target temperature
//		2. Failure: we have waited a specified maximum number of minutes and still haven't reached target
func (s *Session) waitForTargetTemperature(target float64, tolerance float64, maximumMinutes int) error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	if verbosity > 2 || debug {
		fmt.Printf("Session/waitForTargetTemperature entered, to %g tol %g max wait %d\n", target, tolerance, maximumMinutes)
	}
	secondsElapsed := 0
	maximumSeconds := maximumMinutes * 60
	coolStartPollSeconds := viper.GetInt(config.StartPollSecondsSetting)
	//	First temperature is sometimes nonsense, so read and ignore one
	_, _ = s.theSkyService.GetCameraTemperature()
	for {
		//fmt.Println("Seconds elapsed waiting for cooling:", secondsElapsed)
		if secondsElapsed > maximumSeconds {
			return errors.New("timed out waiting for target temperature")
		}
		currentTemperature, err := s.theSkyService.GetCameraTemperature()
		//fmt.Println("  Current temperature:", currentTemperature)
		if err != nil {
			fmt.Println("Error in Session WaitForTargetTemperature:", err)
			return err
		}
		if math.Abs(currentTemperature-target) <= tolerance {
			if verbosity > 2 {
				fmt.Printf("Current temperature %g is within tolerance %g of target %g\n", currentTemperature, tolerance, target)
				fmt.Println("waitForTargetTemperature exits")
			}
			return nil
		}
		if verbosity > 1 {
			fmt.Printf("  Current camera temperature is %.1f, target is %.1f, waiting %d seconds for cooling to stabilize.\n",
				currentTemperature, target, coolStartPollSeconds)
		}
		waitedSeconds, err := s.delayService.DelayDuration(coolStartPollSeconds)
		//fmt.Println("  Waited seconds:", waitedSeconds)
		secondsElapsed = secondsElapsed + waitedSeconds
	}
}

// CaptureFrames directs the server to capture the specified sets of bias and dark frames
// Progress recorded in a state file
// If, on entry, the state file indicates that a session was already underway and partially
// completed, that session is continued.  So the given bias and dark list represents the
// total set of frames wanted - not necessarily the captures that will be done on this call
func (s *Session) CaptureFrames(
	areDarksFirst bool,
	biasFrames []string,
	darkFrames []string) error {
	if viper.GetInt(config.VerbositySetting) > 2 || viper.GetBool(config.DebugSetting) {
		fmt.Println("Session/CaptureFrames entered")
		fmt.Println("  bias frames:", biasFrames)
		fmt.Println("  dark frames:", darkFrames)
	}

	//	Is the session open?
	if !s.isConnected {
		return errors.New("session not connected")
	}

	//  Get plan for captures needed, including state of what is already done
	capturePlan, err := s.getCapturePlan(biasFrames, darkFrames)
	if err != nil {
		fmt.Println("Error in Session CaptureFrames, getting capture plan:", err)
		return err
	}

	//	Ensure we have download time measurements for all the binning factors we will use
	if err := s.updateDownloadTimes(capturePlan); err != nil {
		fmt.Println("Error in Session CaptureFrames, updating download times")
		return err
	}

	//	Capture frames as needed
	if err := s.captureFrames(areDarksFirst, capturePlan); err != nil {
		fmt.Println("Error in Session capturing frames")
		return err
	}

	//  Update the saved plan one last time (has been updated during capture)
	if err := s.stateFileService.SavePlanToFile(capturePlan); err != nil {
		fmt.Println("Error in Session saving capture plan")
		return err
	}

	if viper.GetInt(config.VerbositySetting) > 2 || viper.GetBool(config.DebugSetting) {
		fmt.Println("Session/CaptureFrames exits")
	}
	return nil
}

func (s *Session) StopCooling() error {
	if viper.GetBool(config.UseCoolerSetting) && viper.GetBool(config.CoolerOffAtEndSetting) {
		if err := s.theSkyService.StopCooling(); err != nil {
			fmt.Println("Error in Session StopCooling:", err)
			return err
		}
		if viper.GetInt(config.VerbositySetting) > 1 {
			fmt.Printf("Cooling switched off at end of session")
		}
	}
	return nil
}

// getCapturePlan creates a plan for capturing the frames, based on the configuration and the state file
// We start with a plan based on the config file, then update it to reflect work already done as recorded in
//
//	the state file.
//	If the state file includes captures not in the current config, we ignore them - we are using only the "how many frames are done"
//	info from the state file, plus the download times for each binning level that may be recorded
func (s *Session) getCapturePlan(biasFrames []string, darkFrames []string) (*CapturePlan, error) {
	if viper.GetInt(config.VerbositySetting) > 2 || viper.GetBool(config.DebugSetting) {
		fmt.Println("Session/getCapturePlan entered")
		fmt.Println("  bias sets:", biasFrames)
		fmt.Println("  dark sets:", darkFrames)
	}
	capturePlan, err := s.createPlanFromConfig(biasFrames, darkFrames)
	if err != nil {
		fmt.Println("error in Session getCapturePlan, creating plan from config:", err)
		return nil, err
	}
	if viper.GetInt(config.VerbositySetting) > 2 || viper.GetBool(config.DebugSetting) {
		fmt.Println("Session/getCapturePlan got captureplan from createPlanFromConfig:", capturePlan)
	}
	err = s.stateFileService.UpdatePlanFromFile(capturePlan)
	if err != nil {
		fmt.Println("Error in Session getCapturePlan, updating plan from state file:", err)
		return nil, err
	}

	if viper.GetBool(config.ClearDoneSetting) {
		for k := range capturePlan.DarksDone {
			capturePlan.DarksDone[k] = 0
		}
		for k := range capturePlan.BiasDone {
			capturePlan.BiasDone[k] = 0
		}
	}
	if viper.GetInt(config.VerbositySetting) > 2 || viper.GetBool(config.DebugSetting) {
		fmt.Println("Session/getCapturePlan exits, returning:")
		fmt.Printf("  Capture plan: %#v\n", *capturePlan)
	}
	return capturePlan, nil
}

func (s *Session) createPlanFromConfig(biasSets []string, darkSets []string) (*CapturePlan, error) {
	verbosity := viper.GetInt(config.VerbositySetting)
	if verbosity > 3 || viper.GetBool(config.DebugSetting) {
		fmt.Println("createPlanFromConfig entered")
		fmt.Println("  bias sets:", biasSets)
		fmt.Println("  dark sets:", darkSets)
	}
	capturePlan := &CapturePlan{}

	capturePlan.DarksRequired = darkSets
	capturePlan.BiasRequired = biasSets

	//	Create the empty maps for what is done and download time
	capturePlan.DarksDone = make(map[string]int)
	capturePlan.BiasDone = make(map[string]int)
	capturePlan.DownloadTimes = make(map[int]float64)

	//	Create a DownloadTime entry and zero the "done" count for every dark set
	for _, darkSet := range darkSets {
		if verbosity > 2 {
			fmt.Printf("Session/createPlanFromConfig creating downloadtime and done entry for dark set: %s\n", darkSet)
		}
		count, exposure, binning, err := config.ParseDarkSet(darkSet)
		if err != nil {
			fmt.Printf("Error in Session createPlanFromConfig, parsing dark set %s: %s\n:", darkSet, err)
			return nil, err
		}
		key := MakeDarkKey(count, exposure, binning)
		capturePlan.DarksDone[key] = 0

		if _, ok := capturePlan.DownloadTimes[binning]; !ok {
			capturePlan.DownloadTimes[binning] = 0
		}
	}

	//	Create a DownloadTime entry and zero the "done" count for every bias dark set
	for _, biasSet := range biasSets {
		if verbosity > 2 {
			fmt.Printf("Session/createPlanFromConfig creating downloadtime and done entry for bias set: %s\n", biasSet)
		}
		count, binning, err := config.ParseBiasSet(biasSet)
		if err != nil {
			fmt.Printf("Error in Session createPlanFromConfig, parsing bias set %s: %s\n", biasSet, err)
			return nil, err
		}
		key := MakeBiasKey(count, binning)
		capturePlan.BiasDone[key] = 0

		if _, ok := capturePlan.DownloadTimes[binning]; !ok {
			capturePlan.DownloadTimes[binning] = 0
		}
	}
	if verbosity > 3 || viper.GetBool(config.DebugSetting) {
		fmt.Println("createPlanFromConfig exits")
		fmt.Printf("  capture plan: %#v\n", capturePlan)
	}

	return capturePlan, nil
}

func MakeDarkKey(count int, exposure float64, binning int) string {
	return fmt.Sprintf("Dark_%d_%.4f_%d", count, exposure, binning)
}

func MakeBiasKey(count int, binning int) string {
	return fmt.Sprintf("Bias_%d_%d", count, binning)
}

func (s *Session) updateDownloadTimes(capturePlan *CapturePlan) error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	if debug || verbosity > 2 {
		fmt.Printf("updateDownloadTimes. CapturePlan: %#v\n", *capturePlan)
	}
	for binning, seconds := range capturePlan.DownloadTimes {
		//fmt.Printf("  Binning %d, download time %g\n", binning, seconds)
		if seconds == 0 {
			if verbosity > 1 || debug {
				fmt.Printf("Measuring download time for binning %d\n", binning)
			}
			measuredTime, err := s.theSkyService.MeasureDownloadTime(binning)
			if err != nil {
				return errors.New("error measuring download time")
			}
			capturePlan.DownloadTimes[binning] = measuredTime
		}
	}
	if debug || verbosity > 3 {
		fmt.Printf("  updateDownloadTimes exits, Download times now %#v\n", capturePlan.DownloadTimes)
	}
	return nil
}

func (s *Session) captureFrames(careDarksFirst bool, capturePlan *CapturePlan) error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	if verbosity > 2 || debug {
		fmt.Printf("captureFrames. CapturePlan: %#v\n", *capturePlan)
	}

	//	We might be asked to do either the dark or bias frames first
	//	Determine which, then do a 2-pass loop so each gets done, in the desired order
	darksThisPass := careDarksFirst
	for i := 0; i < 2; i++ {
		if darksThisPass {
			if err := s.captureDarkFrames(capturePlan); err != nil {
				fmt.Println("Error in Session captureFrames, capturing dark frames:", err)
				return err
			}
		} else {
			if err := s.captureBiasFrames(capturePlan); err != nil {
				fmt.Println("Error in Session captureFrames, capturing dark frames:", err)
				return err
			}
		}
		darksThisPass = !darksThisPass
	}
	return nil
}

func (s *Session) captureDarkFrames(capturePlan *CapturePlan) error {
	if viper.GetInt(config.VerbositySetting) > 3 {
		fmt.Println("captureDarkFrames ")
		fmt.Printf("   Frames required: %v\n", capturePlan.DarksRequired)
		fmt.Printf("   Frames done: %v\n", capturePlan.DarksDone)
		fmt.Printf("   Download times: %v\n", capturePlan.DownloadTimes)
	}
	if viper.GetBool(config.NoDarkSetting) {
		if viper.GetInt(config.VerbositySetting) > 2 {
			fmt.Println("nodark flag, skipping dark frames")
		}
		return nil
	}
	for _, set := range capturePlan.DarksRequired {
		//fmt.Printf("   Checking dark set %s: %v\n", key, set)
		if err := s.captureDarkSet(capturePlan, set); err != nil {
			fmt.Println("Error in Session captureDarkFrames, capturing dark set:", err)
			return err
		}
	}
	return nil
}

func (s *Session) captureDarkSet(plan *CapturePlan, set string) error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	count, exposure, binning, err := config.ParseDarkSet(set)
	if verbosity > 0 || debug {
		fmt.Printf("Handling dark frames set: %d frames of %.1f seconds binned %d\n", count, exposure, binning)
	}
	if err != nil {
		fmt.Println("Error in Session captureDarkSet, parsing dark set:", err)
		return err
	}
	key := MakeDarkKey(count, exposure, binning)
	if plan.DarksDone[key] >= count {
		if verbosity > 1 {
			fmt.Printf("  Already have all %d dark frames in set %s\n", count, key)
		}
		return nil
	}

	framesNeeded := count - plan.DarksDone[key]
	if framesNeeded > 0 {
		if verbosity > 1 {
			fmt.Printf("  Still need %d dark frames (of %d) in set %s\n", framesNeeded, count, key)
		}
	}
	frameCount := 0
	for plan.DarksDone[key] < count {
		abandon, err := s.CheckAbandonForCooling()
		if err != nil {
			fmt.Println("Error in Session captureDarkSet, checking for cooling abandon:", err)
			return err
		}
		if abandon {
			const message = "abandoning dark frame capture due to temperature exceeding cooling tolerance"
			fmt.Println(message)
			return errors.New(message)
		}

		frameCount++
		if verbosity > 1 {
			fmt.Printf("    Capturing dark frame %d of %d:  %.2f seconds binned %d\n", frameCount, framesNeeded, exposure, binning)
		}

		if err := s.theSkyService.CaptureDarkFrame(binning, exposure, plan.DownloadTimes[binning]); err != nil {
			fmt.Println("Error in Session captureDarkSet, capturing dark frame:", err)
			return err
		}
		plan.DarksDone[key]++
		if err := s.stateFileService.SavePlanToFile(plan); err != nil {
			fmt.Println("Error in Session captureDarkSet, saving plan:", err)
			return err
		}
	}
	return nil
}

func (s *Session) captureBiasFrames(capturePlan *CapturePlan) error {
	//fmt.Println("captureBiasFrames ")
	//fmt.Printf("   Frames required: %v\n", capturePlan.BiasRequired)
	//fmt.Printf("   Frames done: %v\n", capturePlan.BiasDone)
	//fmt.Printf("   Download times: %v\n", capturePlan.DownloadTimes)
	//fmt.Printf("   Cooling config: %v\n", coolingConfig)

	if viper.GetBool(config.NoBiasSetting) {
		return nil
	}
	for _, set := range capturePlan.BiasRequired {
		//fmt.Printf("   Checking bias set %s: %v\n", key, set)
		if err := s.captureBiasSet(capturePlan, set); err != nil {
			fmt.Println("Error in Session captureBiasFrames, capturing bias set:", err)
			return err
		}
	}
	return nil
}

func (s *Session) captureBiasSet(plan *CapturePlan, set string) error {
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)
	count, binning, err := config.ParseBiasSet(set)
	if err != nil {
		fmt.Println("Error in Session captureBiasSet, parsing bias set:", err)
		return err
	}
	if verbosity > 0 || debug {
		fmt.Printf("Handling bias frames set: %d frames  binned %d\n", count, binning)
	}
	key := MakeBiasKey(count, binning)
	if plan.BiasDone[key] >= count {
		if verbosity > 1 {
			fmt.Printf("  Already have all %d bias frames in set %s\n", count, key)
		}
		return nil
	}

	framesNeeded := count - plan.BiasDone[key]
	if framesNeeded > 0 {
		if verbosity > 1 {
			fmt.Printf("  Still need %d bias frames (of %d) in set %s\n", framesNeeded, count, key)
		}
	}
	frameCount := 0
	for plan.BiasDone[key] < count {
		abandon, err := s.CheckAbandonForCooling()
		if err != nil {
			fmt.Println("Error in Session captureBiasSet, checking for cooling abandon:", err)
			return err
		}
		if abandon {
			const message = "abandoning bias frame capture due to temperature exceeding cooling tolerance"
			fmt.Println(message)
			return errors.New(message)
		}

		frameCount++
		if verbosity > 1 {
			fmt.Printf("    Capturing bias frame %d of %d, binned %d\n", frameCount, framesNeeded, binning)
		}

		if err := s.theSkyService.CaptureBiasFrame(binning, plan.DownloadTimes[binning]); err != nil {
			fmt.Println("Error in Session captureBiasSet, capturing bias frame:", err)
			return err
		}
		plan.BiasDone[key]++
		if err := s.stateFileService.SavePlanToFile(plan); err != nil {
			fmt.Println("Error in Session captureBiasSet, saving plan:", err)
			return err
		}
	}
	return nil
}

func (s *Session) CheckAbandonForCooling() (bool, error) {
	if viper.GetInt(config.VerbositySetting) > 2 {
		fmt.Println("CheckAbandonForCooling")
	}
	if !viper.GetBool(config.UseCoolerSetting) {
		return false, nil
	}
	if !viper.GetBool(config.AbortOnCoolingSetting) {
		return false, nil
	}
	cameraTemperature, err := s.theSkyService.GetCameraTemperature()
	if viper.GetInt(config.VerbositySetting) > 2 {
		fmt.Println("  Camera temperature:", cameraTemperature)
	}
	if err != nil {
		fmt.Println("Error in Session CheckAbandonForCooling, getting camera temperature:", err)
		return false, err
	}
	variation := math.Abs(cameraTemperature - viper.GetFloat64(config.CoolToSetting))
	//fmt.Printf("  Temp %g and target %g = variation %g\n", cameraTemperature, coolingConfig.CoolTo, variation)
	if variation >= viper.GetFloat64(config.CoolAbortTolSetting) {
		// Camera temperature is unacceptable - return an abort request
		return true, nil
	}
	return false, nil
}
