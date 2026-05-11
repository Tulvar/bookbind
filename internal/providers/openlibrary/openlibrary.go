package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/providers"
)

const defaultBaseURL = "https://openlibrary.org"

type Provider struct {
	baseURL    string
	httpClient *http.Client
	limit      int
}

type Option func(*Provider)

func New(options ...Option) *Provider {
	provider := &Provider{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		limit: 10,
	}
	for _, option := range options {
		option(provider)
	}
	return provider
}

func WithBaseURL(baseURL string) Option {
	return func(provider *Provider) {
		provider.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(provider *Provider) {
		provider.httpClient = client
	}
}

func WithLimit(limit int) Option {
	return func(provider *Provider) {
		provider.limit = limit
	}
}

func (p *Provider) Name() string {
	return "openlibrary"
}

func (p *Provider) Search(ctx context.Context, query providers.SearchQuery) ([]providers.Candidate, error) {
	requestURL, err := p.searchURL(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "bookbind")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openlibrary returned %s", resp.Status)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode openlibrary response: %w", err)
	}

	candidates := make([]providers.Candidate, 0, len(result.Docs))
	for _, doc := range result.Docs {
		candidate := doc.candidate(p.Name())
		if candidate.ID == "" && candidate.Title == "" {
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (p *Provider) Get(ctx context.Context, id string) (providers.Candidate, error) {
	results, err := p.Search(ctx, providers.SearchQuery{Title: id})
	if err != nil {
		return providers.Candidate{}, err
	}
	for _, candidate := range results {
		if candidate.ID == id {
			return candidate, nil
		}
	}
	return providers.Candidate{}, fmt.Errorf("candidate not found: %s", id)
}

func (p *Provider) searchURL(query providers.SearchQuery) (string, error) {
	if query.Empty() {
		return "", fmt.Errorf("search query is empty")
	}

	values := url.Values{}
	if query.Title != "" {
		values.Set("title", query.Title)
	}
	if query.Author != "" {
		values.Set("author", query.Author)
	}
	if query.ISBN != "" {
		values.Set("isbn", query.ISBN)
	}
	if query.Language != "" {
		values.Set("lang", query.Language)
	}
	values.Set("limit", strconv.Itoa(p.limit))
	values.Set("fields", "key,title,author_name,first_publish_year,cover_i")

	return p.baseURL + "/search.json?" + values.Encode(), nil
}

type searchResponse struct {
	Docs []searchDoc `json:"docs"`
}

type searchDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	CoverID          int      `json:"cover_i"`
}

func (doc searchDoc) candidate(provider string) providers.Candidate {
	return providers.Candidate{
		Provider: provider,
		ID:       doc.Key,
		Title:    doc.Title,
		Authors:  doc.AuthorName,
		Year:     doc.FirstPublishYear,
		CoverURL: coverURL(doc.CoverID),
	}
}

func coverURL(coverID int) string {
	if coverID == 0 {
		return ""
	}
	return fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", coverID)
}
