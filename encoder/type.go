package encoder

type LastFm struct {
	Scrobbles []Scrobble
}

type Scrobble struct {
	Artist         string `csv:""`
	Album          string `csv:""`
	Title          string `csv:""`
	DateTimePlayed string `csv:""`
}
