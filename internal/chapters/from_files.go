package chapters

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
)

func FromAudioFiles(files []audio.File) ([]Chapter, error) {
	result := make([]Chapter, 0, len(files))
	var cursor time.Duration

	for i, file := range files {
		if file.Duration <= 0 {
			return nil, fmt.Errorf("file duration is required for chapters: %s", file.Path)
		}

		start := cursor
		end := start + file.Duration
		result = append(result, Chapter{
			Title: chapterTitle(i, file),
			Start: start,
			End:   end,
		})
		cursor = end
	}

	return result, nil
}

func chapterTitle(index int, file audio.File) string {
	title := strings.TrimSpace(file.Tags.Title)
	if title != "" {
		return title
	}
	name := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Sprintf("Chapter %03d", index+1)
	}
	return name
}
