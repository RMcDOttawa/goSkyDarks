package delaypkg

import (
	"fmt"
	"time"
)

//	DelayService provides a simple delaypkg until a given time (or for a given duration)
//	It is implemented as a service so it can be injected into other services which,
//	in turn, will facilitate testing those other services with a mock delaypkg

type DelayService interface {
	DelayDuration(seconds int) (int, error)
	DelayUntil(target time.Time) error

	SetDebug(debug bool)
	SetVerbosity(verbosity int)
}

type DelayServiceInstance struct {
	debug     bool
	verbosity int
}

func (s *DelayServiceInstance) SetDebug(debug bool) {
	s.debug = debug
}

func (s *DelayServiceInstance) SetVerbosity(verbosity int) {
	s.verbosity = verbosity
}

func NewDelayService(debug bool,
	verbosity int) DelayService {
	service := &DelayServiceInstance{
		debug:     debug,
		verbosity: verbosity,
	}
	return service
}

// DelayDuration implements a simple sleep for the given number of seconds
//
//	We return the number of seconds to facilitate mocking with time tracking
func (s *DelayServiceInstance) DelayDuration(seconds int) (int, error) {
	if s.verbosity >= 4 {
		fmt.Println("DelayServiceInstance DelayDuration:", seconds)
	}
	if seconds <= 0 {
		return 0, nil
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	return seconds, nil
}

func (s *DelayServiceInstance) DelayUntil(target time.Time) error {
	//	Calculate duration from now until the target time
	now := time.Now()
	duration := target.Sub(now)

	//	Delay for that long
	if duration > 0 {
		if s.verbosity >= 4 || s.debug {
			fmt.Printf("Waiting until %v (duration: %v)\n", target, duration)
		}
		_, _ = s.DelayDuration(int(duration / time.Second))
		if s.verbosity >= 4 || s.debug {
			fmt.Println("Reached the target time!")
		}
	} else {
		if s.verbosity >= 4 || s.debug {
			fmt.Println("The target time is already in the past.")
		}
	}
	return nil
}
