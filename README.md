# bookbind

`bookbind` converts MP3 audiobook files into M4B.

Current milestone: `v0.2.0`.

The current focus is a reliable Go core that can be reused by CLI, tests, and
the future desktop UI.

## Current CLI

Inspect a single MP3 file or a directory with MP3 files:

```bash
go run ./cmd/bookbind inspect ./book.mp3
go run ./cmd/bookbind inspect ./book-directory
```

`inspect` shows audio properties, embedded metadata, and existing chapters when
they are present.

Create a metadata template next to the input:

```bash
go run ./cmd/bookbind template ./book.mp3
go run ./cmd/bookbind template ./book-directory --output ./bookbind.yaml
```

The template command tries to infer `title`, `author`, `series`, and
`series_index` from embedded MP3 tags and common filename patterns.

Search metadata candidates:

```bash
go run ./cmd/bookbind search --title "Ночной дозор" --author "Лукьяненко"
go run ./cmd/bookbind search --title "Ночной дозор" --provider openlibrary
```

List available metadata providers:

```bash
go run ./cmd/bookbind providers
```

The first real metadata providers are Open Library and Google Books. Search uses
all enabled providers by default; pass `--provider openlibrary,googlebooks` to
limit a search to selected sources.

Save a selected metadata candidate:

```bash
go run ./cmd/bookbind metadata --provider googlebooks --id <candidate-id> --output bookbind.yaml
```

Convert a single MP3 file or a directory with MP3 files:

```bash
go run ./cmd/bookbind convert ./book.mp3 --output ./book.m4b
go run ./cmd/bookbind convert ./book-directory --output ./book.m4b
```

When converting a directory, MP3 files are sorted by filename and written as M4B
chapters using their filenames as chapter titles.

Preview conversion without writing output:

```bash
go run ./cmd/bookbind convert ./book.mp3 --output ./book.m4b --dry-run
```

Create synthetic chapters for a single MP3:

```bash
go run ./cmd/bookbind convert ./book.mp3 --chapter-every 10m --output ./book.m4b
```

Use manual metadata:

```bash
go run ./cmd/bookbind convert ./book.mp3 --metadata ./bookbind.yaml --output ./book.m4b
```

Attach a local cover:

```bash
go run ./cmd/bookbind convert ./book.mp3 --cover ./cover.jpg --output ./book.m4b
```

You can also set `cover: "cover.jpg"` in `bookbind.yaml`. Relative cover paths
inside YAML are resolved relative to the YAML file.

Example metadata:

```yaml
title: "Ночной дозор"
author: "Сергей Лукьяненко"
narrator: ""
series: "Дозоры"
series_index: "1"
language: "ru"
genre: "Фантастика"
published_year: 1998
description: |
  Описание книги.
```

## Development

Run the checks used by CI:

```bash
go vet ./...
go test ./...
go build ./cmd/bookbind
```

`ffmpeg` and `ffprobe` must be available on `PATH` for real inspect/convert
runs.

## Milestones

### v0.2.0

- metadata template generation
- filename parser for common audiobook naming patterns
- embedded MP3 tags in templates
- synthetic chapters via `--chapter-every`
- richer `inspect` output for metadata and chapters

### v0.1.0

- MP3 and MP3 directory conversion to M4B
- chapters from files
- local cover support
- manual metadata YAML
- dry-run and overwrite protection
