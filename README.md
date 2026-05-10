# bookbind

`bookbind` converts MP3 audiobook files into M4B.

Current milestone: `v0.1.0`.

The current focus is a reliable Go core that can be reused by CLI, tests, and
the future desktop UI.

## Current CLI

Inspect a single MP3 file or a directory with MP3 files:

```bash
go run ./cmd/bookbind inspect ./book.mp3
go run ./cmd/bookbind inspect ./book-directory
```

Create a metadata template next to the input:

```bash
go run ./cmd/bookbind template ./book.mp3
go run ./cmd/bookbind template ./book-directory --output ./bookbind.yaml
```

The template command tries to infer `title`, `author`, `series`, and
`series_index` from common filename patterns.

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
