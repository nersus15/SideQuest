package desktop

import "testing"

func TestBusyState(t *testing.T) {
	busy := []int{qunsNotPresent, qunsBusy, qunsRunningD3DFullScreen, qunsPresentationMode, qunsApp}
	free := []int{0, qunsAcceptsNotification, qunsQuietTime, 99}
	for _, s := range busy {
		if !busyState(s) {
			t.Errorf("state %d should be busy", s)
		}
	}
	for _, s := range free {
		if busyState(s) {
			t.Errorf("state %d should not be busy", s)
		}
	}
}
