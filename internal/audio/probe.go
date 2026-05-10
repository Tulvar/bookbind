package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	return &FFProbe{Path: path}
}

func (p *FFProbe) Probe(ctx context.Context, path string) (ProbeResult, error) {
	cmd := exec.CommandContext(ctx, p.Path,
		"-v", "error",
		"-print_format", "json",
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
		return ProbeResult{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}

	var data ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		return ProbeResult{}, fmt.Errorf("parse ffprobe output for %s: %w", path, err)
	}

	audioStream := data.firstAudioStream()
	duration := parseDuration(data.Format.Duration)
	if duration == 0 {
		duration = parseDuration(audioStream.Duration)
	}

	bitrate := parseInt(data.Format.BitRate)
	if bitrate == 0 {
		bitrate = parseInt(audioStream.BitRate)
	}

	return ProbeResult{
		Duration: duration,
		Codec:    audioStream.CodecName,
		Bitrate:  bitrate,
		Channels: audioStream.Channels,
		Tags:     embeddedTags(data.Format.Tags),
	}, nil
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Duration  string `json:"duration"`
	BitRate   string `json:"bit_rate"`
	Channels  int    `json:"channels"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
	Tags     ffTags `json:"tags"`
}

type ffTags map[string]string

func (o ffprobeOutput) firstAudioStream() ffprobeStream {
	for _, stream := range o.Streams {
		if stream.CodecType == "audio" {
			return stream
		}
	}
	return ffprobeStream{}
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
