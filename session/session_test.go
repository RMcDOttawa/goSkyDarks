package session

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"goskydarks/config"
	"goskydarks/delay"
	"goskydarks/theSkyX"
	"sync"
	"testing"
)

// TestCoolForStart tests the cooling-for-start function of the session.
// We mock the TheSkyService service to simulate responses from the server
func TestCoolForStart(t *testing.T) {

	var mutex sync.Mutex
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//	Test successfully starting cooling and reaching target immediately
	t.Run("cooling reaches temp immediately", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		coolingConfig := config.CoolingConfig{
			UseCooler:       true,
			CoolTo:          -10,
			CoolStartTol:    2,
			CoolWaitMinutes: 30,
		}
		serverConfig := config.ServerConfig{
			Address: "localhost",
			Port:    3040,
		}
		settings := config.SettingsType{
			Cooling: coolingConfig,
		}
		session, err := NewSession(settings)
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockTheSkyService.EXPECT().Connect(serverConfig.Address, serverConfig.Port).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(coolingConfig.CoolTo).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(coolingConfig.CoolTo, nil)

		err = session.ConnectToServer(serverConfig)
		require.Nil(t, err, "Can't connect to server")

		err = session.CoolForStart(coolingConfig)
		require.Nil(t, err, "Can't cool for start")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

	//	Test starting cooling and reaching target after several polls
	t.Run("cooling reaches temp after several tries", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		coolingConfig := config.CoolingConfig{
			UseCooler:       true,
			CoolTo:          -10,
			CoolStartTol:    2,
			CoolWaitMinutes: 30,
		}
		serverConfig := config.ServerConfig{
			Address: "localhost",
			Port:    3040,
		}
		settings := config.SettingsType{
			Cooling: coolingConfig,
		}
		session, err := NewSession(settings)
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockDelayService.EXPECT().DelayDuration(120).AnyTimes().Return(120, nil)
		mockTheSkyService.EXPECT().Connect(serverConfig.Address, serverConfig.Port).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(coolingConfig.CoolTo).Return(nil)
		gomock.InOrder(
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(0.0, nil),
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(-5.0, nil),
			mockTheSkyService.EXPECT().GetCameraTemperature().Return(-9.0, nil),
		)

		err = session.ConnectToServer(serverConfig)
		require.Nil(t, err, "Can't connect to server")

		err = session.CoolForStart(coolingConfig)
		require.Nil(t, err, "Can't cool for start")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

	//	Test starting cooling and not reaching target before timeout
	t.Run("cooling fails to reach temp", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		coolingConfig := config.CoolingConfig{
			UseCooler:       true,
			CoolTo:          -10,
			CoolStartTol:    2,
			CoolWaitMinutes: 30,
		}
		serverConfig := config.ServerConfig{
			Address: "localhost",
			Port:    3040,
		}
		settings := config.SettingsType{
			Cooling: coolingConfig,
		}
		session, err := NewSession(settings)
		require.Nil(t, err, "Can't create session")
		mockDelayService := delay.NewMockDelayService(ctrl)
		session.SetDelayService(mockDelayService)
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)

		mockDelayService.EXPECT().DelayDuration(120).AnyTimes().Return(120, nil)
		mockTheSkyService.EXPECT().Connect(serverConfig.Address, serverConfig.Port).Return(nil)
		mockTheSkyService.EXPECT().Close().Return(nil)
		mockTheSkyService.EXPECT().StartCooling(coolingConfig.CoolTo).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-1.0, nil)

		err = session.ConnectToServer(serverConfig)
		require.Nil(t, err, "Can't connect to server")

		err = session.CoolForStart(coolingConfig)
		require.NotNil(t, err, "Expected cooling to fail on timeout")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

}

// Test capturing dark frames
func TestDarkFrameCapture(t *testing.T) {

	var mutex sync.Mutex

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Capture all frames since no intermediate results", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		_, coolingConfig, session, capturePlan := setUpDarkFrameCapture(t)
		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err := session.captureDarkFrames(capturePlan, coolingConfig)
		require.Nil(t, err, "Dark frame capture should not report error")
	})

	t.Run("Capture remaining frames - state file says some are done", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		darkSet, coolingConfig, session, capturePlan := setUpDarkFrameCapture(t)
		//	Modify plan for so 1 of the 3 dark frames is already done
		capturePlan.DarksDone[MakeDarkKey(darkSet)] = 1
		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up mock expects. Now there should only be 2 captures because 1 is done
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).Return(nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().AnyTimes().Return(-10.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err := session.captureDarkFrames(capturePlan, coolingConfig)
		require.Nil(t, err, "Dark frame capture should not report error")
	})

	t.Run("Capturing all frames - but abort when temperature rises too far", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		_, coolingConfig, session, capturePlan := setUpDarkFrameCapture(t)
		//	Mock services
		mockTheSkyService := theSkyX.NewMockTheSkyService(ctrl)
		session.SetTheSkyService(mockTheSkyService)
		mockStateFileService := NewMockStateFileService(ctrl)
		session.SetStateFileService(mockStateFileService)

		//	Set up mock expects
		mockTheSkyService.EXPECT().CaptureDarkFrame(1, 5.0, 5.0).AnyTimes().Return(nil)
		// Mock temperature rising beyond the tolerance after one successful frame
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-10.0, nil)
		mockTheSkyService.EXPECT().GetCameraTemperature().Return(-7.0, nil)
		mockStateFileService.EXPECT().SavePlanToFile(capturePlan).AnyTimes().Return(nil)
		err := session.captureDarkFrames(capturePlan, coolingConfig)
		require.NotNil(t, err, "Dark frame capture should report error")
		require.ErrorContains(t, err, "exceeding cooling tolerance", "Error message should contain 'exceeding cooling tolerance'")
	})
}

func setUpDarkFrameCapture(t *testing.T) (config.DarkSet, config.CoolingConfig, *Session, *CapturePlan) {
	coolingConfig := config.CoolingConfig{
		UseCooler:       true,
		CoolTo:          -10,
		CoolStartTol:    2,
		CoolWaitMinutes: 30,
		AbortOnCooling:  true,
		CoolAbortTol:    2.0,
	}
	settings := config.SettingsType{
		Cooling: coolingConfig,
	}
	session, err := NewSession(settings)
	require.Nil(t, err, "Can't create session")
	//	Set up plan for 3 dark frames
	darksRequired := make(map[string]config.DarkSet)
	darkSet := config.DarkSet{
		Frames:  3,
		Seconds: 5.0,
		Binning: 1,
	}
	darksRequired[MakeDarkKey(darkSet)] = darkSet
	darksDone := make(map[string]int)
	darksDone[MakeDarkKey(darkSet)] = 0
	biasRequired := make(map[string]config.BiasSet)
	biasDone := make(map[string]int)
	downloadTimes := make(map[int]float64)
	downloadTimes[1] = 5.0
	capturePlan := &CapturePlan{
		DarksRequired: darksRequired,
		DarksDone:     darksDone,
		BiasRequired:  biasRequired, // none
		BiasDone:      biasDone,
		DownloadTimes: downloadTimes,
	}
	return darkSet, coolingConfig, session, capturePlan
}
