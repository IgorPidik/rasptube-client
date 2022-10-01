package models

type Track struct {
	ID   uint
	Name string
	Url  string
}

type Playlist struct {
	ID     uint
	Tracks []*Track
}

type PlaybackState struct {
	PlaylistID uint
	TrackID    uint
	Playing    bool
}

type InitState struct {
	State    *PlaybackState
	Playlist *Playlist
}
