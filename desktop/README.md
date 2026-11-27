# bookbind desktop

Wails desktop shell for bookbind.

## Development

Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1
```

Run the desktop app in development mode:

```bash
cd desktop
wails dev
```

Build a local desktop package:

```bash
cd desktop
wails build
```

The frontend is a React/TypeScript app in `frontend/`. The Go bridge lives in
`app.go` and reuses the root module through the local `replace` directive in
`go.mod`.

## Release packages

Release builds use `wails build -ldflags "-X main.bundledFFmpeg=1"` and ship
FFmpeg 9.1 and its matching `ffprobe`. The Go bridge passes absolute paths to
both tools. The package layout is `bin/` next to the executable on Linux and
Windows, and `Contents/Resources/bin/` inside the macOS app.

Run `bash desktop/build/bundle-ffmpeg.sh DESTINATION` from the repository root
after the Wails build. Use `desktop/build/bin` as the destination on Linux or
Windows and `desktop/build/bin/bookbind-desktop.app/Contents/Resources` on
macOS. The script needs Bash, curl, xz, make and a native C toolchain, with NASM
on x86-64. Use MSYS2 UCRT64 on Windows. On macOS, sign the tools and then the
app after inserting the bundle, as shown in the release workflow.

The script fixes the upstream source version and SHA-256. It produces an
LGPL 2.1-or-later build of FFmpeg and includes the matching source archive,
license and configure flags in `ffmpeg-source/`. Keep that directory in the
release package. When updating FFmpeg, verify the upstream release signature
and update the version and SHA-256 together.
