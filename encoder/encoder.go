package encoder

import (
	"fmt"
	"strings"
	"time"

	"github.com/mariolopjr/apple-csv-to-lastfm/decoder"
)

func (lfp *LastFm) Marshal(amp decoder.AppleMusic) error {
	lfp.Scrobbles = make([]Scrobble, 0, len(amp.Plays))
	keys := make(map[string]bool)

	for _, play := range amp.Plays {
		// Timestamp is tweaked to not include timezone
		scrobble := Scrobble{
			Artist:         play.ArtistName,
			Album:          play.AlbumName,
			Title:          play.SongName,
			DateTimePlayed: strings.TrimSuffix(play.PlayTimestamp.Format(time.RFC822), " UTC"),
		}

		// Remove dupes
		scrobbleKey := fmt.Sprintf("%s-%s-%s-%s", scrobble.Artist, scrobble.Album, scrobble.Title, scrobble.DateTimePlayed)
		if _, value := keys[scrobbleKey]; !value {
			keys[scrobbleKey] = true
			lfp.Scrobbles = append(lfp.Scrobbles, scrobble)
		}
	}

	return nil
}
