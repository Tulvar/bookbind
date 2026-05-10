package m4b

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/chapters"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Builder struct {
	FFmpegPath string
	Runner     Runner
}

func NewBuilder(ffmpegPath string) *Builder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	return &Builder{
		FFmpegPath: ffmpegPath,
		Runner:     ExecRunner{},
	}
}

type BuildRequest struct {
	Input        audio.Input
	Metadata     metadata.Book
	CoverPath    string
	OutputPath   string
	Overwrite    bool
	DryRun       bool
	ChapterEvery time.Duration
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
	if err := b.Runner.Run(ctx, command[0], command[1:]...); err != nil {
		return BuildResult{}, fmt.Errorf("ffmpeg conversion failed: %w", err)
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
		if len(bookChapters) > 0 {
			args = append(args, "-map_chapters", "1")
		}
		args = appendCoverArgs(args, req.CoverPath)
		args = append(args, req.OutputPath)
		return append([]string{b.FFmpegPath}, args...), cleanup, nil
	}

	listPath, cleanupConcat, err := writeConcatList(req.Input.Files)
	if err != nil {
		return nil, nil, err
	}
	cleanup := cleanupConcat

	bookChapters, err := chapters.FromAudioFiles(req.Input.Files)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	metadataPath, cleanupMetadata, err := writeFFMetadata(req.Metadata, bookChapters)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	cleanup = joinCleanup(cleanup, cleanupMetadata)

	args = append(args,
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
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
		"-map_chapters", "1",
		"-c:a", "aac",
		"-b:a", "64k",
	)
	args = appendCoverArgs(args, req.CoverPath)
	args = append(args, req.OutputPath)
	return append([]string{b.FFmpegPath}, args...), cleanup, nil
}

func singleFileChapters(file audio.File, chapterEvery time.Duration) ([]chapters.Chapter, error) {
	if chapterEvery == 0 {
		return nil, nil
	}
	return chapters.Synthetic(file.Duration, chapterEvery)
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
