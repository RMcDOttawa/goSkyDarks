package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"goskydarks/config"
	"os"
)

// StateFileService abstracts reading and writing capture plan information to a state file.
// It is packaged as a separate service so it can be mocked for testing

type StateFileService interface {
	SavePlanToFile(plan *CapturePlan) error
	UpdatePlanFromFile(plan *CapturePlan) error
	ReadStateFile() (*CapturePlan, error)
}

type StateFileServiceInstance struct {
	StateFilePath string
}

func NewStateFileService(stateFilePath string) StateFileService {
	service := &StateFileServiceInstance{}
	service.StateFilePath = stateFilePath
	return service
}

func (sfs *StateFileServiceInstance) SavePlanToFile(capturePlan *CapturePlan) error {
	//fmt.Println("StateFileServiceInstance/SavePlanToFile()")
	if viper.GetInt(config.VerbositySetting) > 2 {
		fmt.Println("StateFileService/SavePlanToFile()")
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
	numWritten, err := file.Write([]byte(jsonBytes))
	if err != nil {
		fmt.Println("Unable to write new state data file")
		return err
	}
	if numWritten != len(jsonBytes) {
		fmt.Printf("Expected to write %d bytes to state file, but actually wrote %d bytes.\n",
			len(jsonBytes), numWritten)
	}

	return nil
}

func (sfs *StateFileServiceInstance) UpdatePlanFromFile(capturePlan *CapturePlan) error {
	//fmt.Println("StateFileServiceInstance/UpdatePlanFromFile()")

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

	//	Update counts of what is already done
	for key, count := range capturePlan.BiasDone {
		stateFileCount := stateFilePlan.BiasDone[key]
		if stateFileCount > count {
			capturePlan.BiasDone[key] = stateFileCount
		}
	}
	for key, count := range capturePlan.DarksDone {
		stateFileCount := stateFilePlan.DarksDone[key]
		if stateFileCount > count {
			capturePlan.DarksDone[key] = stateFileCount
		}
	}

	//	Update download times
	for binning, downloadTime := range capturePlan.DownloadTimes {
		stateFileTime := stateFilePlan.DownloadTimes[binning]
		if stateFileTime > downloadTime {
			capturePlan.DownloadTimes[binning] = stateFileTime
		}
	}

	return nil
}

func (sfs *StateFileServiceInstance) ReadStateFile() (*CapturePlan, error) {
	//fmt.Println("StateFileServiceInstance/ReadStateFile()")
	//fmt.Printf("ReadStateFile.  Path: %s\n", stateFilePath)

	//	See if file exists
	_, err := os.Stat(sfs.StateFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	//	Read file into json string
	fileContentsBytes, err := os.ReadFile(sfs.StateFilePath)
	fileContents := string(fileContentsBytes)

	//	Unmarshall JSON to data structure
	var stateFilePlan = &CapturePlan{}
	err = json.Unmarshal([]byte(fileContents), stateFilePlan)
	if err != nil {
		return nil, errors.New("error unmarshalling state file")
	}

	return stateFilePlan, nil
}
