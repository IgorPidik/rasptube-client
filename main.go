package main

import (
	"log"
	"rasptube-client/client"
	"rasptube-client/ui"
)

func main() {
	client, clientErr := client.NewClient()
	if clientErr != nil {
		log.Fatalf("failed to init client: %v\n", clientErr)
	}
	defer client.Close()

	initState, initErr := client.Init()
	if initErr != nil {
		log.Fatal(initErr)
	}

	playlist := initState.Playlist
	uiHandler, uiErr := ui.NewUIHandler(playlist)
	if uiErr != nil {
		log.Fatal(uiErr)
	}
	defer uiHandler.Close()

	uiHandler.Update(initState.State)

	uiEvents := uiHandler.StartHandlingEvents()
	clientEvents := client.PollClientEvents()

	for {
		select {
		case e := <-uiEvents:
			switch e {
			case ui.Exit:
				return
			case ui.PlaybackNext:
				client.PlaybackNext()
			case ui.PlaybackPrev:
				client.PlaybackPrev()
			case ui.PlaybackToggle:
				client.PlaybackTogglePlay()
			}
		case state := <-clientEvents:
			uiHandler.Update(&state)
		}
	}
}
