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
}

type ProbeResult struct {
	Duration time.Duration
	Codec    string
	Bitrate  int
	Channels int
}
