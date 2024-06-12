package theSkyX

import (
	"errors"
	"fmt"
	"goskydarks/config"
	"goskydarks/delay"
	"math"
)

//	TheSkyService is a high-level interface to the set of logical services we use to control
//	the TheSkyX app running on the network. It abstracts away the complexities of making up
//	JavaScript command packets and using sockets to communicate.

type TheSkyService interface {
	//	Open and close persistent socket connection to the server
	Connect(server string, port int) error
	Close() error
	StartCooling(targetTemp float64) error
	GetCameraTemperature() (float64, error)
	StopCooling() error
	MeasureDownloadTime(binning int) (float64, error)
	CaptureDarkFrame(binning int, seconds float64, downloadTime float64) error
}

type TheSkyServiceInstance struct {
	settings     config.SettingsType
	driver       TheSkyDriver
	isOpen       bool
	delayService delay.DelayService
}

// NewTheSkyService is the constructor for the instance of this service
func NewTheSkyService(settings config.SettingsType, delayService delay.DelayService) TheSkyService {
	service := &TheSkyServiceInstance{
		settings:     settings,
		isOpen:       false,
		driver:       NewTheSkyDriver(settings.Debug, settings.Verbosity),
		delayService: delayService,
	}
	return service
}

// Connect opens a connection to the TheSkyX application, via the low-level driver.
// The connection is kept open, ready to use.
func (service *TheSkyServiceInstance) Connect(server string, port int) error {
	//fmt.Printf("TheSkyServiceInstance/Connect(%s,%d)\n", server, port)
	if service.isOpen {
		fmt.Printf("TheSkyServiceInstance/Connect(%s,%d): Already connected\n", server, port)
		return nil // already open, nothing to do
	}

	if err := service.driver.Connect(server, port); err != nil {
		return err
	}
	service.isOpen = true
	return nil
}

// Close closes the connection to the TheSkyX server
func (service *TheSkyServiceInstance) Close() error {
	//fmt.Println("TheSkyServiceInstance/Close() ")
	if !service.isOpen {
		fmt.Println("TheSkyServiceInstance/Close(): Not open")
		return nil
	}

	if err := service.driver.Close(); err != nil {
		return err
	}
	service.isOpen = false
	return nil
}

// StartCooling turns on the camera's thermoelectric cooler (TEC) and sets target temp
func (service *TheSkyServiceInstance) StartCooling(targetTemp float64) error {
	//if service.settings.Debug || service.settings.Verbosity > 2 {
	//	fmt.Printf("TheSkyServiceInstance/startCooling(%g) entered\n", targetTemp)
	//}
	if !service.isOpen {
		return errors.New("TheSkyServiceInstance/StartCooling: Connection not open")
	}

	if err := service.driver.StartCooling(targetTemp); err != nil {
		fmt.Println("TheSkyServiceInstance/StartCooling error from driver:", err)
		return err
	}
	//if service.settings.Debug || service.settings.Verbosity > 2 {
	//	fmt.Printf("TheSkyServiceInstance/startCooling(%g) exits\n", targetTemp)
	//}
	return nil
}

func (service *TheSkyServiceInstance) StopCooling() error {
	//fmt.Println("TheSkyServiceInstance/StopCooling()")
	if !service.isOpen {
		return errors.New("TheSkyServiceInstance/StopCooling: Connection not open")
	}
	err := service.driver.StopCooling()
	if err != nil {
		fmt.Println("TheSkyServiceInstance/StopCooling error from driver:", err)
		return err
	}
	return nil

}

func (service *TheSkyServiceInstance) GetCameraTemperature() (float64, error) {
	//fmt.Println("TheSkyServiceInstance/GetCameraTemperature()")
	if !service.isOpen {
		return 0.0, errors.New("TheSkyServiceInstance/GetCameraTemperature: Connection not open")
	}
	temp, err := service.driver.GetCameraTemperature()
	if err != nil {
		fmt.Println("TheSkyServiceInstance/GetCameraTemperature error from driver:", err)
		return temp, err
	}
	return temp, nil
}

func (service *TheSkyServiceInstance) MeasureDownloadTime(binning int) (float64, error) {
	if !service.isOpen {
		return 0.0, errors.New("TheSkyServiceInstance/MeasureDownloadTime: Connection not open")
	}
	time, err := service.driver.MeasureDownloadTime(binning)
	if err != nil {
		fmt.Println("TheSkyServiceInstance/MeasureDownloadTime error from driver:", err)
		return time, err
	}
	return time, nil
}

const andALittleExtra = 0.5
const pollingInterval = 1.0 //	seconds between polls
const timeoutFactor = 5.0   // How much longer to wait than the exposure time

func (service *TheSkyServiceInstance) CaptureDarkFrame(binning int, seconds float64, downloadTime float64) error {
	if service.settings.Verbosity > 2 || service.settings.Debug {
		fmt.Printf("TheSkyServiceInstance/CaptureDarkFrame(%d, %g, %g) \n", binning, seconds, downloadTime)
	}
	err := service.driver.StartDarkFrameCapture(binning, seconds, downloadTime)
	if err != nil {
		fmt.Println("TheSkyServiceInstance/CaptureDarkFrame error from driver:", err)
		return err
	}
	//	Now we'll wait until the exposure is probably over - exposure time + download time
	delayUntilComplete := int(math.Round(seconds + downloadTime + andALittleExtra))
	if service.settings.Verbosity > 2 {
		fmt.Println("Exposure started. Waiting for ", delayUntilComplete)
	}
	if _, err := service.delayService.DelayDuration(delayUntilComplete); err != nil {
		fmt.Println("TheSkyServiceInstance/CaptureDarkFrame error from delay service:", err)
		return err
	}
	//	Now we poll the camera repeatedly until it reports done
	maximumWaitSeconds := (seconds + downloadTime) * timeoutFactor
	secondsWaitedSoFar := 0.0
	for {
		done, err := service.driver.IsCaptureDone()
		if err != nil {
			fmt.Println("TheSkyServiceInstance/CaptureDarkFrame error from IsCaptureDone:", err)
			return err
		}
		if done {
			if service.settings.Verbosity > 2 {
				fmt.Println("capture is done, returning")
			}
			return nil
		}
		if secondsWaitedSoFar > maximumWaitSeconds {
			return errors.New("TheSkyServiceInstance/CaptureDarkFrame: Timeout waiting for capture to finish")
		}
		if service.settings.Verbosity > 2 {
			fmt.Println("Camera not finished. Delaying ", pollingInterval)
		}
		if _, err := service.delayService.DelayDuration(int(math.Round(pollingInterval))); err != nil {
			fmt.Println("TheSkyServiceInstance/CaptureDarkFrame error from polling delay service:", err)
			return err
		}
		secondsWaitedSoFar += pollingInterval
	}
}
