package chapters

import (
	"testing"
	"time"
)

func TestSyntheticChapters(t *testing.T) {
	got, err := Synthetic(25*time.Minute, 10*time.Minute)
	if err != nil {
		t.Fatalf("Synthetic() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Title != "Chapter 001" || got[0].Start != 0 || got[0].End != 10*time.Minute {
		t.Fatalf("first chapter = %#v", got[0])
	}
	if got[2].Title != "Chapter 003" || got[2].Start != 20*time.Minute || got[2].End != 25*time.Minute {
		t.Fatalf("last chapter = %#v", got[2])
	}
}

func TestSyntheticRejectsMissingDuration(t *testing.T) {
	_, err := Synthetic(0, 10*time.Minute)
	if err == nil {
		t.Fatal("Synthetic() error = nil, want error")
	}
}

func TestSyntheticRejectsInvalidInterval(t *testing.T) {
	_, err := Synthetic(time.Hour, 0)
	if err == nil {
		t.Fatal("Synthetic() error = nil, want error")
	}
}
