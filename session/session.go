package session

import (
	"errors"
	"fmt"
	"goskydarks/config"
	"goskydarks/theSkyX"
	"math"
	"time"
)

// Session struct implements the session service, used for overall session control
// such as start time or resuming from saved state
type Session struct {
	delayService  DelayService //	Used to delay start; replace with mock for testing
	settings      config.SettingsType
	theSkyService theSkyX.TheSkyService
	isConnected   bool
}

func NewSession(settings config.SettingsType) (*Session, error) {
	concreteDelayService := &DelayServiceInstance{
		settings: settings,
	}
	tsxService := theSkyX.NewTheSkyService(
		settings,
	)
	session := &Session{
		delayService:  concreteDelayService,
		settings:      settings,
		theSkyService: tsxService,
	}
	return session, nil
}

// SetDelayService allows delay service to be replaced with a mock for testing
func (s *Session) SetDelayService(delayService DelayService) {
	s.delayService = delayService
}

// SetTheSkyService allows server service to be replaced with a mock for testing
func (s *Session) SetTheSkyService(theSkyService theSkyX.TheSkyService) {
	s.theSkyService = theSkyService
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
func (s *Session) CaptureFrames(biasSets []config.BiasSet, darkSets []config.DarkSet) error {
	fmt.Println("CaptureFrames STUB")
	fmt.Printf("   Bias: %v\n", biasSets)
	fmt.Printf("   Dark: %v\n", darkSets)
	return nil
}
