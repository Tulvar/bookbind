package chapters

import (
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
)

func TestFromAudioFilesBuildsSequentialChapters(t *testing.T) {
	got, err := FromAudioFiles([]audio.File{
		{Name: "01 - Start.mp3", Path: "01 - Start.mp3", Duration: 30 * time.Second},
		{Name: "02 - Continue.mp3", Path: "02 - Continue.mp3", Duration: 45 * time.Second},
	})
	if err != nil {
		t.Fatalf("FromAudioFiles() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Title != "01 - Start" {
		t.Fatalf("first title = %q", got[0].Title)
	}
	if got[0].Start != 0 || got[0].End != 30*time.Second {
		t.Fatalf("first timing = %v..%v", got[0].Start, got[0].End)
	}
	if got[1].Start != 30*time.Second || got[1].End != 75*time.Second {
		t.Fatalf("second timing = %v..%v", got[1].Start, got[1].End)
	}
}

func TestFromAudioFilesRejectsMissingDuration(t *testing.T) {
	_, err := FromAudioFiles([]audio.File{{Name: "01.mp3", Path: "01.mp3"}})
	if err == nil {
		t.Fatal("FromAudioFiles() error = nil, want error")
	}
}
