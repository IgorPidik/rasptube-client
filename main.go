package main

import (
	"context"
	"log"
	"rasptube-client/client"
	"rasptube-client/ui"
	"sync"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	client, clientErr := client.NewClient(ctx)
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
	clientEvents := client.PollClientEvents(wg)

out:
	for {
		select {
		case e := <-uiEvents:
			switch e.Type {
			case ui.Exit:
				cancel()
				break out
			case ui.PlaybackNext:
				client.PlaybackNext()
			case ui.PlaybackPrev:
				client.PlaybackPrev()
			case ui.PlaybackToggle:
				client.PlaybackTogglePlay()
			case ui.PlayTrackByID:
				if payload, ok := e.Payload.(ui.PlayTrackByIDPayload); ok {
					client.PlayTrack(payload.TrackID)
				}
			}
		case state := <-clientEvents:
			uiHandler.Update(&state)
		}
	}
	wg.Wait()
}
