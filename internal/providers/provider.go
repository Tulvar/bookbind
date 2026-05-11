package providers

import (
	"context"
	"time"
)

type Provider interface {
	Name() string
	Search(ctx context.Context, query SearchQuery) ([]Candidate, error)
	Get(ctx context.Context, id string) (Candidate, error)
}

type SearchQuery struct {
	Title       string
	Author      string
	Narrator    string
	Series      string
	SeriesIndex string
	Language    string
	ISBN        string
	ASIN        string
	Duration    time.Duration
}

func (q SearchQuery) Empty() bool {
	return q.Title == "" &&
		q.Author == "" &&
		q.Narrator == "" &&
		q.Series == "" &&
		q.SeriesIndex == "" &&
		q.Language == "" &&
		q.ISBN == "" &&
		q.ASIN == "" &&
		q.Duration == 0
}

type Candidate struct {
	Provider    string
	ID          string
	Title       string
	Authors     []string
	Narrators   []string
	Series      string
	SeriesIndex string
	Year        int
	Duration    time.Duration
	CoverURL    string
	Confidence  float64
}
