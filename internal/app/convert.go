package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/m4b"
	"github.com/Tulvar/bookbind/internal/metadata"
)

type ConvertRequest struct {
	InputPath    string
	OutputPath   string
	MetadataPath string
	Metadata     metadata.Book
	CoverPath    string
	ChapterEvery string
	DryRun       bool
	Overwrite    bool
	Progress     io.Writer
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
	input, err := a.inspector.Inspect(ctx, req.InputPath)
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
	coverPath, err := resolveCoverPath(ctx, req.CoverPath, req.MetadataPath, book)
	if err != nil {
		return ConvertResult{}, err
	}
	chapterEvery, err := parseChapterInterval(req.ChapterEvery)
	if err != nil {
		return ConvertResult{}, err
	}

	build, err := a.builder.Build(ctx, m4b.BuildRequest{
		Input:          input,
		Metadata:       book,
		CoverPath:      coverPath,
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

func loadMetadata(path string) (metadata.Book, error) {
	if strings.TrimSpace(path) == "" {
		return metadata.Book{}, nil
	}
	return metadata.LoadYAML(path)
}

var cacheDir = CacheDir

func resolveCoverPath(ctx context.Context, cliCoverPath, metadataPath string, book metadata.Book) (string, error) {
	coverPath := strings.TrimSpace(cliCoverPath)
	if coverPath == "" {
		coverPath = strings.TrimSpace(book.Cover)
		if isRemoteURL(coverPath) {
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

func downloadCover(ctx context.Context, rawURL string) (string, error) {
	cacheRoot, err := cacheDir()
	if err != nil {
		return "", err
	}
	coverDir := filepath.Join(cacheRoot, "covers")
	if err := os.MkdirAll(coverDir, 0o755); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("cover url: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download cover: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download cover: unexpected status %s", resp.Status)
	}

	path := filepath.Join(coverDir, cachedCoverName(rawURL, resp.Header.Get("Content-Type")))
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("save cover: %w", err)
	}
	return path, nil
}

func cachedCoverName(rawURL, contentType string) string {
	sum := sha256.Sum256([]byte(rawURL))
	ext := strings.ToLower(filepath.Ext(parsedURLPath(rawURL)))
	switch ext {
	case ".jpg", ".jpeg", ".png":
	default:
		if strings.Contains(strings.ToLower(contentType), "png") {
			ext = ".png"
		} else {
			ext = ".jpg"
		}
	}
	return hex.EncodeToString(sum[:]) + ext
}

func parsedURLPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Path
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
