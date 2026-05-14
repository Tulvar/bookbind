package metadata

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYAML(path string) (Book, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Book{}, err
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return Book{}, fmt.Errorf("parse metadata yaml %s: %w", path, err)
	}
	normalizeEmptyPublishedYear(&node)

	var book Book
	if err := node.Decode(&book); err != nil {
		return Book{}, fmt.Errorf("parse metadata yaml %s: %w", path, err)
	}
	return book, nil
}

func MarshalYAML(book Book) ([]byte, error) {
	data, err := yaml.Marshal(book)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata yaml: %w", err)
	}
	return data, nil
}

func MarshalTemplateYAML(book Book) ([]byte, error) {
	template := yamlTemplate{
		Title:             book.Title,
		Author:            book.Author,
		Narrator:          book.Narrator,
		Translator:        book.Translator,
		Series:            book.Series,
		SeriesIndex:       book.SeriesIndex,
		Language:          book.Language,
		Genre:             book.Genre,
		Publisher:         book.Publisher,
		Description:       book.Description,
		Cover:             book.Cover,
		ChaptersFromFiles: true,
	}
	if book.PublishedYear > 0 {
		template.PublishedYear = fmt.Sprintf("%d", book.PublishedYear)
	}

	data, err := yaml.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata template yaml: %w", err)
	}
	return data, nil
}

type yamlTemplate struct {
	Title             string `yaml:"title"`
	Author            string `yaml:"author"`
	Narrator          string `yaml:"narrator"`
	Translator        string `yaml:"translator"`
	Series            string `yaml:"series"`
	SeriesIndex       string `yaml:"series_index"`
	Language          string `yaml:"language"`
	Genre             string `yaml:"genre"`
	Publisher         string `yaml:"publisher"`
	PublishedYear     string `yaml:"published_year"`
	Description       string `yaml:"description"`
	Cover             string `yaml:"cover"`
	ChaptersFromFiles bool   `yaml:"chapters_from_files"`
}

func normalizeEmptyPublishedYear(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		normalizeEmptyPublishedYear(node.Content[0])
		return
	}
	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]
		if key.Value == "published_year" && value.Kind == yaml.ScalarNode {
			if value.Value == "" {
				value.Value = "0"
			}
			value.Tag = "!!int"
		}
	}
}
