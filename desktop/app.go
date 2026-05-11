package main

import (
	"context"

	coreapp "github.com/Tulvar/bookbind/internal/app"
	"github.com/Tulvar/bookbind/pkg/version"
)

type App struct {
	ctx  context.Context
	core *coreapp.App
}

func NewApp() *App {
	return &App{
		core: coreapp.New(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) AppVersion() string {
	return version.Version
}

func (a *App) AvailableProviders() []coreapp.ProviderInfo {
	return coreapp.AvailableProviders()
}
