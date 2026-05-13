package providers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
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
	var providerErrors []error
	for _, provider := range r.providers {
		candidates, err := provider.Search(ctx, query)
		if err != nil {
			providerErrors = append(providerErrors, fmt.Errorf("%s search: %w", provider.Name(), err))
			continue
		}
		for _, candidate := range candidates {
			if candidate.Provider == "" {
				candidate.Provider = provider.Name()
			}
			candidate.Confidence = Score(query, candidate)
			all = append(all, candidate)
		}
	}
	if len(all) == 0 && len(providerErrors) > 0 {
		return nil, errors.Join(providerErrors...)
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Confidence > all[j].Confidence
	})
	return all, nil
}

func (r *Registry) Get(ctx context.Context, providerName, id string) (Candidate, error) {
	providerName = strings.TrimSpace(providerName)
	if providerName == "" {
		return Candidate{}, fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(id) == "" {
		return Candidate{}, fmt.Errorf("candidate id is required")
	}

	for _, provider := range r.providers {
		if strings.EqualFold(provider.Name(), providerName) {
			candidate, err := provider.Get(ctx, id)
			if err != nil {
				return Candidate{}, err
			}
			if candidate.Provider == "" {
				candidate.Provider = provider.Name()
			}
			return candidate, nil
		}
	}
	return Candidate{}, fmt.Errorf("unknown provider %q", providerName)
}
