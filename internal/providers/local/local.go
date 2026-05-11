package local

import (
	"context"
	"fmt"

	"github.com/Tulvar/bookbind/internal/providers"
)

type Provider struct {
	name       string
	candidates []providers.Candidate
}

func New(candidates []providers.Candidate) *Provider {
	return &Provider{
		name:       "local",
		candidates: candidates,
	}
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Search(ctx context.Context, query providers.SearchQuery) ([]providers.Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]providers.Candidate, len(p.candidates))
	copy(result, p.candidates)
	return result, nil
}

func (p *Provider) Get(ctx context.Context, id string) (providers.Candidate, error) {
	if err := ctx.Err(); err != nil {
		return providers.Candidate{}, err
	}

	for _, candidate := range p.candidates {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return providers.Candidate{}, fmt.Errorf("candidate not found: %s", id)
}
