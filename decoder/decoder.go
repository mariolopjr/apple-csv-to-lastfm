package decoder

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "musicbrainz"
	password = "musicbrainz"
	dbname   = "musicbrainz_db"
)

var db *sql.DB

func (amp *AppleMusic) Unmarshal(srcFilePath string, useMusicBrainz bool) error {
	zipReader, err := zip.OpenReader(srcFilePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Connect to MusicBrainz DB
	if useMusicBrainz {
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Printf("error connecting to MusicBrainz database: %v", err)
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Printf("error pinging MusicBrainz database: %v", err)
		}
	}

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "Apple_Media_Services.zip") {
			mediaReader, err := file.Open()
			if err != nil {
				return err
			}
			defer mediaReader.Close()

			mediaZip, err := NewZipFromReader(mediaReader, file.FileInfo().Size())
			if err != nil {
				return err
			}

			for _, mediaFile := range mediaZip.File {
				// Build list of library tracks
				if strings.Contains(mediaFile.Name, "Apple Music Library Tracks.json.zip") {
					libraryReader, err := mediaFile.Open()
					if err != nil {
						return err
					}
					defer libraryReader.Close()

					libraryZip, err := NewZipFromReader(libraryReader, mediaFile.FileInfo().Size())
					if err != nil {
						return err
					}

					for _, libraryFile := range libraryZip.File {
						if strings.Contains(libraryFile.Name, "Apple Music Library Tracks.json") {
							br, err := libraryFile.Open()
							if err != nil {
								return err
							}

							// Grab all tracks from json in a temp data structure
							temp := struct {
								Tracks []Track
							}{}
							jsonDecoder := json.NewDecoder(br)
							err = jsonDecoder.Decode(&temp.Tracks)
							br.Close()
							if err != nil {
								return err
							}

							// Create a bucket system mapping tracks to the artist
							amp.Tracks = make(map[string][]Track)
							for _, track := range temp.Tracks {
								if track.Album != "" && track.Title != "" {
									if track.Artist != "" {
										if _, ok := amp.Tracks[track.Artist]; !ok {
											amp.Tracks[track.Artist] = make([]Track, 0, 10)
										}
										amp.Tracks[track.Artist] = append(amp.Tracks[track.Artist], track)
									}

									// Bucket for album artist
									if track.AlbumArtist != "" {
										if _, ok := amp.Tracks[track.AlbumArtist]; !ok {
											amp.Tracks[track.AlbumArtist] = make([]Track, 0, 10)
										}
										amp.Tracks[track.AlbumArtist] = append(amp.Tracks[track.AlbumArtist], track)
									}
								}
							}
						}
					}
				}

				// Build list of plays, grab album from library tracks
				if strings.Contains(mediaFile.Name, "Apple Music Play Activity.csv") {
					notFound := make([]Play, 0, 100)
					br, err := mediaFile.Open()
					if err != nil {
						return err
					}
					defer br.Close()

					csvReader := csv.NewReader(br)
					dec, err := csvutil.NewDecoder(csvReader)
					if err != nil {
						log.Fatalf("error establishing csv decoder: %v", err)
					}

					for {
						p := Play{}

						if err := dec.Decode(&p); err == io.EOF {
							break
						} else if err != nil {
							log.Fatalf("unable to decode row: %v", err)
						}

						// Skip tune-in video junk
						if p.SongName == "Tune-In Video" || p.SongName == "Apple Music 1" || p.SongName == "" || p.ArtistName == "" {
							continue
						}

						// Perform checks to ensure this is an appropriate 'play'
						eventStart, startErr := time.Parse(time.RFC3339, p.EventStartTimestamp)
						eventEnd, err := time.Parse(time.RFC3339, p.EventEndTimestamp)
						if startErr != nil && err == nil {
							eventStart = eventEnd
						} else if startErr != nil && err != nil {
							continue
						}
						p.PlayTimestamp = eventStart

						// Removes bad rows before 2015 since that's not possible
						if p.PlayTimestamp.Year() < 2015 {
							continue
						}

						// Check if played completely
						playedCompletely := false
						if p.EndReasonType == "NATURAL_END_OF_TRACK" {
							playedCompletely = true
						} else if p.PlayDurationMilliseconds >= p.MediaDurationInMilliseconds {
							playedCompletely = true
						}

						// Get play duration
						playDurationInMilliseconds := p.MediaDurationInMilliseconds
						if eventStart.Day() == eventEnd.Day() {
							playDurationInMilliseconds = eventEnd.Sub(eventStart).Milliseconds()
						}
						if !playedCompletely && p.PlayDurationMilliseconds > 0 {
							playDurationInMilliseconds = p.PlayDurationMilliseconds
						}
						if float64(playDurationInMilliseconds)/float64(60000) > 60 {
							playDurationInMilliseconds = p.MediaDurationInMilliseconds
						}

						if float64(playDurationInMilliseconds)/float64(p.MediaDurationInMilliseconds) < 0.5 {
							continue
						}

						// Find album in media duration bucket
						if tracks, ok := amp.Tracks[p.ArtistName]; ok {
							// Iterate thru tracks in bucket
							for _, track := range tracks {
								// Check if we have a match to the track, including loose title comparisons
								// Example: Fugue - Trivium vs. Fugue
								var shortest, longest string
								if p.SongName > track.Title {
									shortest = track.Title
									longest = p.SongName
								} else {
									shortest = p.SongName
									longest = track.Title
								}
								if (p.SongName == track.Title || strings.Contains(longest, shortest)) &&
									(p.ArtistName == track.Artist || p.ArtistName == track.AlbumArtist) {
									p.AlbumName = track.Album
									break
								}
							}
						}
						if p.AlbumName == "" {
							notFound = append(notFound, p)
							continue
						}
						amp.Plays = append(amp.Plays, p)
					}
					if useMusicBrainz {
						_, err := queryMusicBrainzDatabase(notFound)
						if err != nil {
							log.Printf("error getting results from musicbrainz: %v", err)
						}
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("unable to process csv")
}

// Shamelessly ripped from SO: https://stackoverflow.com/a/56099384
func NewZipFromReader(file io.ReadCloser, size int64) (*zip.Reader, error) {
	in := file.(io.Reader)

	if _, ok := in.(io.ReaderAt); ok != true {
		buffer, err := ioutil.ReadAll(in)

		if err != nil {
			return nil, err
		}

		in = bytes.NewReader(buffer)
		size = int64(len(buffer))
	}

	reader, err := zip.NewReader(in.(io.ReaderAt), size)

	if err != nil {
		return nil, err
	}

	return reader, nil
}

// Enable sorting plays by timestamp in descending order (new plays at beginning of csv)
func (amp AppleMusic) Len() int { return len(amp.Plays) }
func (amp AppleMusic) Less(i, j int) bool {
	return amp.Plays[i].PlayTimestamp.UnixMilli() > amp.Plays[j].PlayTimestamp.UnixMilli()
}
func (amp AppleMusic) Swap(i, j int) { amp.Plays[i], amp.Plays[j] = amp.Plays[j], amp.Plays[i] }

type song struct {
	artist string
	album  string
	title  string
}

func queryMusicBrainzDatabase(plays []Play) ([]Play, error) {
	sqlStatement := `
SELECT
	A.name, R.name, T.name
FROM
	track T INNER JOIN
	medium M ON M.id = T.medium INNER JOIN
	release R ON R.id = M.release INNER JOIN
	artist A ON A.id = R.artist_credit
WHERE
	T.name ~* $1 AND
	A.name ~* $2
`

	err := db.Ping()
	if err != nil {
		return nil, err
	}

	// Iterate through 1000 results
	var iter int
	processed := make([]Play, 0, len(plays))
	for {
		bucket := make(map[string]song)
		artists, titles := playsToString(plays[iter : iter+1000])
		log.Println(artists, titles)

		rows, err := db.Query(sqlStatement, titles, artists)
		if err != nil {
			log.Println(err)
			break
		}

		defer rows.Close()

		var artist, album, title string
		for rows.Next() {
			err := rows.Scan(&artist, &album, &title)
			if err != nil {
				fmt.Println(err)
				break
			}

			bucket[artist] = song{artist: artist, album: album, title: title}
		}

		log.Print(bucket)
		break
	}
	return processed, nil
}

func playsToString(plays []Play) (string, string) {
	var artists, titles string

	for _, play := range plays {
		if artists == "" && titles == "" {
			artists += play.ArtistName
			titles += play.SongName
			continue
		}

		artists += fmt.Sprintf("|%s", play.ArtistName)
		titles += fmt.Sprintf("|%s", play.SongName)
	}

	for _, c := range "(){}*" {
		char := string(c)
		artists = strings.ReplaceAll(artists, char, "\\"+char)
		titles = strings.ReplaceAll(titles, char, "\\"+char)
	}

	return artists, titles
}
