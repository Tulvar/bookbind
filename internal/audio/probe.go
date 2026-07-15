package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/ffmpeg"
)

type Prober interface {
	Probe(ctx context.Context, path string) (ProbeResult, error)
}

type FFProbe struct {
	Path string
}

func NewFFProbe(path string) *FFProbe {
	if path == "" {
		path = "ffprobe"
	}
	path = ffmpeg.ResolveBinary(path)
	return &FFProbe{Path: path}
}

func (p *FFProbe) Probe(ctx context.Context, path string) (ProbeResult, error) {
	cmd := exec.CommandContext(ctx, p.Path,
		"-v", "error",
		"-print_format", "json",
		"-show_chapters",
		"-show_format",
		"-show_streams",
		path,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return ProbeResult{}, fmt.Errorf("ffprobe %s: %s", path, stderr.String())
		}
		if isExecutableNotFound(err) {
			return ProbeResult{}, fmt.Errorf("ffprobe %s: ffprobe not found. Install ffmpeg, or put ffprobe on PATH", path)
		}
		return ProbeResult{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}

	var data ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		return ProbeResult{}, fmt.Errorf("parse ffprobe output for %s: %w", path, err)
	}

	return data.probeResult(), nil
}

func isExecutableNotFound(err error) bool {
	return errors.Is(err, exec.ErrNotFound)
}

type ffprobeOutput struct {
	Streams  []ffprobeStream  `json:"streams"`
	Format   ffprobeFormat    `json:"format"`
	Chapters []ffprobeChapter `json:"chapters"`
}

type ffprobeStream struct {
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	Duration      string `json:"duration"`
	BitRate       string `json:"bit_rate"`
	SampleRate    string `json:"sample_rate"`
	SampleFormat  string `json:"sample_fmt"`
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
	TimeBase      string `json:"time_base"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
	Tags     ffTags `json:"tags"`
}

type ffTags map[string]string

type ffprobeChapter struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Tags      ffTags `json:"tags"`
}

func (o ffprobeOutput) firstAudioStream() ffprobeStream {
	for _, stream := range o.Streams {
		if stream.CodecType == "audio" {
			return stream
		}
	}
	return ffprobeStream{}
}

func (o ffprobeOutput) probeResult() ProbeResult {
	audioStream := o.firstAudioStream()
	duration := parseDuration(o.Format.Duration)
	if duration == 0 {
		duration = parseDuration(audioStream.Duration)
	}

	bitrate := parseInt(o.Format.BitRate)
	if bitrate == 0 {
		bitrate = parseInt(audioStream.BitRate)
	}

	audioStreams := 0
	for _, stream := range o.Streams {
		if stream.CodecType == "audio" {
			audioStreams++
		}
	}

	return ProbeResult{
		Duration:        duration,
		Codec:           audioStream.CodecName,
		Bitrate:         bitrate,
		SampleRate:      parseInt(audioStream.SampleRate),
		SampleFormat:    audioStream.SampleFormat,
		Channels:        audioStream.Channels,
		ChannelLayout:   audioStream.ChannelLayout,
		TimeBase:        audioStream.TimeBase,
		AudioStreams:    audioStreams,
		NonAudioStreams: len(o.Streams) - audioStreams,
		Tags:            embeddedTags(o.Format.Tags),
		Chapters:        chapters(o.Chapters),
	}
}

func parseDuration(value string) time.Duration {
	if value == "" || value == "N/A" {
		return 0
	}

	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func parseInt(value string) int {
	if value == "" || value == "N/A" {
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func embeddedTags(tags ffTags) EmbeddedTags {
	return EmbeddedTags{
		Title:       tagValue(tags, "title"),
		Artist:      tagValue(tags, "artist"),
		Album:       tagValue(tags, "album"),
		AlbumArtist: tagValue(tags, "album_artist", "albumartist", "album artist"),
		Composer:    tagValue(tags, "composer"),
		Genre:       tagValue(tags, "genre"),
		Date:        tagValue(tags, "date", "year"),
		Comment:     tagValue(tags, "comment", "description"),
		Language:    tagValue(tags, "language"),
	}
}

func tagValue(tags ffTags, names ...string) string {
	for _, name := range names {
		for key, value := range tags {
			if strings.EqualFold(key, name) {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func chapters(values []ffprobeChapter) []Chapter {
	result := make([]Chapter, 0, len(values))
	for i, value := range values {
		title := tagValue(value.Tags, "title")
		if title == "" {
			title = fmt.Sprintf("Chapter %03d", i+1)
		}
		result = append(result, Chapter{
			Title: title,
			Start: parseDuration(value.StartTime),
			End:   parseDuration(value.EndTime),
		})
	}
	return result
}
