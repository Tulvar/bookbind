package filename

import "testing"

func TestParseAuthorTitle(t *testing.T) {
	book := ParsePath("Сергей Лукьяненко - Ночной дозор.mp3")

	if book.Author != "Сергей Лукьяненко" {
		t.Fatalf("Author = %q", book.Author)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
}

func TestParseAuthorSeriesIndexTitle(t *testing.T) {
	book := ParsePath("Лукьяненко Сергей - Дозоры 01 - Ночной дозор.mp3")

	if book.Author != "Лукьяненко Сергей" {
		t.Fatalf("Author = %q", book.Author)
	}
	if book.Series != "Дозоры" {
		t.Fatalf("Series = %q", book.Series)
	}
	if book.SeriesIndex != "01" {
		t.Fatalf("SeriesIndex = %q", book.SeriesIndex)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
}

func TestParseSeriesIndexTitle(t *testing.T) {
	book := ParsePath("Дозоры 01. Ночной дозор.mp3")

	if book.Series != "Дозоры" {
		t.Fatalf("Series = %q", book.Series)
	}
	if book.SeriesIndex != "01" {
		t.Fatalf("SeriesIndex = %q", book.SeriesIndex)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
}

func TestParseLeadingTrackNumber(t *testing.T) {
	book := ParsePath("01 - Ночной дозор.mp3")

	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
	if book.Author != "" {
		t.Fatalf("Author = %q", book.Author)
	}
}

func TestParsePathUsesSeriesDirectory(t *testing.T) {
	book := ParsePath("/books/Дозоры/01 - Ночной дозор.mp3")

	if book.Series != "Дозоры" {
		t.Fatalf("Series = %q", book.Series)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
}

func TestParseNormalizesSeparators(t *testing.T) {
	book := ParsePath("Сергей  Лукьяненко — Ночной_дозор.mp3")

	if book.Author != "Сергей Лукьяненко" {
		t.Fatalf("Author = %q", book.Author)
	}
	if book.Title != "Ночной дозор" {
		t.Fatalf("Title = %q", book.Title)
	}
}
