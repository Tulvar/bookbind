package googlebooks

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

const defaultBaseURL = "https://www.googleapis.com/books/v1"

type Provider struct {
	baseURL    string
	httpClient *http.Client
	limit      int
	apiKey     string
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

func WithAPIKey(apiKey string) Option {
	return func(provider *Provider) {
		provider.apiKey = strings.TrimSpace(apiKey)
	}
}

func (p *Provider) Name() string {
	return "googlebooks"
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
		return nil, fmt.Errorf("googlebooks returned %s", resp.Status)
	}

	var result volumesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode googlebooks response: %w", err)
	}

	candidates := make([]providers.Candidate, 0, len(result.Items))
	for _, item := range result.Items {
		candidate := item.candidate(p.Name())
		if candidate.ID == "" && candidate.Title == "" {
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (p *Provider) Get(ctx context.Context, id string) (providers.Candidate, error) {
	if strings.TrimSpace(id) == "" {
		return providers.Candidate{}, fmt.Errorf("candidate id is required")
	}

	requestURL := p.baseURL + "/volumes/" + url.PathEscape(id)
	if p.apiKey != "" {
		values := url.Values{}
		values.Set("key", p.apiKey)
		requestURL += "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return providers.Candidate{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "bookbind")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return providers.Candidate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return providers.Candidate{}, fmt.Errorf("googlebooks returned %s", resp.Status)
	}

	var item volumeItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return providers.Candidate{}, fmt.Errorf("decode googlebooks response: %w", err)
	}
	return item.candidate(p.Name()), nil
}

func (p *Provider) searchURL(query providers.SearchQuery) (string, error) {
	if query.Empty() {
		return "", fmt.Errorf("search query is empty")
	}

	values := url.Values{}
	values.Set("q", googleQuery(query))
	values.Set("maxResults", strconv.Itoa(p.limit))
	values.Set("projection", "lite")
	if p.apiKey != "" {
		values.Set("key", p.apiKey)
	}
	return p.baseURL + "/volumes?" + values.Encode(), nil
}

func googleQuery(query providers.SearchQuery) string {
	var parts []string
	if query.ISBN != "" {
		parts = append(parts, "isbn:"+query.ISBN)
	}
	if query.Title != "" {
		parts = append(parts, "intitle:"+quoteTerm(query.Title))
	}
	if query.Author != "" {
		parts = append(parts, "inauthor:"+quoteTerm(query.Author))
	}
	if len(parts) == 0 {
		parts = append(parts, query.Title, query.Author)
	}
	return strings.Join(parts, " ")
}

func quoteTerm(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, " ") {
		return `"` + value + `"`
	}
	return value
}

type volumesResponse struct {
	Items []volumeItem `json:"items"`
}

type volumeItem struct {
	ID         string     `json:"id"`
	VolumeInfo volumeInfo `json:"volumeInfo"`
}

type volumeInfo struct {
	Title         string     `json:"title"`
	Subtitle      string     `json:"subtitle"`
	Authors       []string   `json:"authors"`
	Publisher     string     `json:"publisher"`
	PublishedDate string     `json:"publishedDate"`
	Description   string     `json:"description"`
	Categories    []string   `json:"categories"`
	ImageLinks    imageLinks `json:"imageLinks"`
}

type imageLinks struct {
	Thumbnail string `json:"thumbnail"`
	Small     string `json:"small"`
	Medium    string `json:"medium"`
	Large     string `json:"large"`
}

func (item volumeItem) candidate(provider string) providers.Candidate {
	return providers.Candidate{
		Provider: provider,
		ID:       item.ID,
		Title:    fullTitle(item.VolumeInfo.Title, item.VolumeInfo.Subtitle),
		Authors:  item.VolumeInfo.Authors,
		Year:     publishedYear(item.VolumeInfo.PublishedDate),
		CoverURL: bestImage(item.VolumeInfo.ImageLinks),
	}
}

func fullTitle(title, subtitle string) string {
	if strings.TrimSpace(subtitle) == "" {
		return title
	}
	return title + ": " + subtitle
}

func publishedYear(value string) int {
	if len(value) >= 4 {
		year, err := strconv.Atoi(value[:4])
		if err == nil {
			return year
		}
	}
	return 0
}

func bestImage(links imageLinks) string {
	for _, value := range []string{links.Large, links.Medium, links.Small, links.Thumbnail} {
		if value != "" {
			return strings.Replace(value, "http://", "https://", 1)
		}
	}
	return ""
}
