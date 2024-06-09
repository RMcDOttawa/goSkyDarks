package session

import (
	"fmt"
	"goskydarks/config"
	"time"
)

//	DelayService provides a simple delay until a given time (or for a given duration)
//	It is implemented as a service so it can be injected into other services which,
//	in turn, will facilitate testing those other services with a mock delay

type DelayService interface {
	DelayDuration(seconds int64) error
	DelayUntil(target time.Time) error
}

type ConcreteDelayService struct {
	settings config.SettingsType
}

func (s *ConcreteDelayService) DelayDuration(seconds int64) error {
	fmt.Println("ConcreteDelayService DelayDuration:", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func (s *ConcreteDelayService) DelayUntil(target time.Time) error {
	//	Calculate duration from now until the target time
	now := time.Now()
	duration := target.Sub(now)

	//	Delay for that long
	if duration > 0 {
		if s.settings.Verbosity > 3 || s.settings.Debug {
			fmt.Printf("Waiting until %v (duration: %v)\n", target, duration)
		}
		_ = s.DelayDuration(int64(duration / time.Second))
		if s.settings.Verbosity > 3 || s.settings.Debug {
			fmt.Println("Reached the target time!")
		}
	} else {
		//fmt.Println("The target time is already in the past.")
	}
	return nil
}
