# Release Notes

## v0.4.2

`v0.4.2` improves the desktop conversion flow after the first release.

### Added

- Live ffmpeg conversion progress in the desktop Convert screen.
- A conversion log panel that shows ffmpeg output while conversion is running.
- English/Russian language selector for the desktop UI.

### Fixed

- Google Books remote cover URLs in metadata no longer break conversion by being
  treated as local file paths.
- Desktop conversion errors include ffmpeg stderr instead of only an exit code.
- Empty desktop output path now defaults next to the input instead of using a
  confusing relative `book.m4b`.

## v0.4.0

`v0.4.0` introduces the first desktop release on top of the existing Go core.

### Added

- Wails desktop app scaffold for Windows, macOS, and Linux.
- React/TypeScript frontend shell with Import, Metadata, Convert, and Cache
  screens.
- Import screen with audio file/folder picker, input inspection, durations,
  embedded tags, chapter counts, metadata YAML picker, and cover picker.
- Metadata screen with provider selection, candidate search, preview, and
  metadata YAML export.
- Convert screen with input/output fields, output save dialog, metadata/cover
  paths, chapter interval, overwrite control, dry-run preview, real conversion,
  and command/status log.
- Cache screen with cache path selection, entry listing, size summary, and clean
  action.
- Desktop CI smoke job for frontend build, desktop Go vet/test, and Wails build.
- Release workflow artifacts for desktop builds on Linux, macOS, and Windows.

### Changed

- The release workflow now produces both CLI and desktop artifacts.
- The default release workflow version is `v0.4.0`.
- `bookbind version` reports `0.4.0` for source builds on this release branch.

### Notes

- Real inspection and conversion still require `ffmpeg` and `ffprobe` on
  `PATH`.
- Linux desktop CI installs GTK/WebKit dependencies required by Wails; runner
  package installation can still be slower than pure Go jobs.
- Desktop artifacts are unsigned first-release builds.

## v0.3.0

`v0.3.0` focuses on metadata discovery, candidate selection, caching, and
release automation for the CLI.

### Added

- Open Library metadata provider.
- Google Books metadata provider.
- `bookbind providers` to list available metadata sources.
- `bookbind search --provider ...` to limit metadata search sources.
- Candidate scoring and tabular search output.
- `bookbind search --select N --output bookbind.yaml` to save metadata from a
  selected search candidate.
- `bookbind metadata --provider ... --id ... --preview` to inspect a candidate.
- `bookbind metadata --provider ... --id ... --output bookbind.yaml` to export a
  selected candidate.
- `bookbind convert --interactive --select N` to search and save metadata before
  conversion.
- Local provider response cache.
- `bookbind cache list` and `bookbind cache clean`.
- Versioned release workflow for Linux, macOS, and Windows CLI artifacts.

### Changed

- Release binaries include the version in the filename.
- `bookbind version` can be overridden by the release workflow from the release
  tag.
- CI now runs for `release/**` and `ci/**` branches.

### Notes

- Real conversion still requires `ffmpeg` and `ffprobe` on `PATH`.
- Desktop UI work remains planned after the CLI/core flow is stable.
