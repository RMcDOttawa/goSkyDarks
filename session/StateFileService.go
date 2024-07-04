package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"goskydarks/config"
	"os"
	"strconv"
	"strings"
	"sync"
)

// StateFileService abstracts reading and writing capture plan information to a state file.
// It is packaged as a separate service, so it can be mocked for testing

var mutex sync.Mutex

type StateFileService interface {
	SavePlanToFile(plan *CapturePlan) error
	UpdatePlanFromFile(plan *CapturePlan) error
	ReadStateFile() (*CapturePlan, error)
}

type StateFileServiceInstance struct {
	StateFilePathInput string
	StateFilePath      string // includes temperature
}

func NewStateFileService(stateFilePath string, temperature float64) StateFileService {
	service := &StateFileServiceInstance{}
	service.StateFilePathInput = stateFilePath
	tempAsString := strconv.FormatFloat(temperature, 'f', 3, 64)
	service.StateFilePath = stateFilePath + "_" + strings.ReplaceAll(tempAsString, ".", "_") + ".state"
	return service
}

func (sfs *StateFileServiceInstance) SavePlanToFile(capturePlan *CapturePlan) error {
	mutex.Lock()
	defer mutex.Unlock()

	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Println("StateFileService/SavePlanToFile()")
		fmt.Printf("  Plan: %#v\n", capturePlan)
	}
	jsonBytes, err := json.MarshalIndent(capturePlan, "", "   ")
	if err != nil {
		fmt.Println("Error in Session saveCapturePlan, marshalling plan:", err)
		return err
	}
	//fmt.Println("\n\n***\n\nJSON to save to file:", string(jsonBytes))

	file, err := os.OpenFile(sfs.StateFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Could not open file to write data:", err)
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	numWritten, err := file.Write(jsonBytes)
	if err != nil {
		fmt.Println("Unable to write new state data file")
		return err
	}
	if numWritten != len(jsonBytes) {
		fmt.Printf("Expected to write %d bytes to state file, but actually wrote %d bytes.\n",
			len(jsonBytes), numWritten)
	}

	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 5 {
		fmt.Println("SavePlanToFile exits")
	}
	return nil
}

func (sfs *StateFileServiceInstance) UpdatePlanFromFile(capturePlan *CapturePlan) error {
	mutex.Lock()
	defer mutex.Unlock()
	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Println("StateFileServiceInstance/UpdatePlanFromFile()")
		fmt.Println("   Plan:", capturePlan)
	}

	//	Read state file into a separate plan on the side.
	//  Note that "file not found" is not an error and results in a nil stateFilePlan
	stateFilePlan, err := sfs.ReadStateFile()
	if err != nil {
		fmt.Println("Error in Session updatePlanFromStateFile, reading state file:", err)
		return err
	}
	if stateFilePlan == nil {
		//	No state file, so nothing to update
		return nil
	}
	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 3 {
		fmt.Println("  Plan read from state file:", stateFilePlan)
	}

	//	Update counts of what is already done
	for key, count := range capturePlan.BiasDone {
		stateFileCount := stateFilePlan.BiasDone[key]
		if stateFileCount > count {
			if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 3 {
				fmt.Printf("  Replacing BiasDone[%s] %d with %d\n", key, count, stateFileCount)
			}
			capturePlan.BiasDone[key] = stateFileCount
		}
	}
	for key, count := range capturePlan.DarksDone {
		stateFileCount := stateFilePlan.DarksDone[key]
		if stateFileCount > count {
			if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 3 {
				fmt.Printf("  Replacing DarksDone[%s] %d with %d\n", key, count, stateFileCount)
			}
			capturePlan.DarksDone[key] = stateFileCount
		}
	}

	//	Update download times
	for binning, downloadTime := range capturePlan.DownloadTimes {
		stateFileTime := stateFilePlan.DownloadTimes[binning]
		if stateFileTime > downloadTime {
			if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 3 {
				fmt.Printf("  Replacing DownloadTime[%d] %g with %g\n", binning, downloadTime, stateFileTime)
			}
			capturePlan.DownloadTimes[binning] = stateFileTime
		}
	}

	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Println("UpdatePlanFromFile exits")
	}
	return nil
}

func (sfs *StateFileServiceInstance) ReadStateFile() (*CapturePlan, error) {
	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Printf("ReadStateFile.  Path: %s\n", sfs.StateFilePath)
	}

	//	See if file exists
	_, err := os.Stat(sfs.StateFilePath)
	if errors.Is(err, os.ErrNotExist) {
		if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
			fmt.Println("State file does not exist")
		}
		return nil, nil
	}

	//	Read file into json string
	fileContentsBytes, err := os.ReadFile(sfs.StateFilePath)
	fileContents := string(fileContentsBytes)
	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Printf("  Read %d bytes: %s", len(fileContents), fileContents)
	}
	//	Unmarshall JSON to data structure
	var stateFilePlan = &CapturePlan{}
	err = json.Unmarshal([]byte(fileContents), stateFilePlan)
	if err != nil {
		return nil, errors.New("error unmarshalling state file")
	}

	if viper.GetBool(config.DebugSetting) || viper.GetInt(config.VerbositySetting) >= 4 {
		fmt.Println("ReadStateFile exits")
	}
	return stateFilePlan, nil
}
