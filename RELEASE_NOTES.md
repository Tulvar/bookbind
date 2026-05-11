# Release Notes

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
