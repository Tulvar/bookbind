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
	cmd := exec.CommandContext(ctx, p.Path, ffprobeArgs(path)...)

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

func ffprobeArgs(path string) []string {
	return []string{
		"-v", "error",
		"-print_format", "json",
		"-count_packets",
		"-show_chapters",
		"-show_format",
		"-show_streams",
		path,
	}
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
	CodecType     string             `json:"codec_type"`
	CodecName     string             `json:"codec_name"`
	Duration      string             `json:"duration"`
	BitRate       string             `json:"bit_rate"`
	SampleRate    string             `json:"sample_rate"`
	SampleFormat  string             `json:"sample_fmt"`
	Channels      int                `json:"channels"`
	ChannelLayout string             `json:"channel_layout"`
	TimeBase      string             `json:"time_base"`
	ReadPackets   string             `json:"nb_read_packets"`
	Disposition   ffprobeDisposition `json:"disposition"`
}

type ffprobeDisposition struct {
	AttachedPic int `json:"attached_pic"`
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
	duration = accurateMP3Duration(duration, audioStream)

	bitrate := parseInt(o.Format.BitRate)
	if bitrate == 0 {
		bitrate = parseInt(audioStream.BitRate)
	}

	audioStreams := 0
	hasAttachedPicture := false
	attachedPictureStream := 0
	videoStream := 0
	for _, stream := range o.Streams {
		if stream.CodecType == "audio" {
			audioStreams++
		}
		if stream.CodecType == "video" {
			if !hasAttachedPicture && stream.Disposition.AttachedPic != 0 {
				hasAttachedPicture = true
				attachedPictureStream = videoStream
			}
			videoStream++
		}
	}

	return ProbeResult{
		Duration:              duration,
		Codec:                 audioStream.CodecName,
		Bitrate:               bitrate,
		SampleRate:            parseInt(audioStream.SampleRate),
		SampleFormat:          audioStream.SampleFormat,
		Channels:              audioStream.Channels,
		ChannelLayout:         audioStream.ChannelLayout,
		TimeBase:              audioStream.TimeBase,
		AudioStreams:          audioStreams,
		NonAudioStreams:       len(o.Streams) - audioStreams,
		HasAttachedPicture:    hasAttachedPicture,
		AttachedPictureStream: attachedPictureStream,
		Tags:                  embeddedTags(o.Format.Tags),
		Chapters:              chapters(o.Chapters),
	}
}

func accurateMP3Duration(reported time.Duration, stream ffprobeStream) time.Duration {
	if !strings.EqualFold(stream.CodecName, "mp3") {
		return reported
	}

	sampleRate := parseInt(stream.SampleRate)
	packetCount := parseInt64(stream.ReadPackets)
	samplesPerPacket := mp3SamplesPerPacket(sampleRate)
	if packetCount <= 0 || samplesPerPacket == 0 {
		return reported
	}
	if packetCount > int64(^uint64(0)>>1)/samplesPerPacket {
		return reported
	}

	counted := sampleDuration(packetCount*samplesPerPacket, sampleRate)
	if reported <= 0 {
		return counted
	}

	// Xing/LAME gapless metadata can trim encoder delay and padding from the
	// packet duration. Keep that reported duration when the difference is no
	// larger than two MP3 frames; larger differences indicate a bitrate-based
	// estimate, which drifts at chapter boundaries.
	frameDuration := sampleDuration(samplesPerPacket, sampleRate)
	if absDuration(counted-reported) <= 2*frameDuration {
		return reported
	}
	return counted
}

func mp3SamplesPerPacket(sampleRate int) int64 {
	if sampleRate <= 0 {
		return 0
	}
	if sampleRate <= 28000 {
		return 576
	}
	return 1152
}

func sampleDuration(samples int64, sampleRate int) time.Duration {
	if samples <= 0 || sampleRate <= 0 {
		return 0
	}
	seconds := samples / int64(sampleRate)
	remainder := samples % int64(sampleRate)
	return time.Duration(seconds)*time.Second +
		time.Duration(remainder)*time.Second/time.Duration(sampleRate)
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
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

func parseInt64(value string) int64 {
	if value == "" || value == "N/A" {
		return 0
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
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
