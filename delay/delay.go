package delay

import (
	"fmt"
	"github.com/spf13/viper"
	"goskydarks/config"
	"time"
)

//	DelayService provides a simple delay until a given time (or for a given duration)
//	It is implemented as a service so it can be injected into other services which,
//	in turn, will facilitate testing those other services with a mock delay

type DelayService interface {
	DelayDuration(seconds int) (int, error)
	DelayUntil(target time.Time) error
}

type DelayServiceInstance struct {
}

func NewDelayService() DelayService {
	service := &DelayServiceInstance{}
	return service
}

// DelayDuration implements a simple sleep for the given number of seconds
//
//	We return the number of seconds to facilitate mocking with time tracking
func (s *DelayServiceInstance) DelayDuration(seconds int) (int, error) {
	if viper.GetInt(config.VerbositySetting) > 4 {
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
	verbosity := viper.GetInt(config.VerbositySetting)
	debug := viper.GetBool(config.DebugSetting)

	//	Delay for that long
	if duration > 0 {
		if verbosity > 3 || debug {
			fmt.Printf("Waiting until %v (duration: %v)\n", target, duration)
		}
		_, _ = s.DelayDuration(int(duration / time.Second))
		if verbosity > 3 || debug {
			fmt.Println("Reached the target time!")
		}
	} else {
		//fmt.Println("The target time is already in the past.")
	}
	return nil
}
