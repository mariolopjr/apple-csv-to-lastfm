# Apple Music Data Export to LastFM CSV
Small go program that takes an Apple Media Services information data export and extracts the Apple Music play activity to convert to the LastFM CSV format for easy import into Maloja.

To run, git clone this repo and then run the command `go run main.go` (you'll need Go installed).

FIXME: Add detailed instructions here

## Additional Matching
WARNING: Not Implemented Fully (CLI flag is commented out to prevent use)
Unfortunately, the Apple Music play activity only has Artist and Title information. I do try to get Album information from the Library JSON file in the data export, but this wouldn't work for music not 'added' to the Library. To work around that, you can set up a local musicbrainz database server in Docker following these directions: https://github.com/metabrainz/musicbrainz-docker#installation. You'll need to include a `docker-compose.override.yml` in the same directory to expose the postgres port, use the following:
```

```


Once the containers are running, run the export with the following command `go run main.go --use-musicbrainz`.
