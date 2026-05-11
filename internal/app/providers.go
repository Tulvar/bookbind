package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Tulvar/bookbind/internal/providers"
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
		{Name: "googlebooks", Enabled: true},
	}
}

func NewProviderRegistry(names []string) (*providers.Registry, error) {
	if len(names) == 0 {
		names = defaultProviderNames()
	}

	selected := make([]providers.Provider, 0, len(names))
	for _, name := range names {
		provider, err := providerByName(name)
		if err != nil {
			return nil, err
		}
		selected = append(selected, provider)
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
	provider, err := providerByName(name)
	if err != nil {
		return "", err
	}
	return provider.Name(), nil
}

func defaultProviderNames() []string {
	return []string{"openlibrary", "googlebooks"}
}

func providerByName(name string) (providers.Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openlibrary":
		return openlibrary.New(), nil
	case "googlebooks", "google":
		return googlebooks.New(), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
