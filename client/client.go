package client

import (
	"context"
	"encoding/json"
	"log"
	"rasptube-client/models"
	"strconv"
	"sync"

	zmq "github.com/go-zeromq/zmq4"
)

const (
	PLAYBACK_TOGGLE_PLAY = "PLAYBACK_TOGGLE_PLAY"
	PLAYBACK_STOP        = "PLAYBACK_STOP"
	PLAYBACK_PLAY        = "PLAYBACK_PLAY"
	PLAYBACK_NEXT        = "PLAYBACK_NEXT"
	PLAYBACK_PREV        = "PLAYBACK_PREV"
	PLAY_TRACK_BY_ID     = "PLAY_TACK_BY_ID"
	INIT_STATE           = "INIT_STATE"
)

type Client struct {
	req zmq.Socket
	sub zmq.Socket
}

func NewClient(ctx context.Context) (*Client, error) {
	req := zmq.NewReq(ctx)
	if err := req.Dial("tcp://localhost:5559"); err != nil {
		return nil, err
	}

	sub := zmq.NewSub(ctx)
	if err := sub.Dial("tcp://localhost:5563"); err != nil {
		return nil, err
	}

	if err := sub.SetOption(zmq.OptionSubscribe, ""); err != nil {
		return nil, err
	}

	return &Client{
		req: req,
		sub: sub,
	}, nil
}

func (c *Client) Close() {
	c.req.Close()
	c.sub.Close()
}

func (c *Client) Init() (*models.InitState, error) {
	msg, msgErr := c.sendCommand(INIT_STATE)
	if msgErr != nil {
		return nil, msgErr
	}
	var initState models.InitState
	if jsonErr := json.Unmarshal(msg.Frames[0], &initState); jsonErr != nil {
		return nil, jsonErr
	}

	return &initState, nil
}

func (c *Client) sendMessage(msg *zmq.Msg) (*zmq.Msg, error) {
	if err := c.req.Send(*msg); err != nil {
		return nil, err
	}

	recvMsg, recvErr := c.req.Recv()
	if recvErr != nil {
		return nil, recvErr
	}

	return &recvMsg, nil
}

func (c *Client) sendCommand(command string) (*zmq.Msg, error) {
	msg := zmq.NewMsgString(command)
	return c.sendMessage(&msg)
}

func (c *Client) PlaybackTogglePlay() error {
	_, err := c.sendCommand(PLAYBACK_TOGGLE_PLAY)
	return err
}

func (c *Client) PlayTrack(trackID uint32) error {
	command := []byte(PLAY_TRACK_BY_ID)
	stringTrackID := []byte(strconv.FormatUint(uint64(trackID), 10))
	request := zmq.NewMsgFrom(command, stringTrackID)
	_, err := c.sendMessage(&request)
	return err
}

func (c *Client) PlaybackNext() error {
	_, err := c.sendCommand(PLAYBACK_NEXT)
	return err
}

func (c *Client) PlaybackPrev() error {
	_, err := c.sendCommand(PLAYBACK_PREV)
	return err
}

func (c *Client) PollClientEvents(wg *sync.WaitGroup) <-chan models.PlaybackState {
	stateChan := make(chan models.PlaybackState)
	go listenToPublisher(c.sub, stateChan, wg)
	return stateChan
}

func listenToPublisher(sub zmq.Socket, ch chan<- models.PlaybackState, wg *sync.WaitGroup) {
	for {
		msg, err := sub.Recv()
		if err != nil {
			if err == context.Canceled {
				close(ch)
				wg.Done()
				return
			}
			log.Fatalf("could not receive message: %v", err)
		}

		var state models.PlaybackState
		if jsonErr := json.Unmarshal(msg.Frames[0], &state); jsonErr != nil {
			log.Fatalf("failed to unmarshal state: %v", jsonErr)
		}

		ch <- state
	}
}
