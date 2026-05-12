# bookbind

`bookbind` converts MP3 audiobook files into M4B.

Current milestone: `v0.4.0`.

The current focus is a Wails desktop UI for Windows, macOS, and Linux on top of
the existing Go core.

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
go run ./cmd/bookbind search --title "Ночной дозор" --select 1 --output bookbind.yaml
```

Search prints a table with provider, candidate id, title, authors, year, and
confidence. Use `--select` to write metadata from the chosen row, or use the
provider and id columns with the `metadata` command.

List available metadata providers:

```bash
go run ./cmd/bookbind providers
```

The first real metadata providers are Open Library and Google Books. Search uses
all enabled providers by default; pass `--provider openlibrary,googlebooks` to
limit a search to selected sources.

Save a selected metadata candidate:

```bash
go run ./cmd/bookbind metadata --provider googlebooks --id <candidate-id> --preview
go run ./cmd/bookbind metadata --provider googlebooks --id <candidate-id> --output bookbind.yaml
```

Convert a single MP3 file or a directory with MP3 files:

```bash
go run ./cmd/bookbind convert ./book.mp3 --output ./book.m4b
go run ./cmd/bookbind convert ./book-directory --output ./book.m4b
go run ./cmd/bookbind convert ./book.mp3 --interactive --select 1 --output ./book.m4b
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

Inspect or clean the local cache:

```bash
go run ./cmd/bookbind cache list
go run ./cmd/bookbind cache clean
```

Metadata provider responses are cached under the local bookbind cache directory.

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

Run the desktop shell locally:

```bash
cd desktop
wails dev
```

Release builds are created by pushing a version tag:

```bash
git tag v0.4.0
git push origin v0.4.0
```

The release workflow builds CLI artifacts named with the version, for example
`bookbind-v0.4.0-linux-amd64` and `bookbind-v0.4.0-windows-amd64.exe`.
It also builds desktop artifacts named like
`bookbind-desktop-v0.4.0-darwin-arm64.zip`.

`ffmpeg` and `ffprobe` must be available on `PATH` for real inspect/convert
runs.

## Milestones

### v0.4.0

- Wails desktop shell for Windows, macOS, and Linux
- React/TypeScript frontend
- Go bridge over the existing app use cases
- import, metadata, convert, and cache screens
- desktop build and CI smoke checks

### v0.3.0

- Open Library and Google Books metadata providers
- provider selection via `--provider`
- searchable candidate tables with confidence scores
- metadata preview and selected candidate export to `bookbind.yaml`
- interactive metadata selection for convert via `--interactive --select`
- local provider response cache with `cache list` and `cache clean`
- versioned release builds for Windows, macOS, and Linux

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
