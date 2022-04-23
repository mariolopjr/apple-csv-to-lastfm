package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jszwec/csvutil"
	"github.com/mariolopjr/apple-csv-to-lastfm/decoder"
	"github.com/mariolopjr/apple-csv-to-lastfm/encoder"
)

var (
	inputFilePath  string
	outputFilePath string
	useMusicBrainz bool
)

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	defaultInputFile := fmt.Sprintf("%s/Apple Media Services information.zip", currentDir)
	defaultOutputFile := fmt.Sprintf("%s/applemusic-lastfm.csv", currentDir)

	flag.StringVar(&inputFilePath, "apple-music-export-path", defaultInputFile, "The absolute path (location) of a Apple Music .zip data export file")
	flag.StringVar(&outputFilePath, "lastfm-output-path", defaultOutputFile, "The absolute path to where you want to export the CSV to import into Maloja")
	//flag.BoolVar(&useMusicBrainz, "use-musicbrainz", false, "Lookup unmatched tracks using MusicBrainz DB (follow instructions in README)")
}

func main() {
	flag.Parse()

	if inputFilePath == "" || !filepath.IsAbs(inputFilePath) || !strings.HasSuffix(inputFilePath, ".zip") {
		log.Fatalln("invalid file path; pass '--apple-music-export-path=/absolute/path/to/Apple Media Services information.zip'")
	}

	// Decode Apple Music Export
	var appleMusic decoder.AppleMusic
	err := appleMusic.Unmarshal(inputFilePath, useMusicBrainz)
	if err != nil {
		log.Fatal(err)
	}

	// Sort Apple Music Plays
	sort.Sort(appleMusic)

	// Encode to LastFM CSV Export
	var lastFm encoder.LastFm
	err = lastFm.Marshal(appleMusic)

	// Write to CSV
	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)
	enc.AutoHeader = false

	for _, s := range lastFm.Scrobbles {
		if err := enc.Encode(s); err != nil {
			log.Fatalf("unable to encode scrobble [%v]: %v", s, err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Printf("error marshalling scrobbles: %s", err)
	}

	err = os.WriteFile(outputFilePath, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error creating csv: %s", err)
	}
	log.Printf("Number of Apple Music plays converted: %d", len(lastFm.Scrobbles))
}
