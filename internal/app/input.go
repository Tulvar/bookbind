package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Tulvar/bookbind/internal/audio"
)

func (a *App) inspectMP3Input(ctx context.Context, inputPath string) (audio.Input, error) {
	input, err := a.inspector.Inspect(ctx, inputPath)
	if err != nil {
		return audio.Input{}, err
	}

	for _, file := range input.Files {
		if !strings.EqualFold(filepath.Ext(file.Path), ".mp3") {
			return audio.Input{}, fmt.Errorf("input must contain only mp3 files: %s", file.Path)
		}
	}

	return input, nil
}
