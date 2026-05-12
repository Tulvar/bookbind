package main

import (
	"context"
	"strconv"
	"time"

	coreapp "github.com/Tulvar/bookbind/internal/app"
	"github.com/Tulvar/bookbind/internal/audio"
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
