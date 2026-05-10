package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Tulvar/bookbind/internal/audio"
)

func TestReorderFlagArgsAllowsFlagsAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output", "book.m4b", "--dry-run"},
		map[string]bool{
			"output":  true,
			"dry-run": false,
		},
	)
	want := []string{"--output", "book.m4b", "--dry-run", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsKeepsEqualsFlagTogether(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output=book.m4b"},
		map[string]bool{"output": true},
	)
	want := []string{"--output=book.m4b", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsAllowsTemplateOutputAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--output", "custom.yaml"},
		map[string]bool{"output": true},
	)
	want := []string{"--output", "custom.yaml", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestReorderFlagArgsAllowsChapterEveryAfterPositional(t *testing.T) {
	got := reorderFlagArgs(
		[]string{"book.mp3", "--chapter-every", "10m"},
		map[string]bool{"chapter-every": true},
	)
	want := []string{"--chapter-every", "10m", "book.mp3"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestPrintEmbeddedTags(t *testing.T) {
	var buffer bytes.Buffer

	printEmbeddedTags(&buffer, audio.EmbeddedTags{
		Title:  "Night Watch",
		Artist: "Sergey Lukyanenko",
	})

	got := buffer.String()
	for _, want := range []string{
		"Embedded metadata:",
		"title: Night Watch",
		"artist: Sergey Lukyanenko",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}

func TestPrintChapters(t *testing.T) {
	var buffer bytes.Buffer

	printChapters(&buffer, []audio.Chapter{
		{Title: "Intro", Start: 0, End: time.Minute},
	})

	got := buffer.String()
	for _, want := range []string{
		"Chapters: 1",
		"Intro (00:00.000 - 01:00.000)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output does not contain %q:\n%s", want, got)
		}
	}
}
