# bookbind desktop

Wails desktop shell for bookbind.

## Development

Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
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
