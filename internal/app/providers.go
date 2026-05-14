package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Tulvar/bookbind/internal/providers"
	"github.com/Tulvar/bookbind/internal/providers/fantlab"
	"github.com/Tulvar/bookbind/internal/providers/googlebooks"
	"github.com/Tulvar/bookbind/internal/providers/openlibrary"
)

type ProviderInfo struct {
	Name    string
	Enabled bool
}

func AvailableProviders() []ProviderInfo {
	return []ProviderInfo{
		{Name: "openlibrary", Enabled: true},
		{Name: "fantlab", Enabled: true},
		{Name: "googlebooks", Enabled: true},
	}
}

type ProviderConfig struct {
	GoogleBooksAPIKey string
}

func NewProviderRegistry(names []string) (*providers.Registry, error) {
	return NewProviderRegistryWithConfig(names, ProviderConfig{})
}

func NewProviderRegistryWithConfig(names []string, config ProviderConfig) (*providers.Registry, error) {
	cachePath, err := CacheDir()
	if err != nil {
		return nil, err
	}
	return NewProviderRegistryWithCacheAndConfig(names, cachePath, config)
}

func NewProviderRegistryWithCache(names []string, cachePath string) (*providers.Registry, error) {
	return NewProviderRegistryWithCacheAndConfig(names, cachePath, ProviderConfig{})
}

func NewProviderRegistryWithCacheAndConfig(names []string, cachePath string, config ProviderConfig) (*providers.Registry, error) {
	if len(names) == 0 {
		names = defaultProviderNames()
	}

	cache := providers.NewCache(cachePath)
	selected := make([]providers.Provider, 0, len(names))
	for _, name := range names {
		provider, err := providerByName(name, config)
		if err != nil {
			return nil, err
		}
		selected = append(selected, cache.Wrap(provider))
	}
	return providers.NewRegistry(selected...), nil
}

func ProviderNamesCSV(infos []ProviderInfo) string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

func CanonicalProviderName(name string) (string, error) {
	provider, err := providerByName(name, ProviderConfig{})
	if err != nil {
		return "", err
	}
	return provider.Name(), nil
}

func defaultProviderNames() []string {
	return []string{"openlibrary", "fantlab", "googlebooks"}
}

func providerByName(name string, config ProviderConfig) (providers.Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openlibrary":
		return openlibrary.New(), nil
	case "fantlab":
		return fantlab.New(), nil
	case "googlebooks", "google":
		return googlebooks.New(googlebooks.WithAPIKey(config.GoogleBooksAPIKey)), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
