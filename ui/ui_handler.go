package ui

import (
	"fmt"
	"rasptube-client/models"

	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type UIEventType int

const (
	Exit UIEventType = iota
	PlaybackNext
	PlaybackPrev
	PlaybackToggle
	PlayTrackByID
)

type UIEvent struct {
	Type    UIEventType
	Payload interface{}
}

type PlayTrackByIDPayload struct {
	TrackID uint32
}

type UIHandler struct {
	playlistWidget *widgets.List
	grid           *termui.Grid
	controlsWidget *widgets.Table
	seekbarWidget  *widgets.Gauge
	playlist       *models.Playlist
}

func NewUIHandler(playlist *models.Playlist) (*UIHandler, error) {
	if err := termui.Init(); err != nil {
		return nil, err
	}

	l := widgets.NewList()
	l.Title = "Playlist"
	l.TextStyle = termui.NewStyle(termui.ColorBlue)
	l.WrapText = false
	// l.SetRect(0, 0, termWidth, termHeight)

	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	controls := widgets.NewTable()
	controls.Rows = [][]string{
		{"Prev", "Play", "Next"},
		{"<p>", "<space>", "<n>"},
	}

	controls.TextAlignment = termui.AlignCenter
	controls.TextStyle = termui.NewStyle(termui.ColorWhite)
	controls.RowSeparator = false
	controls.ColSeparator = false
	controls.Border = false

	seekbar := widgets.NewGauge()
	seekbar.Percent = 0
	seekbar.Border = false

	grid.Set(
		termui.NewRow(8.0/10, l),
		termui.NewRow(1.0/10,
			termui.NewCol(1.0/3, nil),
			termui.NewCol(1.0/3, controls),
			termui.NewCol(1.0/3, nil),
		),
		termui.NewRow(1.0/10,
			termui.NewCol(1.0/3, nil),
			termui.NewCol(1.0/3, seekbar),
			termui.NewCol(1.0/3, nil),
		),
	)

	termui.Render(grid)
	return &UIHandler{
		playlistWidget: l,
		grid:           grid,
		controlsWidget: controls,
		seekbarWidget:  seekbar,
		playlist:       playlist,
	}, nil
}

func (h *UIHandler) Close() {
	termui.Close()
}

func (h *UIHandler) consumeUIEvents(ch chan<- UIEvent) {
	uiEvents := termui.PollEvents()
	previousKey := ""

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			ch <- UIEvent{Type: Exit}
		case "n":
			ch <- UIEvent{Type: PlaybackNext}
		case "p":
			ch <- UIEvent{Type: PlaybackPrev}
		case "<Space>":
			ch <- UIEvent{Type: PlaybackToggle}
		case "<Enter>":
			ch <- UIEvent{Type: PlayTrackByID, Payload: PlayTrackByIDPayload{TrackID: h.playlist.Tracks[h.playlistWidget.SelectedRow].ID}}
		case "j", "<Down>":
			h.playlistWidget.ScrollDown()
		case "k", "<Up>":
			h.playlistWidget.ScrollUp()
		case "<C-d>":
			h.playlistWidget.ScrollHalfPageDown()
		case "<C-u>":
			h.playlistWidget.ScrollHalfPageUp()
		case "<C-f>":
			h.playlistWidget.ScrollPageDown()
		case "<C-b>":
			h.playlistWidget.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				h.playlistWidget.ScrollTop()
			}
		case "<Home>":
			h.playlistWidget.ScrollTop()
		case "G", "<End>":
			h.playlistWidget.ScrollBottom()
		case "<Resize>":
			x, y := termui.TerminalDimensions()
			h.playlistWidget.SetRect(0, 0, x, y)
		}
		termui.Render(h.grid)
	}
}

func (h *UIHandler) StartHandlingEvents() <-chan UIEvent {
	ch := make(chan UIEvent)
	go h.consumeUIEvents(ch)
	return ch
}

func (h *UIHandler) Update(state *models.PlaybackState) {
	h.playlistWidget.Rows = h.updateRows(state)

	if state.Playing {
		h.controlsWidget.Rows[0][1] = "Pause"
	} else {
		h.controlsWidget.Rows[0][1] = "Play"
	}

	if state.TrackTotalTime > 0 {
		h.seekbarWidget.Percent = int(100 * (float32(state.TrackCurrentTime) / float32(state.TrackTotalTime)))
	} else {
		h.seekbarWidget.Percent = 0
	}

	h.seekbarWidget.Label = fmt.Sprintf("%s/%s", FormatMilliseconds(state.TrackCurrentTime), FormatMilliseconds(state.TrackTotalTime))

	termui.Render(h.grid)
}

func (h *UIHandler) updateRows(state *models.PlaybackState) []string {
	var rows []string
	for _, track := range h.playlist.Tracks {
		name := track.Name
		if state.TrackID == track.ID {
			name = fmt.Sprintf("[▶ %s](fg:red)", track.Name)
		}
		rows = append(rows, name)
	}

	return rows
}
