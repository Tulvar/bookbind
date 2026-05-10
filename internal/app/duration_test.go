package app

import (
	"testing"
	"time"
)

func TestParseChapterInterval(t *testing.T) {
	tests := []struct {
		value string
		want  time.Duration
	}{
		{"", 0},
		{"10m", 10 * time.Minute},
		{"1h", time.Hour},
		{"15", 15 * time.Minute},
	}

	for _, test := range tests {
		got, err := parseChapterInterval(test.value)
		if err != nil {
			t.Fatalf("parseChapterInterval(%q) error = %v", test.value, err)
		}
		if got != test.want {
			t.Fatalf("parseChapterInterval(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}

func TestParseChapterIntervalRejectsInvalidValue(t *testing.T) {
	_, err := parseChapterInterval("soon")
	if err == nil {
		t.Fatal("parseChapterInterval() error = nil, want error")
	}
}
