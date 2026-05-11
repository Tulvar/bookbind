package app

import "testing"

func TestNewProviderRegistryDefaults(t *testing.T) {
	registry, err := NewProviderRegistry(nil)
	if err != nil {
		t.Fatalf("NewProviderRegistry() error = %v", err)
	}
	if registry == nil {
		t.Fatal("registry is nil")
	}
}

func TestNewProviderRegistrySelectsProviders(t *testing.T) {
	registry, err := NewProviderRegistryWithCache([]string{"openlibrary", "google"}, t.TempDir())
	if err != nil {
		t.Fatalf("NewProviderRegistryWithCache() error = %v", err)
	}
	if registry == nil {
		t.Fatal("registry is nil")
	}
}

func TestNewProviderRegistryRejectsUnknownProvider(t *testing.T) {
	_, err := NewProviderRegistryWithCache([]string{"unknown"}, t.TempDir())
	if err == nil {
		t.Fatal("NewProviderRegistryWithCache() error = nil, want error")
	}
}

func TestProviderNamesCSV(t *testing.T) {
	got := ProviderNamesCSV([]ProviderInfo{{Name: "googlebooks"}, {Name: "openlibrary"}})
	if got != "googlebooks,openlibrary" {
		t.Fatalf("ProviderNamesCSV() = %q", got)
	}
}
