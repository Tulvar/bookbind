package providers

import (
	"math"
	"time"

	"github.com/Tulvar/bookbind/internal/matching"
)

func Score(query SearchQuery, candidate Candidate) float64 {
	var score float64
	var max float64

	if query.Title != "" {
		max += 40
		score += titleScore(query.Title, candidate.Title)
	}
	if query.Author != "" {
		max += 25
		if anyPersonMatch(query.Author, candidate.Authors) {
			score += 25
		}
	}
	if query.Series != "" {
		max += 15
		if matching.Normalize(query.Series) == matching.Normalize(candidate.Series) {
			score += 15
		}
	}
	if query.SeriesIndex != "" {
		max += 10
		if matching.Normalize(query.SeriesIndex) == matching.Normalize(candidate.SeriesIndex) {
			score += 10
		}
	}
	if query.Duration > 0 && candidate.Duration > 0 {
		max += 20
		if durationClose(query.Duration, candidate.Duration) {
			score += 20
		}
	}
	if query.Narrator != "" {
		max += 20
		if anyPersonMatch(query.Narrator, candidate.Narrators) {
			score += 20
		}
	}

	if max == 0 {
		return 0
	}
	return math.Round((score/max)*100) / 100
}

func titleScore(queryTitle, candidateTitle string) float64 {
	query := matching.Normalize(queryTitle)
	candidate := matching.Normalize(candidateTitle)
	if query == "" || candidate == "" {
		return 0
	}
	if query == candidate {
		return 40
	}
	if containsWithBoundaries(candidate, query) || containsWithBoundaries(query, candidate) {
		return 30
	}
	return 0
}

func anyPersonMatch(query string, values []string) bool {
	for _, value := range values {
		if matching.SamePerson(query, value) {
			return true
		}
	}
	return false
}

func durationClose(a, b time.Duration) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= 2*time.Minute
}

func containsWithBoundaries(value, sequence string) bool {
	if len(sequence) > len(value) {
		return false
	}
	for i := 0; i+len(sequence) <= len(value); i++ {
		if value[i:i+len(sequence)] != sequence {
			continue
		}
		beforeOK := i == 0 || value[i-1] == ' '
		after := i + len(sequence)
		afterOK := after == len(value) || value[after] == ' '
		if beforeOK && afterOK {
			return true
		}
	}
	return false
}
