package session

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"goskydarks/config"
	"goskydarks/theSkyX"
	"sync"
	"testing"
)

// TestCoolForStart tests the cooling-for-start function of the session.
// We mock the TheSkyService service to simulate responses from the server
func TestCoolForStart(t *testing.T) {

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

	var mutex sync.Mutex

	//	Mock TheSkyService so it doesn't try to use the network
	session, err := NewSession(settings)
	require.Nil(t, err, "Can't create session")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//	Test successfully starting cooling and reaching target immediately
	t.Run("cooling reaches temp immediately", func(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()
		mockDelayService := NewMockDelayService(ctrl)
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
		mockDelayService := NewMockDelayService(ctrl)
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
		failingCoolingConfig := config.CoolingConfig{
			UseCooler:       true,
			CoolTo:          -10,
			CoolStartTol:    2,
			CoolWaitMinutes: 4,
		}
		mockDelayService := NewMockDelayService(ctrl)
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

		err = session.CoolForStart(failingCoolingConfig)
		require.NotNil(t, err, "Expected cooling to fail on timeout")

		err = session.Close()
		require.Nil(t, err, "Can't close session")
	})

}
