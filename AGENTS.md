# AGENTS.md — go-ffmpeg-tools

Go 1.26 + Wails v2.16.0 (Wails CLI pinned to v2.16.0) desktop app, with a Vue3 + TS + Vite + pnpm frontend. Go module name is `go-ffmpeg-tools`.

## Common commands
- Build/package: `./build.ps1` (current platform), `-DebugBuild` for a debug binary, `-Platform "windows/amd64 linux/amd64"` for cross-platform. The script auto-installs the Wails CLI, runs `pnpm install --frozen-lockfile`, `wails build -clean`, and copies ffmpeg from `build/ffmpeg/<goos>/` into the output directory.
- Backend checks only: `go build ./...`, `go vet ./...` (no Go tests in the repo yet).
- Frontend checks only: run `npm run build` in `frontend/` (includes `vue-tsc --noEmit` type checking).
- Dev hot-reload: `wails dev`.
- App icons: `./generate-icon.ps1` rasterizes `frontend/public/favicon.svg` into `build/appicon.png` (1024px) and `build/windows/icon.ico`. Uses ImageMagick → Inkscape → headless Edge/Chrome (first available); without ImageMagick the ICO is assembled via System.Drawing. Run it after changing the SVG.

## Architecture
- `main.go`: wails.Run, `//go:embed all:frontend/dist` embeds the frontend bundle; `Bind` exposes the `App` methods from `app.go` to the frontend.
- `internal/task`: Job model and a concurrency-safe task Manager (cancel, in-order listing).
- `internal/ffmpeg`: ffmpeg argument building and process execution; `builder.go` handles the two command paths — soft subtitles (multi-track mov_text) and hard burn-in.
- `frontend/src`: App.vue is the sidebar shell (menu switching, no vue-router); views/Index.vue is the subtitle tools page; i18n uses vue-i18n with locale packs in `frontend/src/i18n/locales/{zh-CN,en,ja}.ts`.

## Non-obvious gotchas
- `frontend/wailsjs/` contains Wails-generated bindings; do not edit by hand. Model classes generated there (`main.SubmitJobReq`, `task.SubConfig`, etc.) must be constructed with `new` — plain literals are rejected by the convertValues runtime.
- After adding App methods, regenerate bindings via `wails generate module` (or `wails dev`/build).
- File drag-and-drop: `OnFileDrop` is called with `useDropTarget=false`; the frontend resolves the drop zone manually via `data-drop-zone` attributes + `document.elementFromPoint`. The CSS property for the default highlight that is disabled is `--wails-drop-target`.
- ffmpeg binary lookup order at runtime: `build/ffmpeg/<GOOS>/ffmpeg(.exe)` next to the executable → PATH.
- On Windows the ffmpeg child process must hide its console window (`CREATE_NO_WINDOW` in `runner_windows.go`); platform differences are split via build tags.
- Subtitle language codes like `chs/cht` are mapped to ISO 639-2 by `builder.normalizeLang`; the subtitles filter path for hard burn must go through `escapeFilterPath` (Windows drive-letter colons, etc.).
- Any new UI copy must be added to all three locale packs (zh-CN/en/ja); fallback locale is zh-CN.
- In `wails.json`, `frontend:install` uses pnpm and `frontend:build` uses npm — do not mix or change them.
- `docs/history/` holds historical design notes for background only; build artifacts live in `build/bin` and must not be committed.

## Instruction-file maintenance rules
When project structure, build/test commands, architecture boundaries, development conventions, or any other fact recorded in this file changes, update this file in the same change: replace outdated statements with current facts, and do not append a running changelog. Keep only current rules, commands, architecture boundaries, and non-obvious gotchas here; detailed incident retrospectives, performance comparisons, experiment logs, and completed migration records belong in linked standalone documents, read only when relevant to a task. Before migrating or deleting existing content from this file, propose a plan and get explicit user consent; never silently summarize or truncate existing instructions.
