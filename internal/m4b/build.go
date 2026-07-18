package m4b

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/chapters"
	"github.com/Tulvar/bookbind/internal/ffmpeg"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
}

type ProgressRunner interface {
	Runner
	SetProgressWriter(writer io.Writer)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	if err := cmd.Run(); err != nil {
		output := strings.TrimSpace(stderr.String())
		if output != "" {
			return fmt.Errorf("%w: %s", err, output)
		}
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffmpeg not found. Install ffmpeg, or put ffmpeg on PATH")
		}
		return err
	}
	return nil
}

type ReportingRunner struct {
	Writer io.Writer
}

func (r ReportingRunner) Run(ctx context.Context, name string, args ...string) error {
	writer := r.Writer
	if writer == nil {
		writer = io.Discard
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.MultiWriter(os.Stdout, writer)
	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, writer, &stderr)
	if err := cmd.Run(); err != nil {
		output := strings.TrimSpace(stderr.String())
		if output != "" {
			return fmt.Errorf("%w: %s", err, output)
		}
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffmpeg not found. Install ffmpeg, or put ffmpeg on PATH")
		}
		return err
	}
	return nil
}

type Builder struct {
	FFmpegPath string
	Runner     Runner
}

func NewBuilder(ffmpegPath string) *Builder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	ffmpegPath = ffmpeg.ResolveBinary(ffmpegPath)
	return &Builder{
		FFmpegPath: ffmpegPath,
		Runner:     ExecRunner{},
	}
}

type BuildRequest struct {
	Input          audio.Input
	Metadata       metadata.Book
	CoverPath      string
	OutputPath     string
	Overwrite      bool
	DryRun         bool
	ChapterEvery   time.Duration
	ProgressWriter io.Writer
}

type BuildResult struct {
	Command []string
	DryRun  bool
}

func (b *Builder) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
	if len(req.Input.Files) == 0 {
		return BuildResult{}, fmt.Errorf("no input files to convert")
	}

	command, cleanup, err := b.command(req)
	if err != nil {
		return BuildResult{}, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	result := BuildResult{
		Command: command,
		DryRun:  req.DryRun,
	}
	if req.DryRun {
		return result, nil
	}

	if b.Runner == nil {
		return BuildResult{}, fmt.Errorf("ffmpeg runner is not configured")
	}
	temporaryOutput, err := createTemporaryOutput(req.OutputPath)
	if err != nil {
		return BuildResult{}, err
	}
	defer func() {
		_ = os.Remove(temporaryOutput)
	}()

	executionCommand := append([]string(nil), command...)
	executionCommand[1] = "-y"
	executionCommand[len(executionCommand)-1] = temporaryOutput
	if runner, ok := b.Runner.(ProgressRunner); ok {
		runner.SetProgressWriter(req.ProgressWriter)
	}
	if err := b.Runner.Run(ctx, executionCommand[0], executionCommand[1:]...); err != nil {
		return BuildResult{}, fmt.Errorf("ffmpeg conversion failed: %w", err)
	}
	if err := publishTemporaryOutput(temporaryOutput, req.OutputPath, req.Overwrite); err != nil {
		return BuildResult{}, err
	}
	return result, nil
}

func (b *Builder) command(req BuildRequest) ([]string, func(), error) {
	args := []string{}
	if req.Overwrite {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}

	if len(req.Input.Files) == 1 {
		bookChapters, err := singleFileChapters(req.Input.Files[0], req.ChapterEvery)
		if err != nil {
			return nil, nil, err
		}
		metadataPath, cleanupMetadata, err := writeFFMetadata(req.Metadata, bookChapters)
		if err != nil {
			return nil, nil, err
		}
		cleanup := cleanupMetadata
		args = append(args,
			"-i", req.Input.Files[0].Path,
			"-i", metadataPath,
		)
		if req.CoverPath != "" {
			args = append(args,
				"-i", req.CoverPath,
			)
		}
		args = append(args,
			"-map", "0:a",
		)
		if req.CoverPath != "" {
			args = append(args, "-map", "2:v")
		}
		args = append(args,
			"-map_metadata", "1",
			"-c:a", "aac",
			"-b:a", "64k",
		)
		args = appendAudioLanguage(args, req.Metadata.Language)
		if len(bookChapters) > 0 {
			args = append(args, "-map_chapters", "1")
		}
		args = appendCoverArgs(args, req.CoverPath)
		args = append(args, "-movflags", "+faststart")
		args = append(args, req.OutputPath)
		return append([]string{b.FFmpegPath}, args...), cleanup, nil
	}

	bookChapters, err := chapters.FromAudioFiles(req.Input.Files)
	if err != nil {
		return nil, nil, err
	}
	metadataPath, cleanupMetadata, err := writeFFMetadata(req.Metadata, bookChapters)
	if err != nil {
		return nil, nil, err
	}
	cleanup := cleanupMetadata

	metadataInput := 1
	audioMap := "0:a"
	filterGraph := ""
	if concatDemuxerCompatible(req.Input.Files) {
		listPath, cleanupConcat, err := writeConcatList(req.Input.Files)
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		cleanup = joinCleanup(cleanup, cleanupConcat)
		args = append(args,
			"-f", "concat",
			"-safe", "0",
			"-i", listPath,
			"-i", metadataPath,
		)
	} else {
		for _, file := range req.Input.Files {
			args = append(args, "-i", file.Path)
		}
		metadataInput = len(req.Input.Files)
		audioMap = "[bookbind_audio]"
		filterGraph = audioConcatFilter(len(req.Input.Files))
		args = append(args, "-i", metadataPath)
	}

	coverInput := metadataInput + 1
	if req.CoverPath != "" {
		args = append(args,
			"-i", req.CoverPath,
		)
	}
	if filterGraph != "" {
		args = append(args, "-filter_complex", filterGraph)
	}
	args = append(args,
		"-map", audioMap,
	)
	if req.CoverPath != "" {
		args = append(args, "-map", strconv.Itoa(coverInput)+":v:0")
	}
	args = append(args,
		"-map_metadata", strconv.Itoa(metadataInput),
		"-map_chapters", strconv.Itoa(metadataInput),
		"-c:a", "aac",
		"-b:a", "64k",
	)
	args = appendAudioLanguage(args, req.Metadata.Language)
	args = appendCoverArgs(args, req.CoverPath)
	args = append(args, "-movflags", "+faststart")
	args = append(args, req.OutputPath)
	return append([]string{b.FFmpegPath}, args...), cleanup, nil
}

type concatAudioSignature struct {
	codec         string
	sampleRate    int
	sampleFormat  string
	channels      int
	channelLayout string
	timeBase      string
}

func concatDemuxerCompatible(files []audio.File) bool {
	if len(files) < 2 {
		return true
	}

	want, ok := concatSignature(files[0])
	if !ok {
		return false
	}
	for _, file := range files[1:] {
		got, ok := concatSignature(file)
		if !ok || got != want {
			return false
		}
	}
	return true
}

func concatSignature(file audio.File) (concatAudioSignature, bool) {
	if file.AudioStreams != 1 || file.NonAudioStreams != 0 ||
		strings.TrimSpace(file.Codec) == "" || file.SampleRate <= 0 ||
		strings.TrimSpace(file.SampleFormat) == "" || file.Channels <= 0 ||
		strings.TrimSpace(file.ChannelLayout) == "" || strings.TrimSpace(file.TimeBase) == "" {
		return concatAudioSignature{}, false
	}

	return concatAudioSignature{
		codec:         strings.ToLower(strings.TrimSpace(file.Codec)),
		sampleRate:    file.SampleRate,
		sampleFormat:  strings.ToLower(strings.TrimSpace(file.SampleFormat)),
		channels:      file.Channels,
		channelLayout: strings.ToLower(strings.TrimSpace(file.ChannelLayout)),
		timeBase:      strings.TrimSpace(file.TimeBase),
	}, true
}

func audioConcatFilter(fileCount int) string {
	var filter strings.Builder
	for index := 0; index < fileCount; index++ {
		filter.WriteString(fmt.Sprintf(
			"[%d:a:0]asetpts=PTS-STARTPTS[bookbind_a%d];",
			index,
			index,
		))
	}
	for index := 0; index < fileCount; index++ {
		filter.WriteString(fmt.Sprintf("[bookbind_a%d]", index))
	}
	filter.WriteString(fmt.Sprintf(
		"concat=n=%d:v=0:a=1[bookbind_audio]",
		fileCount,
	))
	return filter.String()
}

func singleFileChapters(file audio.File, chapterEvery time.Duration) ([]chapters.Chapter, error) {
	if chapterEvery == 0 {
		return embeddedChapters(file.Chapters), nil
	}
	return chapters.Synthetic(file.Duration, chapterEvery)
}

func embeddedChapters(values []audio.Chapter) []chapters.Chapter {
	if len(values) == 0 {
		return nil
	}
	result := make([]chapters.Chapter, 0, len(values))
	for _, value := range values {
		result = append(result, chapters.Chapter{
			Title: value.Title,
			Start: value.Start,
			End:   value.End,
		})
	}
	return result
}

func appendCoverArgs(args []string, coverPath string) []string {
	if coverPath == "" {
		return append(args, "-vn")
	}
	return append(args,
		"-c:v", "copy",
		"-disposition:v", "attached_pic",
	)
}

func appendAudioLanguage(args []string, language string) []string {
	code := mp4LanguageCode(language)
	if code == "" {
		return args
	}
	return append(args, "-metadata:s:a:0", "language="+code)
}

func mp4LanguageCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if separator := strings.IndexAny(value, "-_"); separator >= 0 {
		value = value[:separator]
	}
	if len(value) == 3 {
		return value
	}
	return map[string]string{
		"ar": "ara",
		"cs": "ces",
		"da": "dan",
		"de": "deu",
		"el": "ell",
		"en": "eng",
		"es": "spa",
		"fi": "fin",
		"fr": "fra",
		"he": "heb",
		"it": "ita",
		"ja": "jpn",
		"ko": "kor",
		"nl": "nld",
		"no": "nor",
		"pl": "pol",
		"pt": "por",
		"ru": "rus",
		"sk": "slk",
		"sv": "swe",
		"tr": "tur",
		"uk": "ukr",
		"zh": "zho",
	}[value]
}

func writeConcatList(files []audio.File) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "bookbind-concat-*")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	listPath := filepath.Join(tempDir, "input.txt")
	var builder strings.Builder
	for _, file := range files {
		builder.WriteString("file '")
		builder.WriteString(escapeConcatPath(file.Path))
		builder.WriteString("'\n")
	}
	if err := os.WriteFile(listPath, []byte(builder.String()), 0o600); err != nil {
		cleanup()
		return "", nil, err
	}

	return listPath, cleanup, nil
}

func escapeConcatPath(path string) string {
	return strings.ReplaceAll(path, "'", "'\\''")
}

func writeFFMetadata(book metadata.Book, bookChapters []chapters.Chapter) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "bookbind-metadata-*")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	metadataPath := filepath.Join(tempDir, "metadata.txt")
	if err := os.WriteFile(metadataPath, []byte(chapters.FFMetadataDocument(book, bookChapters)), 0o600); err != nil {
		cleanup()
		return "", nil, err
	}

	return metadataPath, cleanup, nil
}

func joinCleanup(cleanups ...func()) func() {
	return func() {
		for _, cleanup := range cleanups {
			if cleanup != nil {
				cleanup()
			}
		}
	}
}
