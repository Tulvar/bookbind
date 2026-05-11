package providers

import (
	"context"
	"fmt"
	"sort"
)

type Registry struct {
	providers []Provider
}

func NewRegistry(providers ...Provider) *Registry {
	return &Registry{providers: providers}
}

func (r *Registry) Search(ctx context.Context, query SearchQuery) ([]Candidate, error) {
	if query.Empty() {
		return nil, fmt.Errorf("search query is empty")
	}

	var all []Candidate
	for _, provider := range r.providers {
		candidates, err := provider.Search(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("%s search: %w", provider.Name(), err)
		}
		for _, candidate := range candidates {
			if candidate.Provider == "" {
				candidate.Provider = provider.Name()
			}
			candidate.Confidence = Score(query, candidate)
			all = append(all, candidate)
		}
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Confidence > all[j].Confidence
	})
	return all, nil
}
