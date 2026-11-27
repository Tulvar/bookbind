package chapters

import (
	"strings"
	"testing"
	"time"
)

func TestSyntheticLimitsChapterCount(t *testing.T) {
	interval := time.Millisecond
	total := 10_000 * interval
	got, err := Synthetic(total, interval)
	if err != nil || len(got) != 10_000 {
		t.Fatalf("at limit: got %d chapters, err = %v", len(got), err)
	}
	got, err = Synthetic(total+time.Nanosecond, interval)
	if len(got) != 0 || err == nil || !strings.Contains(err.Error(), "10001 chapters") {
		t.Fatalf("above limit: got %d chapters, err = %v", len(got), err)
	}
}

func TestSyntheticClampsEndWithoutOverflow(t *testing.T) {
	const total = time.Duration(1<<63 - 1)
	got, err := Synthetic(total, total-1)
	if err != nil || len(got) != 2 || got[1].End != total {
		t.Fatalf("Synthetic() = %+v, %v", got, err)
	}
}

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
