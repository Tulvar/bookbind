package app

import (
	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/m4b"
	"github.com/Tulvar/bookbind/internal/providers"
)

// App exposes bookbind use cases to CLI, desktop UI, and tests.
type App struct {
	inspector *audio.Inspector
	builder   *m4b.Builder
	providers *providers.Registry
}

type Option func(*App)

func New(options ...Option) *App {
	app := &App{
		inspector: audio.NewInspector(),
		builder:   m4b.NewBuilder("ffmpeg"),
		providers: providers.NewRegistry(),
	}
	for _, option := range options {
		option(app)
	}
	return app
}

func WithInspector(inspector *audio.Inspector) Option {
	return func(app *App) {
		app.inspector = inspector
	}
}

func WithBuilder(builder *m4b.Builder) Option {
	return func(app *App) {
		app.builder = builder
	}
}

func WithProviders(registry *providers.Registry) Option {
	return func(app *App) {
		app.providers = registry
	}
}
