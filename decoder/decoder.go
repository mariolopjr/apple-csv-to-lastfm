package decoder

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
)

func (amp *AppleMusic) Unmarshal(srcFilePath string) error {
	zipReader, err := zip.OpenReader(srcFilePath)
	if err != nil {
		return err
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "Apple_Media_Services.zip") {
			zipReader, err := file.Open()
			if err != nil {
				return err
			}

			mediaZip, err := NewZipFromReader(zipReader, file.FileInfo().Size())
			if err != nil {
				return err
			}

			var libraryCount int
			for _, mediaFile := range mediaZip.File {
				// Build list of library tracks
				if strings.Contains(mediaFile.Name, "Apple Music Library Tracks.json.zip") {
					libraryReader, err := mediaFile.Open()
					if err != nil {
						return err
					}

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
										libraryCount++
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
							continue
						}
						amp.Plays = append(amp.Plays, p)
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
