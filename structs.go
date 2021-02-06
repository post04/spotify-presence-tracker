package main

// Database - the database that's loaded from the json, it's just temp and used for saving
type Database []MainStruct

// MainStruct - the main struct for the database, holds specific users
type MainStruct struct {
	Userid         string              `json:"userid"`
	User           string              `json:"user"`
	ListeningArray []ListenArrayStruct `json:"listeningArray"`
}

// ListenArrayStruct - the struct that holds artists, songs, etc.
type ListenArrayStruct struct {
	ArtistName   string        `json:"artistName"`
	Songs        []SongsStruct `json:"songs"`
	TimesListend int           `json:"timesListend"`
}

// SongsStruct - the struct that holds songs, used in the listen array struct
type SongsStruct struct {
	SongName      string `json:"songName"`
	TimesListened int    `json:"timesListened"`
	Artist        string
}

// MainData - the database that we actually work with
type MainData struct {
	Userid         string             `json:"userid"`
	User           string             `json:"user"`
	ListeningArray map[string]*Artist `json:"listeningArray"`
}

// Artist - the map that has artists and songs per artist
type Artist struct {
	ArtistName   string           `json:"artistName"`
	Songs        map[string]*Song `json:"songs"`
	TimesListend int              `json:"timesListened"`
}

// Song - used in the Artist struct to hold their songs
type Song struct {
	SongName      string `json:"songName"`
	TimesListened int    `json:"timesListened"`
}

// ReactionListener - info of what reactions we need to listen for
type ReactionListener struct {
	Type                 string
	Specific             bool
	CurrentPage          int
	PageLimit            int
	UserID               string
	Data                 MainStruct
	SpecificArtist       string
	Author               string
	SpecificArtistStruct ListenArrayStruct
	Songs                []*Song
	LastUsed             int64
}
