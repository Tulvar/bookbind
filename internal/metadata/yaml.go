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

	var book Book
	if err := yaml.Unmarshal(data, &book); err != nil {
		return Book{}, fmt.Errorf("parse metadata yaml %s: %w", path, err)
	}
	return book, nil
}
