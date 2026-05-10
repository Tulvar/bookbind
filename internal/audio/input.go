package audio

import "time"

type Input struct {
	Path      string
	Files     []File
	TotalTime time.Duration
}

type File struct {
	Path     string
	Name     string
	Duration time.Duration
	Codec    string
	Bitrate  int
	Channels int
	Tags     EmbeddedTags
}

type ProbeResult struct {
	Duration time.Duration
	Codec    string
	Bitrate  int
	Channels int
	Tags     EmbeddedTags
}

type EmbeddedTags struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Composer    string
	Genre       string
	Date        string
	Comment     string
	Language    string
}
