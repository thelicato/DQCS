package socket

func RunHost(socketPath string) {
	// Implementation is OS-specific
	runAsHost(socketPath)
}

func RunGuest() {
	// Implementation is OS-specific
	runAsGuest()
}
