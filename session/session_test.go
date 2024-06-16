package session

import (
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"goskydarks/config"
	"goskydarks/delay"
	"goskydarks/theSkyX"
	"sync"
	"testing"
)

const serverAddress = "localhost"
const serverPort = 3040
const targetTemperature = -10.0
const coolStartPollingSeconds = 5

var testFuncMutex sync.Mutex

// TestCoolForStart tests the cooling-for-start function of the session.
// We mock the TheSkyService service to simulate responses from the server
func TestCoolForStart(t *testing.T) {
	testFuncMutex.Lock()
	defer testFuncMutex.Unlock()
	var subTestMutex sync.Mutex
	//	Test successfully starting cooling and reaching target immediately
	t.Run("cooling reaches temp immediately", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockTheSkyService.EXPECT().Connect(serverAddress, serverPort).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(targetTemperature).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(targetTemperature, nil)

		err = session.ConnectToServer()
		require.Nil(t, err, "Can't connect to server")

		err = session.StartCoolingForStart()
		require.Nil(t, err, "Can't cool for start")
		err = session.WaitForTargetTemperature()
		require.Nil(t, err, "Error waiting to reach target temperature")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

	//	Test starting cooling and reaching target after several polls
	t.Run("cooling reaches temp after several tries", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.StartPollSecondsSetting, coolStartPollingSeconds)
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockDelayService.EXPECT().DelayDuration(coolStartPollingSeconds).AnyTimes().Return(coolStartPollingSeconds, nil)
		mockTheSkyService.EXPECT().Connect(serverAddress, serverPort).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(targetTemperature).Return(nil)
		gomock.InOrder(
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(0.0, nil),
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(-5.0, nil),
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(-9.0, nil),
		)

		err = session.ConnectToServer()
		require.Nil(t, err, "Can't connect to server")

		err = session.StartCoolingForStart()
		require.Nil(t, err, "Can't begin cooling for start")
		err = session.WaitForTargetTemperature()
		require.Nil(t, err, "Error waiting to reach target temperature")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

	//	Test starting cooling and not reaching target before timeout
	t.Run("cooling fails to reach temp", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.StartPollSecondsSetting, coolStartPollingSeconds)
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockDelayService.EXPECT().DelayDuration(coolStartPollingSeconds).AnyTimes().Return(coolStartPollingSeconds, nil)
		mockTheSkyService.EXPECT().Connect(serverAddress, serverPort).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(targetTemperature).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-1.0, nil)

		err = session.ConnectToServer()
		require.Nil(t, err, "Can't connect to server")

		err = session.StartCoolingForStart()
		require.Nil(t, err, "Can't begin cooling for start")
		err = session.WaitForTargetTemperature()
		require.NotNil(t, err, "Expected cooling to fail on timeout")
		require.ErrorContains(t, err, "timed out")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

}

// Test capturing dark frames
func TestDarkFrameCapture(t *testing.T) {
	testFuncMutex.Lock()
	defer testFuncMutex.Unlock()
	var subTestMutex sync.Mutex

	t.Run("Capture all frames since no intermediate results", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureDarkFrames(capturePlan)
		require.Nil(t, err, "Dark frame capture should not report error")
	})

	t.Run("Record correct darksDone after capture", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureDarkFrames(capturePlan)
		require.Nil(t, err, "Dark frame capture should not report error")
		require.Equal(t, 1, len(capturePlan.DarksDone), "Should have 1 darksDone entry")
		require.Equal(t, 3, capturePlan.DarksDone[MakeDarkKey(3, 5.0, 1)], "Should have 3 dark frames done")
		require.Equal(t, 1, len(capturePlan.BiasDone), "Should have one biasDone entry")
		require.Equal(t, 0, capturePlan.BiasDone[MakeBiasKey(3, 1)], "Should have 0 bias frames done")
	})

	t.Run("Capture remaining frames - state file says some are done", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames, of which 1 is already done, so only 2 more need to be captured
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 1 // 1 frame already done
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects. Now there should only be 2 captures because 1 is done
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureDarkFrames(capturePlan)
		require.Nil(t, err, "Dark frame capture should not report error")
	})

	t.Run("Capturing all frames - but abort when temperature rises too far", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.AbortOnCoolingSetting, true)      // Abort if cooling fails
		viper.Set(config.CoolAbortTolSetting, 2.0)         // Abort if cooling fails by this much
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)

		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames, of which 1 is already done, so only 2 more need to be captured
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 1 // 1 frame already done
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).AnyTimes().Return(nil)
		// Mock temperature rising beyond the tolerance after one successful frame
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-10.0, nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-7.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureDarkFrames(capturePlan)
		require.NotNil(t, err, "Dark frame capture should report error")
		require.ErrorContains(t, err, "exceeding cooling tolerance", "Error message should contain 'exceeding cooling tolerance'")
	})

	t.Run("Confirm nodark flag prevents dark frames", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, false)
		viper.Set(config.NoDarkSetting, true)
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureDarkFrames(capturePlan)
		require.Nil(t, err, "Dark frame capture should not report error")
	})
}

// Test capturing bias frames
func TestBiasFrameCapture(t *testing.T) {
	testFuncMutex.Lock()
	defer testFuncMutex.Unlock()
	var subTestMutex sync.Mutex

	t.Run("Capture all frames since no intermediate results", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.AbortOnCoolingSetting, true)
		viper.Set(config.CoolAbortTolSetting, 2.0)
		viper.Set(config.CoolWaitMinutesSetting, 30) // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureBiasFrames(capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")
	})

	t.Run("Record correct biasDone after capture", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureBiasFrames(capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")
		require.Equal(t, 1, len(capturePlan.DarksDone), "Should have 1 darksDone entry")
		require.Equal(t, 0, capturePlan.DarksDone[MakeDarkKey(3, 5.0, 1)], "Should have 0 dark frames done")
		require.Equal(t, 1, len(capturePlan.BiasDone), "Should have one biasDone entry")
		require.Equal(t, 3, capturePlan.BiasDone[MakeBiasKey(3, 1)], "Should have 3 bias frames done")
	})

	t.Run("Capture remaining frames - state file says some are done", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, true)
		viper.Set(config.AbortOnCoolingSetting, false)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames, of which 1 is already done, so only 2 more need to be captured
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0 // 1 frame already done
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 1
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects. Now there should only be 2 captures because 1 is done
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Return(nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureBiasFrames(capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")
	})

	t.Run("Capturing all frames - but abort when temperature rises too far", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.CoolToSetting, targetTemperature) // Cool to this temperature
		viper.Set(config.CoolStartTolSetting, 2.0)         // Plus or minus this much
		viper.Set(config.CoolWaitMinutesSetting, 30)       // And wait this long, no longer
		viper.Set(config.AbortOnCoolingSetting, true)      // Abort if cooling fails
		viper.Set(config.CoolAbortTolSetting, 2.0)         // Abort if cooling fails by this much
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 dark frames, of which 1 is already done, so only 2 more need to be captured
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0 // 1 frame already done
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{""},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).AnyTimes().Return(nil)
		// Mock temperature rising beyond the tolerance after one successful frame
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-10.0, nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-7.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureBiasFrames(capturePlan)
		require.NotNil(t, err, "Bias frame capture should report error")
		require.ErrorContains(t, err, "exceeding cooling tolerance", "Error message should contain 'exceeding cooling tolerance'")
	})

	t.Run("Nobias flag prevents bias capture", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, false)
		viper.Set(config.NoBiasSetting, true)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureBiasFrames(capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")
	})
}

func TestCaptureOrder(t *testing.T) {
	testFuncMutex.Lock()
	defer testFuncMutex.Unlock()
	var subTestMutex sync.Mutex

	t.Run("Capture bias first", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, false)
		viper.Set(config.BiasFirstSetting, true) // ** This is the test
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.NoDarkSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames and 3 dark frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		gomock.InOrder(
			mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Times(3).Return(nil),
			mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Times(3).Return(nil),
		)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureFrames(false, capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")

	})

	t.Run("Capture dark first", func(t *testing.T) {
		subTestMutex.Lock()
		defer subTestMutex.Unlock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		//	Fake minimal viper settings needed for this test
		viper.Set(config.ServerAddressSetting, serverAddress)
		viper.Set(config.ServerPortSetting, serverPort)
		viper.Set(config.UseCoolerSetting, false)
		viper.Set(config.DarkFirstSetting, true) // ** This is the test
		viper.Set(config.NoBiasSetting, false)
		viper.Set(config.BiasFirstSetting, false)
		viper.Set(config.DarkFirstSetting, false)
		session, err := NewSession()
		require.Nil(t, err, "Can't create session")

		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up plan for 3 bias frames and 3 dark frames
		darksDone := make(map[string]int)
		darksDone[MakeDarkKey(3, 5.0, 1)] = 0
		biasDone := make(map[string]int)
		biasDone[MakeBiasKey(3, 1)] = 0
		downloadTimes := make(map[int]float64)
		downloadTimes[1] = 5.0
		capturePlan := &CapturePlan{
			DarksRequired: []string{"3,5.0,1"},
			BiasRequired:  []string{"3,1"},
			DarksDone:     darksDone,
			BiasDone:      biasDone,
			DownloadTimes: downloadTimes,
		}

		//	Set up mock expects
		gomock.InOrder(
			mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Times(3).Return(nil),
			mockTheSkyService.EXPECT().CaptureBiasFrame(1, 5.0).Times(3).Return(nil),
		)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err = session.captureFrames(true, capturePlan)
		require.Nil(t, err, "Bias frame capture should not report error")
	})

}
