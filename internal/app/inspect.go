package app

import (
	"context"

	"github.com/Tulvar/bookbind/internal/audio"
)

type InspectRequest struct {
	InputPath string
}

type InspectResult struct {
	Input audio.Input
}

func (a *App) InspectInput(ctx context.Context, req InspectRequest) (InspectResult, error) {
	input, err := a.inspector.Inspect(ctx, req.InputPath)
	if err != nil {
		return InspectResult{}, err
	}

	return InspectResult{Input: input}, nil
}
