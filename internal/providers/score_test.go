package providers

import (
	"testing"
	"time"
)

func TestScoreExactMatch(t *testing.T) {
	got := Score(
		SearchQuery{
			Title:       "Ночной дозор",
			Author:      "Сергей Лукьяненко",
			Series:      "Дозоры",
			SeriesIndex: "1",
			Duration:    12 * time.Hour,
		},
		Candidate{
			Title:       "Ночной дозор",
			Authors:     []string{"Лукьяненко Сергей"},
			Series:      "Дозоры",
			SeriesIndex: "1",
			Duration:    12*time.Hour + time.Minute,
		},
	)

	if got != 1 {
		t.Fatalf("Score() = %v, want 1", got)
	}
}

func TestScorePartialTitleMatch(t *testing.T) {
	got := Score(
		SearchQuery{Title: "Ночной дозор"},
		Candidate{Title: "Дозоры 01 - Ночной дозор"},
	)

	if got != 0.75 {
		t.Fatalf("Score() = %v, want 0.75", got)
	}
}

func TestScoreMismatch(t *testing.T) {
	got := Score(
		SearchQuery{Title: "Ночной дозор", Author: "Сергей Лукьяненко"},
		Candidate{Title: "Дневной дозор", Authors: []string{"Другой Автор"}},
	)

	if got != 0 {
		t.Fatalf("Score() = %v, want 0", got)
	}
}
