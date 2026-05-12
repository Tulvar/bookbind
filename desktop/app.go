package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	coreapp "github.com/Tulvar/bookbind/internal/app"
	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
	"github.com/Tulvar/bookbind/internal/providers"
	"github.com/Tulvar/bookbind/pkg/version"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx  context.Context
	core *coreapp.App
}

func NewApp() *App {
	return &App{
		core: coreapp.New(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) AppVersion() string {
	return version.Version
}

func (a *App) AvailableProviders() []coreapp.ProviderInfo {
	return coreapp.AvailableProviders()
}

func (a *App) SelectAudioFile() (string, error) {
	return wailsruntime.OpenFileDialog(a.dialogContext(), wailsruntime.OpenDialogOptions{
		Title: "Select audiobook MP3",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "MP3 audio", Pattern: "*.mp3"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) SelectAudioDirectory() (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.dialogContext(), wailsruntime.OpenDialogOptions{
		Title: "Select audiobook folder",
	})
}

func (a *App) SelectCacheDirectory() (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.dialogContext(), wailsruntime.OpenDialogOptions{
		Title: "Select cache folder",
	})
}

func (a *App) SelectMetadataFile() (string, error) {
	return wailsruntime.OpenFileDialog(a.dialogContext(), wailsruntime.OpenDialogOptions{
		Title: "Select metadata YAML",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "YAML metadata", Pattern: "*.yaml;*.yml"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) SelectCoverFile() (string, error) {
	return wailsruntime.OpenFileDialog(a.dialogContext(), wailsruntime.OpenDialogOptions{
		Title: "Select cover image",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Cover image", Pattern: "*.jpg;*.jpeg;*.png"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) SelectOutputFile() (string, error) {
	return wailsruntime.SaveFileDialog(a.dialogContext(), wailsruntime.SaveDialogOptions{
		Title:           "Select output M4B",
		DefaultFilename: "book.m4b",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "M4B audiobook", Pattern: "*.m4b"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) dialogContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

type InspectView struct {
	Path      string
	Files     []InspectFileView
	TotalTime string
}

type InspectFileView struct {
	Path     string
	Name     string
	Duration string
	Codec    string
	Bitrate  string
	Channels int
	Tags     audio.EmbeddedTags
	Chapters int
}

func (a *App) InspectPath(path string) (InspectView, error) {
	result, err := a.core.InspectInput(a.ctx, coreapp.InspectRequest{InputPath: path})
	if err != nil {
		return InspectView{}, err
	}

	files := make([]InspectFileView, 0, len(result.Input.Files))
	var total time.Duration
	for _, file := range result.Input.Files {
		total += file.Duration
		files = append(files, InspectFileView{
			Path:     file.Path,
			Name:     file.Name,
			Duration: formatDuration(file.Duration),
			Codec:    file.Codec,
			Bitrate:  formatBitrate(file.Bitrate),
			Channels: file.Channels,
			Tags:     file.Tags,
			Chapters: len(file.Chapters),
		})
	}

	return InspectView{
		Path:      result.Input.Path,
		Files:     files,
		TotalTime: formatDuration(total),
	}, nil
}

type MetadataSearchView struct {
	Candidates []MetadataCandidateView
}

type MetadataCandidateView struct {
	Provider    string
	ID          string
	Title       string
	Authors     []string
	Narrators   []string
	Series      string
	SeriesIndex string
	Year        int
	Duration    string
	CoverURL    string
	Confidence  string
}

type BookMetadataView struct {
	Title         string
	Subtitle      string
	Authors       []string
	Author        string
	Narrators     []string
	Narrator      string
	Series        string
	SeriesIndex   string
	Language      string
	Genre         string
	Description   string
	Publisher     string
	PublishedYear int
	Cover         string
}

type MetadataPreviewView struct {
	Candidate MetadataCandidateView
	Book      BookMetadataView
}

type MetadataResolveView struct {
	OutputPath string
	Candidate  MetadataCandidateView
	Book       BookMetadataView
}

func (a *App) SearchMetadata(title, author string, providerNames []string) (MetadataSearchView, error) {
	result, err := a.core.SearchMetadata(a.dialogContext(), coreapp.SearchRequest{
		Title:     strings.TrimSpace(title),
		Author:    strings.TrimSpace(author),
		Providers: providerNames,
	})
	if err != nil {
		return MetadataSearchView{}, err
	}

	candidates := make([]MetadataCandidateView, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		candidates = append(candidates, candidateView(candidate))
	}
	return MetadataSearchView{Candidates: candidates}, nil
}

func (a *App) PreviewMetadata(provider, id string) (MetadataPreviewView, error) {
	result, err := a.core.PreviewMetadata(a.dialogContext(), coreapp.PreviewMetadataRequest{
		Provider: provider,
		ID:       id,
	})
	if err != nil {
		return MetadataPreviewView{}, err
	}

	return MetadataPreviewView{
		Candidate: candidateView(result.Candidate),
		Book:      bookView(result.Book),
	}, nil
}

func (a *App) ResolveMetadata(provider, id, outputPath string, overwrite bool) (MetadataResolveView, error) {
	result, err := a.core.ResolveMetadata(a.dialogContext(), coreapp.ResolveMetadataRequest{
		Provider:   provider,
		ID:         id,
		OutputPath: outputPath,
		Overwrite:  overwrite,
	})
	if err != nil {
		return MetadataResolveView{}, err
	}

	return MetadataResolveView{
		OutputPath: result.OutputPath,
		Candidate:  candidateView(result.Candidate),
		Book:       bookView(result.Book),
	}, nil
}

type CacheEntryView struct {
	Path string
	Name string
	Kind string
	Size string
}

type CacheListView struct {
	Path    string
	Entries []CacheEntryView
	Size    string
}

type CacheCleanView struct {
	Path        string
	Removed     int
	RemovedSize string
}

func (a *App) ListCache(path string) (CacheListView, error) {
	result, err := a.core.ListCache(strings.TrimSpace(path))
	if err != nil {
		return CacheListView{}, err
	}

	entries := make([]CacheEntryView, 0, len(result.Entries))
	for _, entry := range result.Entries {
		kind := "file"
		if entry.IsDir {
			kind = "dir"
		}
		entries = append(entries, CacheEntryView{
			Path: entry.Path,
			Name: filepath.Base(entry.Path),
			Kind: kind,
			Size: formatBytes(entry.Size),
		})
	}

	return CacheListView{
		Path:    result.Path,
		Entries: entries,
		Size:    formatBytes(result.Size),
	}, nil
}

func (a *App) CleanCache(path string) (CacheCleanView, error) {
	result, err := a.core.CleanCache(strings.TrimSpace(path))
	if err != nil {
		return CacheCleanView{}, err
	}

	return CacheCleanView{
		Path:        result.Path,
		Removed:     result.Removed,
		RemovedSize: formatBytes(result.RemovedSize),
	}, nil
}

type ConvertView struct {
	InputPath    string
	Files        []InspectFileView
	TotalTime    string
	MetadataPath string
	Title        string
	CoverPath    string
	ChapterEvery string
	OutputPath   string
	DryRun       bool
	Command      []string
	Status       string
}

func (a *App) ConvertAudio(inputPath, outputPath, metadataPath, coverPath, chapterEvery string, dryRun, overwrite bool) (ConvertView, error) {
	result, err := a.core.Convert(a.dialogContext(), coreapp.ConvertRequest{
		InputPath:    strings.TrimSpace(inputPath),
		OutputPath:   strings.TrimSpace(outputPath),
		MetadataPath: strings.TrimSpace(metadataPath),
		CoverPath:    strings.TrimSpace(coverPath),
		ChapterEvery: strings.TrimSpace(chapterEvery),
		DryRun:       dryRun,
		Overwrite:    overwrite,
	})
	if err != nil {
		return ConvertView{}, err
	}

	files := make([]InspectFileView, 0, len(result.Input.Files))
	var total time.Duration
	for _, file := range result.Input.Files {
		total += file.Duration
		files = append(files, InspectFileView{
			Path:     file.Path,
			Name:     file.Name,
			Duration: formatDuration(file.Duration),
			Codec:    file.Codec,
			Bitrate:  formatBitrate(file.Bitrate),
			Channels: file.Channels,
			Tags:     file.Tags,
			Chapters: len(file.Chapters),
		})
	}

	status := "done"
	if result.DryRun {
		status = "planned"
	}
	return ConvertView{
		InputPath:    result.Input.Path,
		Files:        files,
		TotalTime:    formatDuration(total),
		MetadataPath: result.MetadataPath,
		Title:        result.Metadata.Title,
		CoverPath:    result.CoverPath,
		ChapterEvery: result.ChapterEvery,
		OutputPath:   result.OutputPath,
		DryRun:       result.DryRun,
		Command:      result.Command,
		Status:       status,
	}, nil
}

func candidateView(candidate providers.Candidate) MetadataCandidateView {
	return MetadataCandidateView{
		Provider:    candidate.Provider,
		ID:          candidate.ID,
		Title:       candidate.Title,
		Authors:     candidate.Authors,
		Narrators:   candidate.Narrators,
		Series:      candidate.Series,
		SeriesIndex: candidate.SeriesIndex,
		Year:        candidate.Year,
		Duration:    formatDuration(candidate.Duration),
		CoverURL:    candidate.CoverURL,
		Confidence:  formatConfidence(candidate.Confidence),
	}
}

func bookView(book metadata.Book) BookMetadataView {
	return BookMetadataView{
		Title:         book.Title,
		Subtitle:      book.Subtitle,
		Authors:       book.Authors,
		Author:        book.Author,
		Narrators:     book.Narrators,
		Narrator:      book.Narrator,
		Series:        book.Series,
		SeriesIndex:   book.SeriesIndex,
		Language:      book.Language,
		Genre:         book.Genre,
		Description:   book.Description,
		Publisher:     book.Publisher,
		PublishedYear: book.PublishedYear,
		Cover:         book.Cover,
	}
}

func formatDuration(duration time.Duration) string {
	if duration <= 0 {
		return ""
	}
	duration = duration.Round(time.Second)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	if hours > 0 {
		return strconv.Itoa(hours) + "h " + formatTwoDigits(minutes) + "m " + formatTwoDigits(seconds) + "s"
	}
	return strconv.Itoa(minutes) + "m " + formatTwoDigits(seconds) + "s"
}

func formatTwoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

func formatBitrate(bitrate int) string {
	if bitrate <= 0 {
		return ""
	}
	return strconv.Itoa(bitrate/1000) + " kbps"
}

func formatConfidence(confidence float64) string {
	if confidence <= 0 {
		return ""
	}
	return strconv.Itoa(int(confidence*100+0.5)) + "%"
}

func formatBytes(size int64) string {
	const unit = int64(1024)
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	value := float64(size)
	for _, suffix := range []string{"KiB", "MiB", "GiB"} {
		value /= float64(unit)
		if value < float64(unit) {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TiB", value/float64(unit))
}
