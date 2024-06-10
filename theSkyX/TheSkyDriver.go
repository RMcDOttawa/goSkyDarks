package theSkyX

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

// TheSkyDriver is the low-level interface to the TheSkyX application's TCP server, running
// somewhere on the network.  Controlling TheSky involves sending small packets of JavaScript
// to the server via a TCP socket.

type TheSkyDriver interface {
	Connect(server string, port int) error
	Close() error
	StartCooling(temp float64) error
	GetCameraTemperature() (float64, error)
	StopCooling() error
}

type TheSkyDriverInstance struct {
	debug     bool
	verbosity int
	isOpen    bool
	server    string
	port      int
}

const maxTheSkyBuffer = 4096

// NewTheSkyDriver is the constructor for a working instance of the interface
func NewTheSkyDriver(debug bool, verbosity int) TheSkyDriver {
	driver := &TheSkyDriverInstance{
		debug:     debug,
		verbosity: verbosity,
	}
	return driver
}

// Connect simulates opening the connection.
//
//	In fact, all we do is remember the server coordinates. The actual open of the
//	socket is deferred until we have a command to send
func (driver *TheSkyDriverInstance) Connect(server string, port int) error {
	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/Connect(%s,%d) entered\n", server, port)
	}
	if driver.isOpen {
		fmt.Printf("TheSkyDriverInstance/Connect(%s,%d): Already connected\n", server, port)
		return nil // already open, nothing to do
	}
	driver.server = server
	driver.port = port
	driver.isOpen = true
	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/Connect(%s,%d) successful\n", server, port)
	}
	return nil
}

// Close severs the connection to the TCP socket for the TheSkyX server
func (driver *TheSkyDriverInstance) Close() error {
	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/Close() entered\n")
	}
	if !driver.isOpen {
		fmt.Println("TheSkyDriverInstance/Close(): Not open")
		return nil
	}
	driver.isOpen = false
	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/Close() successful\n")
	}
	return nil
}

// StartCooling sends server commands to turn on the TEC and set the target temperature
// No response is expected from these commands
func (driver *TheSkyDriverInstance) StartCooling(temperature float64) error {
	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/StartCooling(%g)  \n", temperature)
	}

	var commands strings.Builder
	commands.WriteString("ccdsoftCamera.Connect();\n")
	commands.WriteString("ccdsoftCamera.RegulateTemperature=true;\n")
	commands.WriteString("ccdsoftCamera.ShutDownTemperatureRegulationOnDisconnect=false;\n")
	commands.WriteString(fmt.Sprintf("ccdsoftCamera.TemperatureSetPoint=%f;\n", temperature))

	if err := driver.sendCommandIgnoreReply(commands.String()); err != nil {
		fmt.Println("StartCooling error from driver:", err)
		return err
	}
	return nil
}

func (driver *TheSkyDriverInstance) StopCooling() error {
	var commands strings.Builder
	commands.WriteString("ccdsoftCamera.Connect();\n")
	commands.WriteString("ccdsoftCamera.RegulateTemperature=false;\n")

	if err := driver.sendCommandIgnoreReply(commands.String()); err != nil {
		fmt.Println("StopCooling error from driver:", err)
		return err
	}
	return nil
}

// GetCameraTemperature polls TheSkyX for the current camera temperature and returns it
func (driver *TheSkyDriverInstance) GetCameraTemperature() (float64, error) {
	if driver.verbosity > 2 || driver.debug {
		fmt.Println("GetCameraTemperature()")
	}
	var commands strings.Builder
	commands.WriteString("ccdsoftCamera.Connect();\n")
	commands.WriteString("var temp=ccdsoftCamera.Temperature;\n")
	commands.WriteString("var Out;\n")
	commands.WriteString("Out=temp + \"\\n\";\n")

	numberResult, err := driver.sendCommandFloatReply(commands.String())
	if err != nil {
		fmt.Println("GetCameraTemperature error from driver:", err)
		return -1.0, err
	}
	return numberResult, nil
}

// sendCommandNoReply is an internal method that sends the given command string to the server.
// This is used for commands where no reply is to be read and processed by the caller
// (There is a reply from the server, but it is used only to verify successful execution)
func (driver *TheSkyDriverInstance) sendCommandIgnoreReply(command string) error {
	if driver.verbosity > 2 || driver.debug {
		fmt.Println("TheSkyDriverInstance/sendCommandIgnoreReply: ", command)
	}
	var message strings.Builder
	message.WriteString("/* Java Script */\n")
	message.WriteString("/* Socket Start Packet */\n")
	message.WriteString(command)
	message.WriteString("/* Socket End Packet */\n")

	_, err := driver.sendCommand(message.String())
	if err != nil {
		fmt.Println("sendCommandNoReply error from driver:", err)
		return err
	}
	return nil
}

// sendCommandFloatReply is an internal method that sends the given command string to the server.
// This is used for commands where a floating point number reply is to be read and processed by the caller
func (driver *TheSkyDriverInstance) sendCommandFloatReply(command string) (float64, error) {
	if driver.verbosity > 2 || driver.debug {
		fmt.Println("TheSkyDriverInstance/sendCommandFloatReply: ", command)
	}
	var message strings.Builder
	message.WriteString("/* Java Script */\n")
	message.WriteString("/* Socket Start Packet */\n")
	message.WriteString(command)
	message.WriteString("/* Socket End Packet */\n")

	responseString, err := driver.sendCommand(message.String())
	trimmedResponse := strings.TrimSpace(responseString)
	if err != nil {
		fmt.Println("sendCommandNoReply error from driver:", err)
		return 0.0, err
	}

	parsedNum, err := strconv.ParseFloat(trimmedResponse, 64)
	if err != nil {
		return parsedNum, errors.New("error parsing numeric result")
	}

	return parsedNum, nil
}

// sendCommand is an internal method that sends the given command packet to the server and
// returns whatever reply is received.
func (driver *TheSkyDriverInstance) sendCommand(command string) (string, error) {
	//fmt.Println("TheSkyDriverInstance/sendCommand:", command)
	//	This function must be mutex-locked in case of parallel activities
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	if driver.verbosity > 2 || driver.debug {
		fmt.Printf("TheSkyDriverInstance/sendCommand() opening socket(%s,%d)\n", driver.server, driver.port)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", driver.server, driver.port))
	if err != nil {
		fmt.Println("Error opening socket:", err)
		return "", err
	}
	defer func(conn net.Conn) {
		if driver.verbosity > 2 || driver.debug {
			fmt.Println("Closing socket")
		}
		_ = conn.Close()
	}(conn)

	numWritten, err := conn.Write([]byte(command))
	if err != nil {
		fmt.Println("sendCommand error from driver:", err)
		return "", err
	}
	if numWritten != len(command) {
		fmt.Println("sendCommand wrong number of bytes from driver")
		return "", errors.New("sendCommand wrong number of bytes from driver")
	}

	responseBuffer := make([]byte, maxTheSkyBuffer)
	numRead, err := conn.Read(responseBuffer)
	if err != nil {
		fmt.Println("sendCommand error from driver:", err)
		return "", err
	}

	//	Response will be of the form <data if any> | error line
	responseParts := strings.Split(string(responseBuffer[:numRead]), "|")
	responseText := responseParts[0]
	errorLine := strings.ToLower(responseParts[1])

	if errorLine == "" {
		return responseText, nil
	}
	if strings.HasPrefix(errorLine, "no error.") {
		return responseText, nil
	}
	return responseText, errors.New("TheSkyX error: " + errorLine)
}
