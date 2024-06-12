package theSkyX

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"goskydarks/config"
	"goskydarks/delay"
	"math"
	"testing"
)

// TestDarkCapture tests the ability to capture a single dark frame.
// We mock the TheSkyDriver service to simulate responses from the driver
func TestDarkCapture(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test of an ideal acquisition - we start the camera acquisition, wait a calculated amount
	// of time, and then see that the camera reports done
	t.Run("capture dark frame ready on time", func(t *testing.T) {
		mockDelayService, service, mockDriver := setUpDarkCaptureTest(ctrl)

		const binning = 1
		const seconds = 20.0
		const downloadTime = 5.0
		mockDriver.EXPECT().StartDarkFrameCapture(1, seconds, downloadTime).Return(nil)
		//	Initial delay while waiting for exposure
		initialDelay := int(math.Round(seconds + downloadTime + AndALittleExtra)) // from service
		mockDelayService.EXPECT().DelayDuration(initialDelay).Return(initialDelay, nil)
		//	Report capture done on first check
		mockDriver.EXPECT().IsCaptureDone().Return(true, nil)
		err := service.CaptureDarkFrame(binning, seconds, downloadTime)

		require.Nil(t, err, "CaptureDarkFrame failed")
	})

	// Test of an acquisition requiring extra wait.  We start the camera acquisition, wait a calculated amount,
	// then find it isn't finished. So we loop and poll two more times, then it is done.
	t.Run("capture dark frame requiring two extra waits", func(t *testing.T) {
		mockDelayService, service, mockDriver := setUpDarkCaptureTest(ctrl)

		const binning = 1
		const seconds = 20.0
		const downloadTime = 5.0
		mockDriver.EXPECT().StartDarkFrameCapture(1, seconds, downloadTime).Return(nil)
		//	Initial delay while waiting for exposure
		initialDelay := int(math.Round(seconds + downloadTime + AndALittleExtra)) // from service
		mockDelayService.EXPECT().DelayDuration(initialDelay).Return(initialDelay, nil)
		//	Extra waits between polls
		mockDelayService.EXPECT().DelayDuration(1).Return(1, nil).Times(2)
		//	Report capture not done on first or second check; done on third
		mockDriver.EXPECT().IsCaptureDone().Return(false, nil)
		mockDriver.EXPECT().IsCaptureDone().Return(false, nil)
		mockDriver.EXPECT().IsCaptureDone().Return(true, nil)

		err := service.CaptureDarkFrame(binning, seconds, downloadTime)
		require.Nil(t, err, "CaptureDarkFrame failed")
	})

	// Test of an acquisition timing out.  we start the camera acquisition, wait a calculated amount,
	// then continue to wait and poll, only to eventually time out with no completion.
	t.Run("capture dark frame requiring two extra waits", func(t *testing.T) {
		mockDelayService, service, mockDriver := setUpDarkCaptureTest(ctrl)

		const binning = 1
		const seconds = 20.0
		const downloadTime = 5.0
		mockDriver.EXPECT().StartDarkFrameCapture(1, seconds, downloadTime).Return(nil)
		//	Initial delay while waiting for exposure
		initialDelay := int(math.Round(seconds + downloadTime + AndALittleExtra)) // from service
		mockDelayService.EXPECT().DelayDuration(initialDelay).Return(initialDelay, nil)
		//	Extra waits between polls
		mockDelayService.EXPECT().DelayDuration(1).AnyTimes().Return(1, nil)
		//	Report capture not done no matter how often we ask
		mockDriver.EXPECT().IsCaptureDone().AnyTimes().Return(false, nil)

		err := service.CaptureDarkFrame(binning, seconds, downloadTime)
		require.NotNil(t, err, "CaptureDarkFrame should have timed out")
		require.ErrorContains(t, err, "Timeout waiting for capture")
	})

}

func setUpDarkCaptureTest(ctrl *gomock.Controller) (*delay.MockDelayService, TheSkyService, *MockTheSkyDriver) {
	//	Set up the service we are testing
	coolingConfig := config.CoolingConfig{
		UseCooler:       true,
		CoolTo:          -10,
		CoolStartTol:    2,
		CoolWaitMinutes: 30,
	}
	settings := config.SettingsType{
		Cooling: coolingConfig,
	}
	mockDelayService := delay.NewMockDelayService(ctrl)
	service := NewTheSkyService(settings, mockDelayService)
	// Plug mock driver into service
	mockDriver := NewMockTheSkyDriver(ctrl)
	service.SetDriver(mockDriver)
	return mockDelayService, service, mockDriver
}
