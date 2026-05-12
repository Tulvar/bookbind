# bookbind: архитектура и план разработки

`bookbind` — утилита для сборки нормальных M4B-аудиокниг из MP3-файлов: аудио + главы + обложка + метаданные.

Основная идея:

```text
MP3 / директория MP3
        ↓
анализ аудио
        ↓
поиск/сбор метаданных
        ↓
ручное подтверждение или metadata.yaml
        ↓
генерация глав
        ↓
конвертация в M4B
        ↓
запись тегов, обложки и chapters
```

---

## 1. Основные сценарии

### Сценарий 1. Один MP3-файл

```bash
bookbind convert ./book.mp3
```

Программа:

```text
1. читает длительность и теги из MP3
2. пытается понять название/автора из файла
3. ищет метаданные
4. предлагает кандидатов
5. скачивает/подставляет обложку
6. делает M4B
```

---

### Сценарий 2. Директория с несколькими MP3

```bash
bookbind convert "./Сергей Лукьяненко/Дозоры/01 - Ночной дозор"
```

Например:

```text
01.mp3
02.mp3
03.mp3
cover.jpg
metadata.yaml
```

Программа:

```text
1. сортирует MP3 по имени
2. считает длительность каждого файла
3. делает каждый MP3 отдельной главой
4. объединяет всё в один M4B
5. записывает главы
```

---

### Сценарий 3. Ручные метаданные

```bash
bookbind template ./book.mp3
```

Создает:

```yaml
title: ""
author: ""
narrator: ""
series: ""
series_index: ""
language: "ru"
publisher: ""
description: ""
cover: ""
chapters_from_files: true
```

Потом:

```bash
bookbind convert ./book.mp3 --metadata bookbind.yaml
```

---

### Сценарий 4. Только поиск метаданных

```bash
bookbind search --title "Ночной дозор" --author "Сергей Лукьяненко"
```

Вывод:

```text
[1] ЛитРес     Ночной дозор — Сергей Лукьяненко — чтец ...
[2] FantLab    Ночной дозор — Сергей Лукьяненко — Дозоры #1
[3] Google     Ночной дозор — Сергей Лукьяненко
```

---

## 2. CLI

Основные команды:

```bash
bookbind convert
bookbind inspect
bookbind search
bookbind template
bookbind metadata
bookbind providers
```

### `convert`

Главная команда.

```bash
bookbind convert ./book.mp3 --profile ru --output ./book.m4b
```

Опции:

```bash
--metadata bookbind.yaml
--profile ru
--provider litres,fantlab,google
--interactive
--output book.m4b
--cover cover.jpg
--chapters-from-files
--chapter-every 10m
--dry-run
--overwrite
```

---

### `inspect`

Показывает информацию о файле.

```bash
bookbind inspect ./book.mp3
```

Пример вывода:

```text
Input: book.mp3
Duration: 12h 43m 10s
Codec: mp3
Bitrate: 128 kbps
Channels: stereo

Embedded metadata:
  title: Ночной дозор
  artist: Сергей Лукьяненко
  album: Дозоры

Chapters:
  none
```

---

### `search`

Поиск метаданных.

```bash
bookbind search --title "Ночной дозор" --author "Лукьяненко"
```

---

### `template`

Создание шаблона.

```bash
bookbind template ./book.mp3
```

Создает:

```text
bookbind.yaml
```

---

## 3. Архитектура проекта

Пример структуры:

```text
bookbind/
  cmd/
    bookbind/
      main.go

  internal/
    app/
      convert.go
      inspect.go
      search.go
      template.go

    audio/
      probe.go
      input.go
      duration.go
      concat.go
      ffmpeg.go

    chapters/
      model.go
      from_files.go
      from_cue.go
      synthetic.go
      ffmetadata.go

    metadata/
      model.go
      merge.go
      normalize.go
      score.go
      yaml.go
      embedded.go

    providers/
      provider.go
      local/
      embedded/
      filename/
      fantlab/
      litres/
      mybook/
      googlebooks/
      openlibrary/
      librivox/

    cover/
      resolve.go
      download.go
      image.go

    m4b/
      build.go
      tags.go
      atoms.go

    cache/
      sqlite.go

    config/
      config.go
      profile.go

    ui/
      interactive.go
      table.go
      prompt.go

  pkg/
    version/
      version.go

  configs/
    profiles/
      ru.yaml
      default.yaml

  examples/
    metadata.ru.yaml

  README.md
  go.mod
```

---

## 4. Основные модули

### 4.1 `audio`

Отвечает за работу с аудиофайлами.

Задачи:

```text
- найти входные MP3
- отсортировать файлы
- получить duration
- получить codec/bitrate/channels
- прочитать embedded tags
- подготовить concat list для ffmpeg
- вызвать ffmpeg/ffprobe
```

Основные типы:

```go
type AudioInput struct {
    Path      string
    Files     []AudioFile
    TotalTime time.Duration
}

type AudioFile struct {
    Path      string
    Name      string
    Duration  time.Duration
    Codec     string
    Bitrate   int
    Channels  int
}
```

---

### 4.2 `metadata`

Главная модель метаданных.

```go
type BookMetadata struct {
    Title         string
    Subtitle      string
    Authors       []string
    Narrators     []string

    Series        string
    SeriesIndex   string

    Language      string
    Genre         string
    Description   string
    Publisher     string
    PublishedYear int

    ISBN10        []string
    ISBN13        []string
    ASIN          string

    Duration      time.Duration

    CoverURL      string
    CoverPath     string

    Chapters      []Chapter

    Source        string
    SourceID      string
    Confidence    float64
}
```

Глава:

```go
type Chapter struct {
    Title string
    Start time.Duration
    End   time.Duration
}
```

---

### 4.3 `providers`

Единый интерфейс для источников метаданных.

```go
type Provider interface {
    Name() string
    Search(ctx context.Context, q SearchQuery) ([]Candidate, error)
    Get(ctx context.Context, id string) (*metadata.BookMetadata, error)
}
```

Запрос:

```go
type SearchQuery struct {
    Title       string
    Author      string
    Narrator    string
    Series      string
    SeriesIndex string
    Language    string
    ISBN        string
    ASIN        string
    Duration    time.Duration
}
```

Кандидат:

```go
type Candidate struct {
    Provider    string
    ID          string
    Title       string
    Authors     []string
    Narrators   []string
    Series      string
    SeriesIndex string
    Year        int
    Duration    time.Duration
    CoverURL    string
    Confidence  float64
}
```

---

## 5. Провайдеры метаданных

Для русских аудиокниг стоит заложить такую цепочку:

```text
1. local metadata.yaml
2. embedded MP3 tags
3. filename parser
4. FantLab
5. ЛитРес experimental
6. MyBook experimental
7. Google Books
8. Open Library
9. LibriVox
```

### `local`

Читает `bookbind.yaml`.

Самый приоритетный источник.

---

### `embedded`

Берет теги из MP3:

```text
title
artist
album
composer
genre
date
comment
cover
```

---

### `filename`

Парсит путь и имена файлов.

Примеры:

```text
Сергей Лукьяненко - Ночной дозор.mp3
Лукьяненко Сергей - Дозоры 01 - Ночной дозор.mp3
01 - Ночной дозор.mp3
```

---

### `fantlab`

Хорош для:

```text
автор
название
серия
номер в серии
аннотация
год
жанры
обложка
```

Минус: не всегда есть аудио-специфичные данные.

---

### `litres` / `mybook`

Полезны именно для аудиоверсий:

```text
чтец
длительность
издательство
аудиообложка
жанр
возрастной рейтинг
```

Но их лучше делать как experimental.

---

## 6. Merge metadata

Нельзя просто взять первый найденный результат. Нужен merge с приоритетами.

Например:

```text
metadata.yaml        priority 100
embedded tags        priority 80
filename             priority 70
litres/mybook        priority 65
fantlab              priority 60
googlebooks          priority 40
openlibrary          priority 30
```

Логика:

```text
- если поле задано руками в YAML — не перетирать
- если есть narrator из ЛитРес — взять его
- если есть серия из FantLab — взять ее
- если описание есть из нескольких источников — выбрать самое полное
- если обложка локальная — не скачивать внешнюю
```

Пример:

```go
type MetadataMerger struct {
    Sources []MetadataSource
}

type MetadataSource struct {
    Name     string
    Priority int
    Data     BookMetadata
}
```

---

## 7. Matching и scoring

Для автоматического выбора кандидата нужен confidence.

Пример scoring:

```text
title exact match              +40
title normalized match         +30
author match                   +25
series match                   +15
series index match             +10
duration close                 +20
narrator match                 +20
year close                     +5

title mismatch                 -30
author mismatch                -40
duration very different        -20
```

Для русского языка обязательно:

```text
- lower case
- е/ё normalization
- удаление пунктуации
- нормализация пробелов
- сравнение "Имя Фамилия" и "Фамилия Имя"
```

Пример:

```text
"Сергей Лукьяненко - Ночной Дозор"
"Лукьяненко Сергей — Ночной дозор"
"С. Лукьяненко. Ночной дозор"
```

Должны считаться близкими.

---

## 8. Главы

Генерация глав — один из самых важных модулей.

Источники глав:

```text
1. embedded chapters
2. metadata.yaml
3. CUE-файл
4. каждый MP3-файл = отдельная глава
5. synthetic chapters каждые N минут
```

### Из файлов

Если вход — директория:

```text
01.mp3  30m
02.mp3  35m
03.mp3  28m
```

Получаем:

```text
Глава 1: 00:00:00
Глава 2: 00:30:00
Глава 3: 01:05:00
```

---

### Synthetic chapters

Если один большой MP3 без глав:

```bash
bookbind convert book.mp3 --chapter-every 10m
```

Получаем:

```text
Chapter 001 — 00:00:00
Chapter 002 — 00:10:00
Chapter 003 — 00:20:00
```

---

## 9. Сборка M4B

Технически лучше опираться на `ffmpeg`.

Пайплайн:

```text
1. подготовить input list
2. подготовить cover
3. подготовить ffmetadata с chapters
4. вызвать ffmpeg
5. проверить выходной m4b через ffprobe
```

Пример ffmpeg-логики:

```bash
ffmpeg \
  -f concat \
  -safe 0 \
  -i input.txt \
  -i cover.jpg \
  -i metadata.txt \
  -map 0:a \
  -map 1:v \
  -map_metadata 2 \
  -map_chapters 2 \
  -c:a aac \
  -b:a 64k \
  -c:v copy \
  -disposition:v attached_pic \
  output.m4b
```

Для одного файла:

```bash
ffmpeg \
  -i input.mp3 \
  -i cover.jpg \
  -i metadata.txt \
  -map 0:a \
  -map 1:v \
  -map_metadata 2 \
  -map_chapters 2 \
  -c:a aac \
  -b:a 64k \
  -c:v copy \
  -disposition:v attached_pic \
  output.m4b
```

---

## 10. Формат `bookbind.yaml`

`bookbind.yaml` — главный способ ручной коррекции.

```yaml
title: "Ночной дозор"
author: "Сергей Лукьяненко"
narrator: ""
series: "Дозоры"
series_index: "1"
language: "ru"
genre: "Фантастика"
publisher: ""
published_year: 1998

description: |
  ...

cover: "cover.jpg"

chapters_from_files: true

chapters:
  - title: "Глава 1"
    start: "00:00:00"
  - title: "Глава 2"
    start: "00:35:12"

source:
  provider: "manual"
  id: ""
```

---

## 11. Конфиг

Глобальный конфиг:

```yaml
profiles:
  default: ru

ffmpeg:
  path: ffmpeg
  ffprobe_path: ffprobe

cache:
  path: ~/.cache/bookbind/bookbind.db

providers:
  fantlab:
    enabled: true

  litres:
    enabled: false
    experimental: true

  mybook:
    enabled: false
    experimental: true

  googlebooks:
    enabled: true

  openlibrary:
    enabled: true

conversion:
  audio_codec: aac
  bitrate: 64k
  chapter_every: 10m
  overwrite: false
```

---

## 12. Профиль `ru`

```yaml
language: ru

preferred_providers:
  - local
  - embedded
  - filename
  - fantlab
  - litres
  - mybook
  - googlebooks
  - openlibrary
  - librivox

normalization:
  normalize_yo: true
  ignore_case: true
  strip_punctuation: true
  detect_reversed_author_name: true
  detect_series_number: true

matching:
  auto_select_threshold: 0.90
  interactive_threshold: 0.55
```

---

## 13. Кеш

Кеш нужен обязательно, чтобы:

```text
- не дергать API постоянно
- быстрее повторно собирать книгу
- сохранять выбранный match
- хранить обложки
```

Для кеша подойдет SQLite.

Таблицы:

```sql
providers_cache
---------------
provider
query_hash
response_json
created_at
expires_at

selected_matches
----------------
input_hash
provider
provider_id
metadata_json
created_at

covers
------
url
local_path
sha256
created_at
```

---

## 14. Безопасность и предсказуемость

Правила поведения программы:

```text
- никогда не перетирать output без --overwrite
- всегда иметь --dry-run
- всегда показывать, какие метаданные будут записаны
- всегда сохранять итоговый metadata.yaml рядом с output
- external providers не должны быть обязательны
- программа должна работать offline с metadata.yaml
```

Пример:

```bash
bookbind convert ./book.mp3 --dry-run
```

Вывод:

```text
Input:
  book.mp3

Detected:
  title: Ночной дозор
  author: Сергей Лукьяненко
  duration: 12h 43m

Metadata source:
  title: filename
  author: filename
  series: fantlab
  cover: cover.jpg
  narrator: manual

Output:
  Сергей Лукьяненко - Дозоры 01 - Ночной дозор.m4b
```

---

## 15. План разработки

### Этап 1. Минимальный рабочий конвертер

Цель: собрать `.m4b` из одного MP3 или директории MP3.

Сделать:

```text
- CLI skeleton
- команда convert
- поиск MP3-файлов
- сортировка файлов
- ffprobe duration
- ffmpeg convert
- output .m4b
- --overwrite
- --dry-run
```

Результат:

```bash
bookbind convert ./book.mp3 --output book.m4b
```

---

### Этап 2. Главы

Цель: нормальный M4B с chapters.

Сделать:

```text
- главы из нескольких MP3
- synthetic chapters каждые N минут
- генерация ffmetadata
- map_chapters через ffmpeg
- inspect chapters
```

Результат:

```bash
bookbind convert ./book-dir --chapters-from-files
bookbind convert ./book.mp3 --chapter-every 10m
```

---

### Этап 3. YAML metadata

Цель: ручное управление метаданными.

Сделать:

```text
- bookbind template
- чтение bookbind.yaml
- merge YAML metadata
- cover из локального файла
- запись title/artist/album/description/genre/date
```

Результат:

```bash
bookbind template ./book.mp3
bookbind convert ./book.mp3 --metadata bookbind.yaml
```

---

### Этап 4. Embedded tags и filename parser

Цель: программа сама предлагает начальные данные.

Сделать:

```text
- чтение тегов из MP3
- парсинг названия файла
- парсинг структуры директорий
- ru normalizer: е/ё, регистр, пунктуация
- определение серии и номера
```

Примеры, которые надо поддержать:

```text
Сергей Лукьяненко - Ночной дозор.mp3
Лукьяненко Сергей - Дозоры 01 - Ночной дозор.mp3
Дозоры 01. Ночной дозор.mp3
01 - Ночной дозор.mp3
```

---

### Этап 5. Search providers

Цель: искать метаданные.

Сделать:

```text
- [x] общий Provider interface
- [x] Google Books provider
- [x] Open Library provider
- [ ] FantLab provider
- [x] Candidate model
- [x] scoring
- [x] команда search
- [x] команда providers
- [x] выбор источников через --provider
```

Результат:

```bash
bookbind search --title "Ночной дозор" --author "Лукьяненко"
bookbind search --title "Ночной дозор" --provider openlibrary
bookbind providers
```

---

### Этап 6. Interactive mode

Цель: пользователь выбирает правильную книгу.

Сделать:

```text
- [x] кандидаты содержат provider:id для выбора
- [x] resolve выбранного provider:id
- [x] сохранение выбранных метаданных в bookbind.yaml
- [x] табличный вывод кандидатов
- [x] просмотр подробностей
- [x] выбор варианта
- [x] convert --interactive
```

Результат:

```bash
bookbind search --title "Ночной дозор"
bookbind search --title "Ночной дозор" --select 1 --output bookbind.yaml
bookbind metadata --provider googlebooks --id <candidate-id> --preview
bookbind metadata --provider googlebooks --id <candidate-id> --output bookbind.yaml
bookbind convert ./book.mp3 --interactive --select 1
```

---

### Этап 7. Cache

Цель: не дергать источники повторно.

Сделать:

```text
- [ ] SQLite cache
- [x] file cache for provider responses
- [ ] cache for selected matches
- [x] cache clean command
- [x] cache list command
```

Команды:

```bash
bookbind cache clean
bookbind cache list
```

---

### Этап 8. Experimental русские audio providers

Цель: подтягивать именно аудио-метаданные.

Сделать:

```text
- LitRes provider experimental
- MyBook provider experimental
- narrator
- audio duration
- audio publisher
- audio cover
```

Их лучше включать явно:

```bash
bookbind convert ./book.mp3 --provider litres --interactive
```

---

### Этап 9. Полировка

Сделать:

```text
- нормальный README
- примеры metadata.yaml
- обработка ошибок
- логирование
- progress ffmpeg
- unit tests для filename parser/scoring
- integration tests с короткими mp3
```

---

## 16. Приоритет MVP

Главный пользовательский интерфейс — нормальное desktop-приложение для
Windows/macOS/Linux. CLI остается обязательным, но в первую очередь как
автоматизируемый слой и способ тестировать ядро без UI.

Важно с самого начала разделить:

```text
core/use cases  -> convert, inspect, search, chapters, metadata
CLI             -> тонкая оболочка над use cases
Desktop UI      -> основная оболочка над теми же use cases
```

Так desktop UI не будет дублировать бизнес-логику, а тесты смогут проверять
ядро напрямую.

### MVP v0.1

```text
core application layer
bookbind convert
bookbind inspect
один MP3 → M4B
директория MP3 → M4B
chapters from files
local cover
metadata.yaml
dry-run
overwrite protection
```

### MVP v0.2

```text
metadata template generation
filename parser
embedded tags
synthetic chapters
richer inspect output
```

### MVP v0.3

```text
search
Google Books
Open Library
provider selection
candidate table
metadata preview
candidate export to bookbind.yaml
convert --interactive --select
provider response cache
versioned CLI release artifacts
```

### MVP v0.4

```text
Wails desktop shell
React/TypeScript frontend
Windows/macOS/Linux target
app bridge over internal/app use cases
import screen: MP3, директория MP3, cover, metadata.yaml
metadata search screen: providers, candidate table, preview, select
convert screen: output path, dry-run, progress/log
cache screen: list/clean
desktop CI smoke checks
```

### MVP v0.5

```text
EPUB TOC import
EPUB chapters → audio files matching
manual chapter matching UI
LitRes/MyBook experimental
duration matching
narrator matching
лучший ru-profile
history/library
```

---

## 17. Desktop UI

Desktop-приложение нужно считать основным интерфейсом для пользователя.
Пользователь не должен быть вынужден понимать CLI-флаги, формат ffmpeg-команд
или внутреннюю структуру metadata.yaml.

Рекомендуемый стек:

```text
Go backend
Wails desktop shell
React/TypeScript frontend
```

Причины:

```text
- backend остается на Go
- можно переиспользовать internal/app и core-модули
- есть сборка под Windows/macOS/Linux
- frontend можно сделать удобным: таблицы, формы, drag-and-drop, progress
```

Пример структуры:

```text
cmd/
  bookbind/
    main.go

internal/
  app/
    convert.go
    inspect.go
    search.go
    template.go
    jobs.go

  desktop/
    bridge.go
    events.go

wails.json
frontend/
  package.json
  src/
    App.tsx
    screens/
      Import.tsx
      Inspect.tsx
      Metadata.tsx
      Chapters.tsx
      Build.tsx
```

Основной пользовательский flow:

```text
1. Import
   выбрать MP3, папку MP3, EPUB, cover, metadata.yaml

2. Inspect
   показать файлы, длительности, теги, проблемы входа

3. Metadata
   поиск кандидатов, сравнение, ручная правка, итоговый preview

4. Chapters
   chapters from files / EPUB TOC / synthetic / ручная правка

5. Build
   dry-run preview, output path, progress, готовый .m4b
```

Desktop UI вызывает только use cases:

```go
InspectInput(ctx, request)
SearchMetadata(ctx, query)
ResolveMetadata(ctx, request)
PreviewMetadata(ctx, request)
TemplateMetadata(ctx, request)
Convert(ctx, request)
ListCache(path)
CleanCache(path)
PreviewChapters(ctx, request)
BuildM4B(ctx, request)
```

### v0.4.0 desktop shell plan

Первый desktop-релиз не должен пытаться сразу заменить всю CLI-функциональность.
Цель `v0.4.0` — открыть приложение, связать frontend с Go backend и провести
пользователя через основной audiobook flow без ручного набора CLI-флагов.

Технический выбор:

```text
- Wails
- React
- TypeScript
- Go bridge в internal/desktop
- переиспользование internal/app без дублирования бизнес-логики
```

Первый PR со scaffold:

```text
- [x] Wails project files
- [x] frontend shell
- [x] backend bridge AppVersion
- [x] навигация Import / Metadata / Convert / Cache
- [x] локальная команда запуска desktop dev mode
```

Следующие PR:

```text
1. [x] Import screen: input path inspect, files, durations, embedded tags
2. [x] Import screen: file/folder picker, cover, metadata.yaml
3. [x] Metadata screen: providers, search, preview, select/export
4. [x] Convert screen: output path, dry-run, convert log
5. [x] Cache screen: list/clean
6. [x] CI smoke для frontend lint/build
7. [x] release workflow для desktop artifacts
```

UI не должен напрямую знать про ffmpeg, scoring, merge-правила, парсинг EPUB
или структуру провайдеров.

---

## 18. EPUB и главы

EPUB нужно использовать как источник структуры книги. Он хорошо дает порядок и
названия глав, но сам по себе не дает точные аудио-таймкоды.

Реалистичный MVP:

```text
- читать TOC из EPUB
- убирать служебные разделы: cover, contents, copyright, notes
- сопоставлять главы EPUB с MP3-файлами по порядку
- если количество глав и файлов совпадает — предлагать match 1:1
- если не совпадает — показывать UI для ручной коррекции
```

Для директории MP3 это особенно полезно:

```text
Audio files                 EPUB chapters
01.mp3  31:20               Пролог
02.mp3  28:44               Глава 1
03.mp3  34:01               Глава 2
```

Для одного большого MP3 автоматическое получение точных глав требует
дополнительного слоя:

```text
- speech-to-text
- fuzzy matching текста EPUB и транскрипта
- forced alignment
- ручное подтверждение подозрительных мест
```

Это стоит рассматривать как отдельный будущий модуль, не как обязательную часть
первого релиза.

---

## 19. Разработка в ветках

Разработка идет небольшими ветками под конкретные изменения.

Базовые правила:

```text
- main всегда должен собираться
- каждая фича делается в отдельной ветке
- ветки должны быть небольшими и проверяемыми
- перед merge обязательно проходят тесты
- крупные изменения разбиваются на core, CLI, UI, tests
```

Пример веток:

```text
feature/core-convert
feature/desktop-shell
feature/metadata-search
feature/epub-toc
feature/chapter-matching-ui
ci/build-and-test
```

Для релизов начиная с `v0.3.0` используется интеграционная релизная ветка:

```text
main
  ↑
release/v0.3.0
  ↑
feature/search-providers
feature/interactive-selection
feature/cache
```

Правила:

```text
- feature-ветки создаются от release/v0.3.0
- feature-ветки мержатся в release/v0.3.0
- main получает только готовый release/vX.Y.Z
- версия в pkg/version обновляется в финальном release PR
```

---

## 20. CI

CI нужен обязательно, даже если не в самый первый день. Он должен проверять, что
приложение собирается и тесты не сломались.

Минимальный CI:

```text
- go fmt / gofmt check
- go vet
- go test ./...
- сборка CLI
- сборка desktop-приложения
```

Кроссплатформенная матрица:

```text
- ubuntu-latest
- macos-latest
- windows-latest
```

Позже добавить:

```text
- integration tests с короткими test MP3
- проверка ffmpeg/ffprobe, если доступны
- frontend lint/test
- [x] release workflow для CLI Windows/macOS/Linux
- [x] Wails release artifacts для desktop UI
```

CI не должен зависеть от внешних metadata providers. Провайдеры тестируются через
моки/fixtures, а реальные сетевые проверки можно вынести в отдельный optional
workflow.

---

## 21. Тестовая стратегия

Тесты обязательны. Цель — быстро понимать, что конвертация, главы, метаданные и
UI-bridge не сломались после изменений.

Unit tests:

```text
- filename parser
- ru normalizer: е/ё, пунктуация, пробелы, порядок имени
- scoring candidates
- metadata merge priorities
- YAML read/write
- chapter generation from files
- synthetic chapters
- EPUB TOC parsing
- EPUB → audio chapter matching
- output path generation
```

Integration tests:

```text
- inspect короткого MP3
- convert одного короткого MP3 в M4B
- convert директории MP3 с chapters from files
- convert с metadata.yaml и cover.jpg
- dry-run не создает output
- overwrite protection работает
```

Provider tests:

```text
- providers работают через recorded fixtures
- сетевые ошибки не ломают convert
- несколько кандидатов с одним названием требуют выбора
- auto-select работает только выше confidence threshold
```

Desktop tests:

```text
- backend bridge вызывает правильные use cases
- long-running build отдает progress events
- cancel build корректно останавливает job
- UI не хранит бизнес-логику, только состояние экрана
```

Golden/fixture files:

```text
testdata/
  audio/
    short-1.mp3
    short-2.mp3
  covers/
    cover.jpg
  metadata/
    valid.bookbind.yaml
    invalid.bookbind.yaml
  epub/
    simple.epub
  providers/
    googlebooks.search.json
    openlibrary.search.json
```

---

## 22. Что сделать первым

Самый правильный первый кусок — не API, а базовый конвертер:

```text
1. core use cases
2. CLI как тонкая оболочка
3. ffprobe
4. ffmpeg
5. chapters from files
6. metadata.yaml
7. cover.jpg
8. первые unit/integration tests
```

Потому что уже после этого программа будет полезной:

```bash
bookbind convert "./Аудиокниги/Сергей Лукьяненко/Дозоры/01 - Ночной дозор" \
  --metadata bookbind.yaml \
  --output "Сергей Лукьяненко - Дозоры 01 - Ночной дозор.m4b"
```

После этого нужно поднимать desktop shell, чтобы вся дальнейшая работа
проверялась уже через реальный пользовательский flow.
