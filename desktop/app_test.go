package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := map[time.Duration]string{
		0:                           "",
		65 * time.Second:            "1m 05s",
		2*time.Hour + 3*time.Second: "2h 00m 03s",
	}

	for duration, want := range tests {
		if got := formatDuration(duration); got != want {
			t.Fatalf("formatDuration(%s) = %q, want %q", duration, got, want)
		}
	}
}

func TestFormatBitrate(t *testing.T) {
	if got := formatBitrate(128000); got != "128 kbps" {
		t.Fatalf("formatBitrate() = %q", got)
	}
}
