package client

import (
	"context"
	"encoding/json"
	"log"
	"rasptube-client/models"

	zmq "github.com/go-zeromq/zmq4"
)

const PLAYBACK_TOGGLE_PLAY = "PLAYBACK_TOGGLE_PLAY"
const PLAYBACK_PLAY = "PLAYBACK_PLAY"
const PLAYBACK_STOP = "PLAYBACK_STOP"
const PLAYBACK_NEXT = "PLAYBACK_NEXT"
const PLAYBACK_PREV = "PLAYBACK_PREV"
const INIT_STATE = "INIT_STATE"

type Client struct {
	req zmq.Socket
	sub zmq.Socket
}

func NewClient() (*Client, error) {
	req := zmq.NewReq(context.Background())
	if err := req.Dial("tcp://localhost:5559"); err != nil {
		return nil, err
	}

	sub := zmq.NewSub(context.Background())

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

func (c *Client) sendCommand(command string) (*zmq.Msg, error) {
	if err := c.req.Send(zmq.NewMsgString(command)); err != nil {
		return nil, err
	}

	msg, recvErr := c.req.Recv()
	if recvErr != nil {
		return nil, recvErr
	}

	return &msg, nil
}

func (c *Client) PlaybackTogglePlay() error {
	_, err := c.sendCommand(PLAYBACK_TOGGLE_PLAY)
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

func (c *Client) PollClientEvents() <-chan models.PlaybackState {
	stateChan := make(chan models.PlaybackState)
	go listenToPublisher(c.sub, stateChan)
	return stateChan
}

func listenToPublisher(sub zmq.Socket, ch chan<- models.PlaybackState) {
	for {
		msg, err := sub.Recv()
		if err != nil {
			log.Fatalf("could not receive message: %v", err)
		}

		var state models.PlaybackState
		if jsonErr := json.Unmarshal(msg.Frames[0], &state); jsonErr != nil {
			log.Fatalf("failed to unmarshal state: %v", jsonErr)
		}

		ch <- state
	}
}
