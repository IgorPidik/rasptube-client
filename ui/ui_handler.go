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
	list     *widgets.List
	grid     *termui.Grid
	playlist *models.Playlist
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
	seekbar.Percent = 100
	seekbar.Label = fmt.Sprintf("%v%% (100MBs free)", seekbar.Percent)
	seekbar.Border = false

	grid.Set(
		termui.NewRow(8.0/10, l),
		termui.NewRow(0.8/10,
			termui.NewCol(1.0/3, nil),
			termui.NewCol(1.0/3, controls),
			termui.NewCol(1.0/3, nil),
		),
		termui.NewRow(0.5/10,
			termui.NewCol(1.0/3, nil),
			termui.NewCol(1.0/3, seekbar),
			termui.NewCol(1.0/3, nil),
		),
	)

	termui.Render(grid)
	return &UIHandler{list: l, grid: grid, playlist: playlist}, nil
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
			ch <- UIEvent{Type: PlayTrackByID, Payload: PlayTrackByIDPayload{TrackID: h.playlist.Tracks[h.list.SelectedRow].ID}}
		case "j", "<Down>":
			h.list.ScrollDown()
		case "k", "<Up>":
			h.list.ScrollUp()
		case "<C-d>":
			h.list.ScrollHalfPageDown()
		case "<C-u>":
			h.list.ScrollHalfPageUp()
		case "<C-f>":
			h.list.ScrollPageDown()
		case "<C-b>":
			h.list.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				h.list.ScrollTop()
			}
		case "<Home>":
			h.list.ScrollTop()
		case "G", "<End>":
			h.list.ScrollBottom()
		case "<Resize>":
			x, y := termui.TerminalDimensions()
			h.list.SetRect(0, 0, x, y)
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
	h.list.Rows = h.updateRows(state)
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
