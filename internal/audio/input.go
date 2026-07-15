package audio

import "time"

type Input struct {
	Path      string
	Files     []File
	TotalTime time.Duration
}

type File struct {
	Path            string
	Name            string
	Duration        time.Duration
	Codec           string
	Bitrate         int
	SampleRate      int
	SampleFormat    string
	Channels        int
	ChannelLayout   string
	TimeBase        string
	AudioStreams    int
	NonAudioStreams int
	Tags            EmbeddedTags
	Chapters        []Chapter
}

type ProbeResult struct {
	Duration        time.Duration
	Codec           string
	Bitrate         int
	SampleRate      int
	SampleFormat    string
	Channels        int
	ChannelLayout   string
	TimeBase        string
	AudioStreams    int
	NonAudioStreams int
	Tags            EmbeddedTags
	Chapters        []Chapter
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

type Chapter struct {
	Title string
	Start time.Duration
	End   time.Duration
}
