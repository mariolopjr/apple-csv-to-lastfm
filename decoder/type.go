package decoder

import "time"

type AppleMusic struct {
	Plays  []Play
	Tracks map[string][]Track
}

type Play struct {
	AppleIdNumber                int    `csv:"Apple ID Number"`
	AppleMusicSubscription       bool   `csv:"Apple Music Subscription,omitempty"`
	ArtistName                   string `csv:"Artist Name"`
	BuildVersion                 string `csv:"Build Version"`
	ClientIpAddress              string `csv:"Client IP Address"`
	DeviceIdentifier             string `csv:"Device Identifier"`
	EndPositionInMilliseconds    int64  `csv:"End Position In Milliseconds,omitempty"`
	EndReasonType                string `csv:"End Reason Type"`
	EventEndTimestamp            string `csv:"Event End Timestamp"`
	EventReasonHintType          string `csv:"Event Reason Hint Type"`
	EventReceivedTimestamp       string `csv:"Event Received Timestamp"`
	EventStartTimestamp          string `csv:"Event Start Timestamp"`
	EventType                    string `csv:"Event Type"`
	FeatureName                  string `csv:"Feature Name"`
	ItemType                     string `csv:"Item Type"`
	MediaDurationInMilliseconds  int64  `csv:"Media Duration In Milliseconds,omitempty"`
	MediaType                    string `csv:"Media Type"`
	MetricsBucketId              int    `csv:"Metrics Bucket Id,omitempty"`
	MetricsClientId              string `csv:"Metrics Client Id"`
	MillisecondsSincePlay        int64  `csv:"Milliseconds Since Play"`
	Offline                      bool   `csv:"Offline,omitempty"`
	PlayDurationMilliseconds     int64  `csv:"Play Duration Milliseconds,omitempty"`
	ProvidedAudioBitDepth        int    `csv:"Provided Audio Bit Depth,omitempty"`
	ProvidedAudioChannel         string `csv:"Provided Audio Channel"`
	ProvidedAudioSampleRate      int    `csv:"Provided Audio Sample Rate,omitempty"`
	ProvidedBitRate              int    `csv:"Provided Bit Rate,omitempty"`
	ProvidedCodec                string `csv:"Provided Codec"`
	ProvidedPlaybackFormat       string `csv:"Provided Playback Format"`
	SessionIsShared              bool   `csv:"Session Is Shared,omitempty"`
	SharedActivityDevicesCurrent string `csv:"Shared Activity Devices-Current"`
	SharedActivityDevicesMax     string `csv:"Shared Activity Devices-Max"`
	SongName                     string `csv:"Song Name"`
	SourceType                   string `csv:"Source Type"`
	StartPositionInMilliseconds  int64  `csv:"Start Position In Milliseconds"`
	StoreFrontName               string `csv:"Store Front Name"`
	UsersAudioQuality            string `csv:"User’s Audio Quality"`
	UsersPlaybackFormat          string `csv:"User’s Playback Format"`
	UTCOffsetInSeconds           int    `csv:"UTC Offset In Seconds,omitempty"`
	AlbumName                    string
	PlayTimestamp                time.Time
}

type Track struct {
	ContentType                   string `json:"Content Type"`
	TrackIdentifier               int    `json:"Track Identifier"`
	Title                         string `json:"Title"`
	SortName                      string `json:"Sort Name"`
	Artist                        string `json:"Artist"`
	SortArtist                    string `json:"Sort Artist"`
	Composer                      string `json:"Composer"`
	IsCompilation                 bool   `json:"Is Part of Compilation"`
	Album                         string `json:"Album"`
	SortAlbum                     string `json:"Sort Album"`
	AlbumArtist                   string `json:"Album Artist"`
	Genre                         string `json:"Genre"`
	TrackYear                     int    `json:"Track Year"`
	TrackNumberOnAlbum            int    `json:"Track Number On Album"`
	TrackCountOnAlbum             int    `json:"Track Count On Album"`
	DiscNumberOfAlbum             int    `json:"Disc Number Of Album"`
	DiscCountOfAlbum              int    `json:"Disc Count Of Album"`
	TrackDuration                 int64  `json:"Track Duration"`
	TrackPlayCount                int    `json:"Track Play Count"`
	DateAddedToLibrary            string `json:"Date Added To Library"`
	DateAddedToiCloudMusicLibrary string `json:"Date Added To iCloud Music Library"`
	LastModifiedDate              string `json:"Last Modified Date"`
	SkipCount                     int    `json:"Skip Count"`
	IsPurchased                   bool   `json:"Is Purchased"`
	AudioFileExtension            string `json:"Audio File Extension"`
	TrackLikeRating               string `json:"Track Like Rating"`
	IsChecked                     bool   `json:"Is Checked"`
	Copyright                     string `json:"Copyright"`
	PlaylistOnlyTrack             bool   `json:"Playlist Only Track"`
	ReleaseDate                   string `json:"Release Date"`
	PurchasedTrackIdentifier      int    `json:"Purchased Track Identifier"`
	AppleMusicTrackIdentifier     int    `json:"Apple Music Track Identifier"`
}
