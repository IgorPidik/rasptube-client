package models

type Track struct {
	ID   uint32
	Name string
	Url  string
}

type Playlist struct {
	ID     uint32
	Tracks []*Track
}

type PlaybackState struct {
	PlaylistID uint32
	TrackID    uint32
	Playing    bool
}

type InitState struct {
	State    *PlaybackState
	Playlist *Playlist
}
