package main

type appState struct {
	connected bool
	ui        interface{}
}

/*
	UI will be a collection of IOScreens, where you can get and output
*/

var myAppState *appState = &appState{
	connected: false,
	ui:        nil,
}

func main() {
	// Set up UI. To be passed to backend loop

	// Set up networking
	err := dialBackend()
	if err != nil {
		// Set failed backend connection message
	}
}
