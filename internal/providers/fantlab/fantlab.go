package fantlab

import (
	"bytes"
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

const defaultBaseURL = "https://api.fantlab.ru"

type Provider struct {
	baseURL    string
	httpClient *http.Client
}

type Option func(*Provider)

func New(options ...Option) *Provider {
	provider := &Provider{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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

func (p *Provider) Name() string {
	return "fantlab"
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
		return nil, fmt.Errorf("fantlab returned %s", resp.Status)
	}

	var results []searchWork
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&results); err != nil {
		return nil, fmt.Errorf("decode fantlab response: %w", err)
	}

	candidates := make([]providers.Candidate, 0, len(results))
	for _, result := range results {
		candidate := result.candidate(p.Name())
		if candidate.ID == "" && candidate.Title == "" {
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (p *Provider) Get(ctx context.Context, id string) (providers.Candidate, error) {
	id = strings.TrimPrefix(strings.TrimSpace(id), "work:")
	if id == "" {
		return providers.Candidate{}, fmt.Errorf("candidate id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/work/"+url.PathEscape(id), nil)
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
		return providers.Candidate{}, fmt.Errorf("fantlab returned %s", resp.Status)
	}

	var work workDetail
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&work); err != nil {
		return providers.Candidate{}, fmt.Errorf("decode fantlab response: %w", err)
	}
	return work.candidate(p.Name()), nil
}

func (p *Provider) searchURL(query providers.SearchQuery) (string, error) {
	search := strings.TrimSpace(strings.Join([]string{query.Title, query.Author}, " "))
	if search == "" {
		return "", fmt.Errorf("search query is empty")
	}

	values := url.Values{}
	values.Set("q", search)
	values.Set("page", "1")
	values.Set("onlymatches", "1")
	return p.baseURL + "/search-works?" + values.Encode(), nil
}

type searchWork struct {
	WorkID          flexibleInt `json:"work_id"`
	RusName         string      `json:"rusname"`
	Name            string      `json:"name"`
	FullName        string      `json:"fullname"`
	AllAuthorRus    string      `json:"all_autor_rusname"`
	FirstAuthorRus  string      `json:"autor1_rusname"`
	Year            flexibleInt `json:"year"`
	PictureEdition  flexibleInt `json:"pic_edition_id"`
	PictureEdition2 flexibleInt `json:"pic_edition_id_auto"`
}

func (w searchWork) candidate(provider string) providers.Candidate {
	return providers.Candidate{
		Provider: provider,
		ID:       strconv.Itoa(w.WorkID.Int()),
		Title:    firstNonEmpty(w.RusName, w.Name, w.FullName),
		Authors:  splitPeople(firstNonEmpty(w.AllAuthorRus, w.FirstAuthorRus)),
		Year:     w.Year.Int(),
		CoverURL: editionCoverURL(firstNonZero(w.PictureEdition.Int(), w.PictureEdition2.Int())),
	}
}

type workDetail struct {
	WorkID      flexibleInt  `json:"work_id"`
	Name        string       `json:"work_name"`
	Original    string       `json:"work_name_orig"`
	Year        flexibleInt  `json:"work_year"`
	Description string       `json:"work_description"`
	Language    string       `json:"lang_code"`
	Image       string       `json:"image"`
	Authors     []workAuthor `json:"authors"`
	GenreGroups []genreGroup `json:"-"`
	WorkType    string       `json:"work_type"`
	Title       string       `json:"title"`
}

type workAuthor struct {
	Name string `json:"name"`
}

func (w workDetail) candidate(provider string) providers.Candidate {
	authors := make([]string, 0, len(w.Authors))
	for _, author := range w.Authors {
		if strings.TrimSpace(author.Name) != "" {
			authors = append(authors, strings.TrimSpace(author.Name))
		}
	}
	return providers.Candidate{
		Provider: provider,
		ID:       strconv.Itoa(w.WorkID.Int()),
		Title:    firstNonEmpty(w.Name, w.Original, titleWithoutAuthors(w.Title)),
		Authors:  authors,
		Year:     w.Year.Int(),
		CoverURL: normalizeFantLabURL(w.Image),
	}
}

type genreGroup struct{}

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte(`""`)) {
		*i = 0
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		parsed, _ := strconv.Atoi(number.String())
		*i = flexibleInt(parsed)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	parsed, _ := strconv.Atoi(strings.TrimSpace(text))
	*i = flexibleInt(parsed)
	return nil
}

func (i flexibleInt) Int() int {
	return int(i)
}

func splitPeople(value string) []string {
	value = stripTags(value)
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func stripTags(value string) string {
	value = strings.ReplaceAll(value, "<br>", ",")
	value = strings.ReplaceAll(value, "<br/>", ",")
	value = strings.ReplaceAll(value, "<br />", ",")
	return value
}

func editionCoverURL(editionID int) string {
	if editionID == 0 {
		return ""
	}
	return fmt.Sprintf("https://fantlab.ru/images/editions/big/%d", editionID)
}

func normalizeFantLabURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "//") {
		return "https:" + value
	}
	if strings.HasPrefix(value, "http://") {
		return "https://" + strings.TrimPrefix(value, "http://")
	}
	if strings.HasPrefix(value, "https://") {
		return value
	}
	return "https://fantlab.ru/" + strings.TrimLeft(value, "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func titleWithoutAuthors(value string) string {
	value = strings.TrimSpace(value)
	if left := strings.Index(value, "«"); left >= 0 {
		if right := strings.LastIndex(value, "»"); right > left {
			return value[left+len("«") : right]
		}
	}
	return value
}
