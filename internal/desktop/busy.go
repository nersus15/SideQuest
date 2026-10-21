package desktop

// Values of QUERY_USER_NOTIFICATION_STATE from SHQueryUserNotificationState.
const (
	qunsNotPresent           = 1 // screen saver, locked or fast user switching
	qunsBusy                 = 2 // a fullscreen app is running
	qunsRunningD3DFullScreen = 3 // a fullscreen Direct3D game
	qunsPresentationMode     = 4 // presentation settings are on
	qunsAcceptsNotification  = 5
	qunsQuietTime            = 6 // first hour after setup; not "busy" for us
	qunsApp                  = 7 // a Windows Store app is running fullscreen
)

// busyState reports whether the user is in something Boo should not
// interrupt: a fullscreen app or game, a presentation, or a locked screen.
func busyState(state int) bool {
	switch state {
	case qunsNotPresent, qunsBusy, qunsRunningD3DFullScreen, qunsPresentationMode, qunsApp:
		return true
	}
	return false
}
