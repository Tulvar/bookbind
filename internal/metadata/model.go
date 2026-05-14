package metadata

type Book struct {
	Title         string   `yaml:"title"`
	Subtitle      string   `yaml:"subtitle"`
	Authors       []string `yaml:"authors"`
	Author        string   `yaml:"author"`
	Narrators     []string `yaml:"narrators"`
	Narrator      string   `yaml:"narrator"`
	Translators   []string `yaml:"translators"`
	Translator    string   `yaml:"translator"`
	Series        string   `yaml:"series"`
	SeriesIndex   string   `yaml:"series_index"`
	Language      string   `yaml:"language"`
	Genre         string   `yaml:"genre"`
	Description   string   `yaml:"description"`
	Publisher     string   `yaml:"publisher"`
	PublishedYear int      `yaml:"published_year"`
	Cover         string   `yaml:"cover"`
}

func (b Book) NormalizedAuthors() []string {
	if len(b.Authors) > 0 {
		return b.Authors
	}
	if b.Author != "" {
		return []string{b.Author}
	}
	return nil
}

func (b Book) NormalizedNarrators() []string {
	if len(b.Narrators) > 0 {
		return b.Narrators
	}
	if b.Narrator != "" {
		return []string{b.Narrator}
	}
	return nil
}

func (b Book) NormalizedTranslators() []string {
	if len(b.Translators) > 0 {
		return b.Translators
	}
	if b.Translator != "" {
		return []string{b.Translator}
	}
	return nil
}

func (b Book) Empty() bool {
	return b.Title == "" &&
		b.Subtitle == "" &&
		len(b.NormalizedAuthors()) == 0 &&
		len(b.NormalizedNarrators()) == 0 &&
		len(b.NormalizedTranslators()) == 0 &&
		b.Series == "" &&
		b.SeriesIndex == "" &&
		b.Language == "" &&
		b.Genre == "" &&
		b.Description == "" &&
		b.Publisher == "" &&
		b.PublishedYear == 0
}
