package session

import (
	"fmt"
	"goskydarks/config"
	"time"
)

// Session struct implements the session service, used for overall session control
// such as start time or resuming from saved state
type Session struct {
	delayService DelayService //	Used to delay start; replace with mock for testing
	settings     config.SettingsType
}

func NewSession(settings config.SettingsType) (*Session, error) {
	concreteDelayService := &ConcreteDelayService{
		settings: settings,
	}
	session := &Session{
		delayService: concreteDelayService,
	}
	return session, nil
}

// SetDelayService allows delay service to be replaced with a mock for testing
func (s Session) SetDelayService(delayService DelayService) {
	s.delayService = delayService
}

func (s Session) DelayStart(startTime time.Time) error {
	fmt.Println("DelayStart to:", startTime)
	return s.delayService.DelayUntil(startTime)
}

func (s Session) ConnectToServer(server config.ServerConfig) error {
	fmt.Printf("ConnectToServer STUB: %#v\n", server)
	return nil
}

func (s Session) CoolForStart(cooling config.CoolingConfig) error {
	fmt.Printf("CoolForStart STUB: %#v\n", cooling)
	return nil
}

func (s Session) CaptureFrames(biasSets []config.BiasSet, darkSets []config.DarkSet) error {
	fmt.Println("CaptureFrames STUB")
	fmt.Printf("   Bias: %v\n", biasSets)
	fmt.Printf("   Dark: %v\n", darkSets)
	return nil
}
