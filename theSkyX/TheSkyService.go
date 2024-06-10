package theSkyX

import (
	"errors"
	"fmt"
	"goskydarks/config"
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
}

type TheSkyServiceInstance struct {
	settings config.SettingsType
	driver   TheSkyDriver
	isOpen   bool
}

// NewTheSkyService is the constructor for the instance of this service
func NewTheSkyService(settings config.SettingsType) TheSkyService {
	service := &TheSkyServiceInstance{
		settings: settings,
		isOpen:   false,
		driver:   NewTheSkyDriver(settings.Debug, settings.Verbosity),
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
