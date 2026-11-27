package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/m4b"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type ConvertRequest struct {
	InputPath    string
	OutputPath   string
	MetadataPath string
	Metadata     metadata.Book
	// MetadataOverride marks Metadata as final values reviewed by the user.
	// Non-empty override fields replace automatically collected metadata.
	MetadataOverride bool
	CoverPath        string
	ChapterEvery     string
	DryRun           bool
	Overwrite        bool
	Progress         io.Writer
}

type ConvertResult struct {
	Input        audio.Input
	Metadata     metadata.Book
	MetadataPath string
	CoverPath    string
	ChapterEvery string
	OutputPath   string
	DryRun       bool
	Command      []string
}

func (a *App) Convert(ctx context.Context, req ConvertRequest) (ConvertResult, error) {
	input, err := a.inspectMP3Input(ctx, req.InputPath)
	if err != nil {
		return ConvertResult{}, err
	}

	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = defaultOutputPath(input.Path)
	}
	if !strings.EqualFold(filepath.Ext(outputPath), ".m4b") {
		return ConvertResult{}, fmt.Errorf("output path must end with .m4b")
	}
	if err := ensureOutputWritable(outputPath, req.Overwrite); err != nil {
		return ConvertResult{}, err
	}

	if a.builder == nil {
		return ConvertResult{}, fmt.Errorf("m4b builder is not configured")
	}

	book, err := a.prepareMetadata(input, req.MetadataPath, req.Metadata)
	if err != nil {
		return ConvertResult{}, err
	}
	if req.MetadataOverride {
		book = overrideBook(book, req.Metadata)
	}
	coverCandidate := strings.TrimSpace(req.CoverPath)
	if coverCandidate == "" && strings.TrimSpace(book.Cover) == "" {
		coverCandidate = localCoverPath(input.Path)
	}
	coverPath, err := resolveCoverPath(ctx, coverCandidate, req.MetadataPath, book, req.DryRun)
	if err != nil {
		return ConvertResult{}, err
	}
	chapterEvery, err := parseChapterInterval(req.ChapterEvery)
	if err != nil {
		return ConvertResult{}, err
	}

	buildCover := coverPath
	if req.DryRun && isRemoteURL(coverPath) {
		buildCover = ""
	}
	build, err := a.builder.Build(ctx, m4b.BuildRequest{
		Input:          input,
		Metadata:       book,
		CoverPath:      buildCover,
		OutputPath:     outputPath,
		Overwrite:      req.Overwrite,
		DryRun:         req.DryRun,
		ChapterEvery:   chapterEvery,
		ProgressWriter: req.Progress,
	})
	if err != nil {
		return ConvertResult{}, err
	}

	return ConvertResult{
		Input:        input,
		Metadata:     book,
		MetadataPath: req.MetadataPath,
		CoverPath:    coverPath,
		ChapterEvery: req.ChapterEvery,
		OutputPath:   outputPath,
		DryRun:       req.DryRun,
		Command:      build.Command,
	}, nil
}

func localCoverPath(inputPath string) string {
	dir := metadataTemplateDir(inputPath)
	name := defaultCover(dir)
	if name == "" {
		return ""
	}
	return filepath.Join(dir, name)
}

func loadMetadata(path string) (metadata.Book, error) {
	if strings.TrimSpace(path) == "" {
		return metadata.Book{}, nil
	}
	return metadata.LoadYAML(path)
}

var cacheDir = CacheDir

func resolveCoverPath(ctx context.Context, cliCoverPath, metadataPath string, book metadata.Book, dryRun bool) (string, error) {
	coverPath := strings.TrimSpace(cliCoverPath)
	if coverPath == "" {
		coverPath = strings.TrimSpace(book.Cover)
		if isRemoteURL(coverPath) {
			if dryRun {
				return coverPath, nil
			}
			return downloadCover(ctx, coverPath)
		}
		if coverPath != "" && metadataPath != "" && !filepath.IsAbs(coverPath) {
			coverPath = filepath.Join(filepath.Dir(metadataPath), coverPath)
		}
	}
	if coverPath == "" {
		return "", nil
	}

	info, err := os.Stat(coverPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("cover path is a directory: %s", coverPath)
	}

	switch strings.ToLower(filepath.Ext(coverPath)) {
	case ".jpg", ".jpeg", ".png":
		return coverPath, nil
	default:
		return "", fmt.Errorf("cover must be .jpg, .jpeg, or .png: %s", coverPath)
	}
}

const (
	maxCoverBytes  = 10 * 1024 * 1024
	maxCoverPixels = 16_000_000
)

var coverClient = &http.Client{Timeout: 15 * time.Second}

func downloadCover(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("cover url: %w", err)
	}
	resp, err := coverClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download cover: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download cover: unexpected status %s", resp.Status)
	}
	if resp.ContentLength > maxCoverBytes {
		return "", fmt.Errorf("cover exceeds the %d MiB limit", maxCoverBytes/(1024*1024))
	}

	cacheRoot, err := cacheDir()
	if err != nil {
		return "", err
	}
	coverDir := filepath.Join(cacheRoot, "covers")
	if err := os.MkdirAll(coverDir, 0o755); err != nil {
		return "", fmt.Errorf("create covers cache: %w", err)
	}
	file, err := os.CreateTemp(coverDir, ".download-*")
	if err != nil {
		return "", fmt.Errorf("create temporary cover: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)

	ext, err := writeCover(file, resp.Body)
	closeErr := file.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", fmt.Errorf("close temporary cover: %w", closeErr)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(rawURL))
	path := filepath.Join(coverDir, hex.EncodeToString(sum[:])+ext)
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", fmt.Errorf("publish cover: %w", err)
	}
	return path, nil
}

func writeCover(file *os.File, body io.Reader) (string, error) {
	size, err := io.Copy(file, io.LimitReader(body, maxCoverBytes+1))
	if err != nil {
		return "", fmt.Errorf("save cover: %w", err)
	}
	if size > maxCoverBytes {
		return "", fmt.Errorf("cover exceeds the %d MiB limit", maxCoverBytes/(1024*1024))
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("read temporary cover: %w", err)
	}
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", fmt.Errorf("invalid cover image: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > maxCoverPixels/config.Height {
		return "", fmt.Errorf("cover exceeds the %d pixel limit", maxCoverPixels)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("read temporary cover: %w", err)
	}
	if _, _, err := image.Decode(file); err != nil {
		return "", fmt.Errorf("invalid cover image: %w", err)
	}
	switch format {
	case "png":
		return ".png", nil
	case "jpeg":
		return ".jpg", nil
	default:
		return "", fmt.Errorf("cover must be a JPEG or PNG image")
	}
}

func isRemoteURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func ensureOutputWritable(outputPath string, overwrite bool) error {
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if _, err := os.Stat(outputPath); err == nil {
		if !overwrite {
			return fmt.Errorf("output already exists: %s", outputPath)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	outputDir := filepath.Dir(outputPath)
	if _, err := os.Stat(outputDir); err != nil {
		return fmt.Errorf("output directory is not available: %w", err)
	}
	return nil
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath + ".m4b"
	}
	return strings.TrimSuffix(inputPath, ext) + ".m4b"
}
