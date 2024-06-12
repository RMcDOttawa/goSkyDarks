package session

import (
	"errors"
	"fmt"
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
	settings         config.SettingsType
	theSkyService    theSkyX.TheSkyService
	stateFileService StateFileService
	isConnected      bool
}

func NewSession(settings config.SettingsType) (*Session, error) {
	concreteDelayService := delay.NewDelayService(settings)
	tsxService := theSkyX.NewTheSkyService(
		settings,
		concreteDelayService,
	)
	stateFileService := NewStateFileService(settings.StateFile)
	session := &Session{
		delayService:     concreteDelayService,
		settings:         settings,
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
	DarksRequired map[string]config.DarkSet
	BiasRequired  map[string]config.BiasSet
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

const pollingDelaySeconds = 120

// DelayStart optionally waits until a specified time before proceeding
// This can be used to initiate a session early in the day but have collection wait until
// later - perhaps when it is dark, or cooler
func (s *Session) DelayStart(startTime time.Time) error {
	fmt.Println("DelayStart to:", startTime)
	return s.delayService.DelayUntil(startTime)
}

// ConnectToServer opens the connection to the high-level communication service, keeping
// it open for subsequent use
func (s *Session) ConnectToServer(server config.ServerConfig) error {
	if s.isConnected {
		fmt.Println("Session already connected")
		return nil
	}

	if err := s.theSkyService.Connect(server.Address, server.Port); err != nil {
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
func (s *Session) CoolForStart(cooling config.CoolingConfig) error {
	if s.settings.Verbosity > 2 || s.settings.Debug {
		fmt.Printf("Session/CoolForStart entered: %#v\n", cooling)
	}
	//	Is the session open?
	if !s.isConnected {
		return errors.New("session not connected")
	}
	//	See if we are being asked to cool the camera at all
	if !cooling.UseCooler {
		if s.settings.Verbosity > 2 || s.settings.Debug {
			fmt.Println("UseCooling is not on, so nothing to do")
			return nil
		}
	}
	//	Cooling is requested.
	//	Start the cooler and set the target temperature

	if err := s.theSkyService.StartCooling(cooling.CoolTo); err != nil {
		fmt.Println("Error in Session CoolForStart, starting cooler:", err)
		return err
	}

	//	Wait until target temperature reached or time-out (too warm, can't cool that far)

	if err := s.waitForTargetTemperature(cooling.CoolTo, cooling.CoolStartTol, cooling.CoolWaitMinutes); err != nil {
		fmt.Println("Error in Session CoolForStart, waiting for cooler:", err)
		return err
	}
	return nil
}

//	 waitForTargetTemperature waits until the camera has cooled to the target temperature
//	 TheSkyX doesn't provide notifications, so we need to poll.  We will check the camera temperature
//		every minute.  We end the loop in one of two ways:
//		1. Success: the camera temperature is within the given tolerance of the target temperature
//		2. Failure: we have waited a specified maximum number of minutes and still haven't reached target
func (s *Session) waitForTargetTemperature(target float64, tolerance float64, maximumMinutes int) error {
	if s.settings.Verbosity > 2 || s.settings.Debug {
		fmt.Printf("Session/waitForTargetTemperature entered, to %g tol %g max wait %d\n", target, tolerance, maximumMinutes)
	}
	secondsElapsed := 0
	maximumSeconds := maximumMinutes * 60
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
			if s.settings.Verbosity > 2 || s.settings.Debug {
				fmt.Printf("Current temperature %g is within tolerance %g of target %g\n", currentTemperature, tolerance, target)
			}
			return nil
		}
		waitedSeconds, err := s.delayService.DelayDuration(pollingDelaySeconds)
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
	biasSets []config.BiasSet,
	darkSets []config.DarkSet,
	coolingConfig config.CoolingConfig) error {

	//	Is the session open?
	if !s.isConnected {
		return errors.New("session not connected")
	}

	//  Get plan for captures needed, including state of what is already done
	capturePlan, err := s.getCapturePlan(biasSets, darkSets)
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
	if err := s.captureFrames(capturePlan, coolingConfig); err != nil {
		fmt.Println("Error in Session capturing frames")
		return err
	}

	//  Update the saved plan one last time (has been updated during capture)
	if err := s.stateFileService.SavePlanToFile(capturePlan); err != nil {
		fmt.Println("Error in Session saving capture plan")
		return err
	}

	return nil
}

func (s *Session) StopCooling(cooling config.CoolingConfig) error {
	if cooling.UseCooler && cooling.OffAtEnd {
		if err := s.theSkyService.StopCooling(); err != nil {
			fmt.Println("Error in Session StopCooling:", err)
			return err
		}
		if s.settings.Verbosity > 1 || s.settings.Debug {
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
func (s *Session) getCapturePlan(biasSets []config.BiasSet, darkSets []config.DarkSet) (*CapturePlan, error) {
	//fmt.Println("getCapturePlan ")
	//fmt.Println("  bias sets:", biasSets)
	//fmt.Println("  dark sets:", darkSets)
	//fmt.Println("  state file:", stateFilePath)
	capturePlan := s.createPlanFromConfig(biasSets, darkSets)
	err := s.stateFileService.UpdatePlanFromFile(capturePlan)
	if err != nil {
		fmt.Println("Error in Session getCapturePlan, updating plan from state file:", err)
		return nil, err
	}
	return capturePlan, nil
}

func (s *Session) createPlanFromConfig(biasSets []config.BiasSet, darkSets []config.DarkSet) *CapturePlan {
	//fmt.Println("createPlanFromConfig STUB")
	//fmt.Println("  bias sets:", biasSets)
	//fmt.Println("  dark sets:", darkSets)
	capturePlan := CapturePlan{}
	//	Create the empty maps
	capturePlan.DarksRequired = make(map[string]config.DarkSet)
	capturePlan.BiasRequired = make(map[string]config.BiasSet)
	capturePlan.DarksDone = make(map[string]int)
	capturePlan.BiasDone = make(map[string]int)
	capturePlan.DownloadTimes = make(map[int]float64)

	//	Initialize the settings for every dark frame set needed
	for _, darkSet := range darkSets {
		key := MakeDarkKey(darkSet)
		capturePlan.DarksRequired[key] = darkSet
		capturePlan.DarksDone[key] = 0
		_, ok := capturePlan.DownloadTimes[darkSet.Binning]
		if !ok {
			capturePlan.DownloadTimes[darkSet.Binning] = 0
		}
	}

	//	Initialize the settings for every bias frame set needed
	for _, biasSet := range biasSets {
		key := MakeBiasKey(biasSet)
		capturePlan.BiasRequired[key] = biasSet
		capturePlan.BiasDone[key] = 0
		_, ok := capturePlan.DownloadTimes[biasSet.Binning]
		if !ok {
			capturePlan.DownloadTimes[biasSet.Binning] = 0
		}
	}

	return &capturePlan
}

func MakeDarkKey(set config.DarkSet) string {
	return fmt.Sprintf("Dark_%d_%.4f_%d", set.Frames, set.Seconds, set.Binning)
}

func MakeBiasKey(set config.BiasSet) string {
	return fmt.Sprintf("Bias_%d_%d", set.Frames, set.Binning)
}

func (s *Session) updateDownloadTimes(capturePlan *CapturePlan) error {
	//fmt.Printf("updateDownloadTimes. CapturePlan: %#v\n", *capturePlan)
	for binning, seconds := range capturePlan.DownloadTimes {
		//fmt.Printf("  Binning %d, download time %g\n", binning, seconds)
		if seconds == 0 {
			if s.settings.Verbosity > 1 || s.settings.Debug {
				fmt.Printf("Measuring download time for binning %d\n", binning)
			}
			measuredTime, err := s.theSkyService.MeasureDownloadTime(binning)
			if err != nil {
				return errors.New("error measuring download time")
			}
			capturePlan.DownloadTimes[binning] = measuredTime
		}
	}
	//fmt.Printf("  Download times now %#v\n", capturePlan.DownloadTimes)
	return nil
}

func (s *Session) captureFrames(capturePlan *CapturePlan, coolingConfig config.CoolingConfig) error {
	if s.settings.Verbosity > 2 || s.settings.Debug {
		fmt.Printf("captureFrames. CapturePlan: %#v, cooling: %#v\n", *capturePlan, coolingConfig)
	}

	if err := s.captureDarkFrames(capturePlan, coolingConfig); err != nil {
		fmt.Println("Error in Session captureFrames, capturing dark frames:", err)
		return err
	}

	if err := s.captureBiasFrames(capturePlan, coolingConfig); err != nil {
		fmt.Println("Error in Session captureFrames, capturing dark frames:", err)
		return err
	}
	return nil
}

func (s *Session) captureDarkFrames(capturePlan *CapturePlan, coolingConfig config.CoolingConfig) error {
	//fmt.Println("captureDarkFrames ")
	//fmt.Printf("   Frames required: %v\n", capturePlan.DarksRequired)
	//fmt.Printf("   Frames done: %v\n", capturePlan.DarksDone)
	//fmt.Printf("   Download times: %v\n", capturePlan.DownloadTimes)
	//fmt.Printf("   Cooling config: %v\n", coolingConfig)
	for key, set := range capturePlan.DarksRequired {
		//fmt.Printf("   Checking dark set %s: %v\n", key, set)
		if err := s.captureDarkSet(capturePlan, key, set, coolingConfig); err != nil {
			fmt.Println("Error in Session captureDarkFrames, capturing dark set:", err)
			return err
		}
	}
	return nil
}

func (s *Session) captureDarkSet(plan *CapturePlan, key string, set config.DarkSet, coolingConfig config.CoolingConfig) error {
	if s.settings.Verbosity > 0 || s.settings.Debug {
		fmt.Printf("Handling dark frames set: %d frames of %.2f seconds binned %d\n", set.Frames, set.Seconds, set.Binning)
	}
	if plan.DarksDone[key] >= set.Frames {
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("  Already have all %d dark frames in set %s\n", set.Frames, key)
		}
		return nil
	}

	framesNeeded := set.Frames - plan.DarksDone[key]
	if framesNeeded > 0 {
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("  Still need %d dark frames (of %d) in set %s\n", framesNeeded, set.Frames, key)
		}
	}
	frameCount := 0
	for plan.DarksDone[key] < set.Frames {
		abandon, err := s.CheckAbandonForCooling(coolingConfig)
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
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("    Capturing dark frame %d of %d:  %.2f seconds binned %d\n", frameCount, framesNeeded, set.Seconds, set.Binning)
		}

		if err := s.theSkyService.CaptureDarkFrame(set.Binning, set.Seconds, plan.DownloadTimes[set.Binning]); err != nil {
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

func (s *Session) captureBiasFrames(capturePlan *CapturePlan, coolingConfig config.CoolingConfig) error {
	//fmt.Println("captureBiasFrames ")
	//fmt.Printf("   Frames required: %v\n", capturePlan.BiasRequired)
	//fmt.Printf("   Frames done: %v\n", capturePlan.BiasDone)
	//fmt.Printf("   Download times: %v\n", capturePlan.DownloadTimes)
	//fmt.Printf("   Cooling config: %v\n", coolingConfig)

	for key, set := range capturePlan.BiasRequired {
		//fmt.Printf("   Checking bias set %s: %v\n", key, set)
		if err := s.captureBiasSet(capturePlan, key, set, coolingConfig); err != nil {
			fmt.Println("Error in Session captureBiasFrames, capturing bias set:", err)
			return err
		}
	}
	return nil
}

func (s *Session) captureBiasSet(plan *CapturePlan, key string, set config.BiasSet, coolingConfig config.CoolingConfig) error {
	if s.settings.Verbosity > 0 || s.settings.Debug {
		fmt.Printf("Handling bias frames set: %d frames  binned %d\n", set.Frames, set.Binning)
	}
	if plan.BiasDone[key] >= set.Frames {
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("  Already have all %d bias frames in set %s\n", set.Frames, key)
		}
		return nil
	}

	framesNeeded := set.Frames - plan.BiasDone[key]
	if framesNeeded > 0 {
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("  Still need %d bias frames (of %d) in set %s\n", framesNeeded, set.Frames, key)
		}
	}
	frameCount := 0
	for plan.BiasDone[key] < set.Frames {
		abandon, err := s.CheckAbandonForCooling(coolingConfig)
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
		if s.settings.Verbosity > 1 || s.settings.Debug {
			fmt.Printf("    Capturing bias frame %d of %d, binned %d\n", frameCount, framesNeeded, set.Binning)
		}

		if err := s.theSkyService.CaptureBiasFrame(set.Binning, plan.DownloadTimes[set.Binning]); err != nil {
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

func (s *Session) CheckAbandonForCooling(coolingConfig config.CoolingConfig) (bool, error) {
	//fmt.Println("CheckAbandonForCooling")
	if !coolingConfig.AbortOnCooling {
		return false, nil
	}
	cameraTemperature, err := s.theSkyService.GetCameraTemperature()
	//fmt.Println("  Camera temperature:", cameraTemperature)
	if err != nil {
		fmt.Println("Error in Session CheckAbandonForCooling, getting camera temperature:", err)
		return false, err
	}
	variation := math.Abs(cameraTemperature - coolingConfig.CoolTo)
	//fmt.Printf("  Temp %g and target %g = variation %g\n", cameraTemperature, coolingConfig.CoolTo, variation)
	if variation >= coolingConfig.CoolAbortTol {
		// Camera temperature is unacceptable - return an abort request
		return true, nil
	}
	return false, nil
}
