# Yumdown Implementation Plan

## Goal

Build a cross-platform desktop app that extracts audio from one or more video URLs, shows per-item and aggregate progress, saves outputs in the chosen format and folder, and keeps a searchable download history.

## Technical baseline

- **Desktop shell/backend:** Wails with Go. Go owns filesystem access, process execution, settings/history persistence, and download lifecycle; the frontend never invokes shell commands directly.
- **Frontend:** React + TypeScript + Vite, styled with Tailwind CSS and Base UI components/patterns. Use PNPM, Prettier, and ESLint.
- **Frontend organization:** Atomic design (`atoms`, `molecules`, `organisms`, `templates`, `pages`) within a feature-oriented structure where practical. Keep Wails-generated bindings at the integration boundary.
- **Downloader:** Invoke the platform-appropriate `yt-dlp` executable as a child process using an argument array (never a shell command). Use `ffmpeg`/`ffprobe` where required for extraction, conversion, or metadata.
- **Local data:** Persist settings and completed/failed history locally. Use a small embedded database (SQLite) behind a Go repository interface; do not store media files in the database.

## Proposed application structure

Adapt names to the Wails project generator and current Wails version during scaffolding:

```text
frontend/
  src/
    app/                 # app shell, routing, shared providers
    components/
      atoms/
      molecules/
      organisms/
      templates/
    features/
      downloads/         # form, queue, progress, download state
      history/           # history table, filters, detail actions
      settings/          # defaults and advanced options
    types/
    styles/
backend (Go)/
  app/                   # Wails-bound application methods
  downloads/             # queue, job lifecycle, progress parsing
  media/                 # yt-dlp/ffmpeg resolution and invocation
  history/               # persistence and queries
  settings/              # settings persistence and validation
  events/                # typed event payloads
```

Keep the downloader and persistence packages independent of Wails so they can be tested without launching the desktop UI.

## Phased implementation

### 1. Confirm scope and scaffold

- Choose and record the initial release OS/architecture matrix (at minimum decide whether Windows, macOS, and Linux are all in the first release).
- Select a supported Wails version and verify its requirements for Go, Node, and platform build tools.
- Scaffold the Wails + React + TypeScript + Vite app; configure PNPM scripts, Tailwind, Prettier, ESLint, and Base UI.
- Establish formatting/lint/test/build scripts and a minimal CI workflow for the chosen target platforms.
- Add a short README covering development prerequisites and how to run the app.

**Exit criteria:** clean development build starts the Wails window; frontend lint/format checks and Go tests run in CI.

### 2. Define contracts and core backend services

- Define typed Go structures for download requests, job state/progress, download results, history entries, settings, and errors.
- Add Wails-bound methods for starting/stopping downloads, querying jobs/history/settings, selecting a destination directory, and opening a directory in the OS file manager.
- Emit typed Wails events for job state and progress updates. Include stable job IDs so frontend updates can be matched to queue items.
- Validate URL, filename, output format, destination, and advanced options on the backend, even when the frontend also validates them.
- Use Go context cancellation and process termination for app shutdown and explicit cancellation; ensure temporary resources are cleaned up.

**Exit criteria:** backend contract tests cover validation and state transitions; frontend can call a harmless backend method and receive a typed event.

### 3. Resolve and package media tools

- Implement a tool resolver keyed by operating system and architecture. Resolve paths from the packaged application resources first, then an explicitly configured tools directory; never silently run an unrelated system binary.
- Define the distribution/update strategy for `yt-dlp`, `ffmpeg`, and `ffprobe` (bundled assets versus first-run managed download). Verify licensing, redistribution requirements, executable permissions, and signing/notarization implications before bundling.
- If tools are downloaded on first run, use official release sources, verify checksums/signatures where available, download over TLS, store in an app-owned directory, and provide clear failure/retry messages. Do not auto-update without a user-visible policy.
- Add a capability check that reports missing/incompatible tools before a job starts.
- Build a small adapter around `yt-dlp` arguments and process output. Pass each option as a separate argument; avoid shell interpolation. Limit logged output and redact sensitive URL query data where practical.

**Exit criteria:** resolver tests cover supported OS/architecture mappings and missing tools; an integration test with a local fake executable verifies argument passing and process errors.

### 4. Implement the download queue and progress

- Support multiple submitted URLs, assigning each a job ID and tracking states such as queued, downloading, converting, completed, failed, and cancelled.
- Run jobs with a configurable bounded worker count; avoid launching unbounded simultaneous processes. Preserve submission order in the UI while allowing jobs to complete independently.
- Configure `yt-dlp` for audio extraction and the selected output format. Use a safe output template, honor an optional filename, avoid accidental overwrite by default, and report the final output path.
- Parse machine-readable progress from `yt-dlp` (prefer its progress-template/structured output options over fragile human-readable strings). Normalize bytes, percentage, speed, and ETA when available.
- Emit per-job progress and derive aggregate progress from known item progress. Clearly show indeterminate aggregate progress when item sizes/progress are unavailable; do not imply that a sum of percentages is byte-weighted.
- Capture exit status and useful error text. On completion, confirm the output file exists and calculate its size; record failed/cancelled jobs distinctly.
- Add cancellation per job and define behavior for closing the app while jobs run.

**Exit criteria:** tests cover queue limits, progress parsing, cancellation, failures, output naming, and aggregate progress; a manual test downloads a permitted test video and produces a valid audio file.

### 5. Build the primary download experience

- Create the main download page with a required multi-URL input, optional filename, destination-folder picker, output-format selector, and a clear start action.
- Allow one URL per line (and optionally comma-separated input only if parsing can be unambiguous); show validation per entered URL before submission.
- Show the selected destination and an **Open folder** action. Use the native OS file manager through a backend method, with validated paths.
- Display queued/running/completed/failed items with per-item progress, status, output name, and actionable errors. Include overall queue progress and an empty state.
- Prevent duplicate submissions caused by repeated clicks and provide disabled/loading states while requests are being accepted.
- Build reusable Base UI-based atoms and compose them into molecules/organisms; keep page-level orchestration out of presentational components.

**Exit criteria:** the primary flow can submit a batch, track each job, open the output folder, and recover from validation and runtime errors.

### 6. Add history and settings

- Persist each terminal job with URL, output folder, completion timestamp, status (`ok`/`error`, with cancellation represented explicitly or mapped consistently), output file size when available, selected format, and error summary when relevant.
- Provide a History tab with newest-first ordering, useful filters/search, and clear empty/loading/error states. Avoid rendering full private query strings unnecessarily.
- Add settings for defaults such as output folder, output format, and concurrency. Validate and persist settings on the backend.
- Add an Advanced tab for carefully scoped options (for example, audio quality/bitrate, embedding metadata/thumbnail, and tool diagnostics). Use safe allowlisted settings rather than exposing arbitrary command-line text in the initial release.
- Define retention/clear-history behavior and confirm destructive actions with the user. Deleting history must not delete media files unless explicitly requested.

**Exit criteria:** history survives app restarts, matches completed job outcomes, and settings are restored on restart.

### 7. Harden, package, and release

- Add unit tests for validation, persistence, tool resolution, progress parsing, queue state transitions, and error mapping; add frontend tests for form validation and key UI states.
- Add end-to-end/manual test cases for single and batch downloads, unsupported formats, invalid/unreachable URLs, missing tools, unwritable destinations, cancellation, and app restart.
- Test path/filename handling across OSes, Unicode names, existing files, long paths, and disk-full/write failures. Ensure subprocesses are terminated when requested and no untrusted input is interpreted by a shell.
- Package each supported OS/architecture with the correct executable assets and permissions. Test clean-machine install, launch, first-run tool setup, upgrade, and uninstall behavior.
- Document supported formats, tool licensing/update behavior, network requirements, limitations, and troubleshooting. Do not claim support for a format unless the installed `ffmpeg` build can encode it.

**Exit criteria:** release artifacts are built and smoke-tested on every declared target; CI runs lint, tests, and available platform builds.

## Key acceptance criteria

- A user can submit one or many valid video URLs, select an output folder and audio format, and see an independent progress/status for every item plus an honest queue-level summary.
- Completed files are present at the selected destination, use the requested names/formats where supported, and have a recorded file size.
- Failed and cancelled jobs give actionable feedback and do not appear as successful history entries.
- History and settings persist between launches; opening the destination folder works on each supported OS.
- The app uses the correct `yt-dlp`/`ffmpeg` tools for its platform, never constructs shell commands from user input, and handles unavailable tools clearly.
- The UI follows Base UI conventions and keeps reusable components organized according to atomic design.

## Decisions and risks to resolve early

1. **Supported targets and architectures:** determine the first-release OS/architecture matrix before implementing packaging and tool resolution.
2. **Media-tool distribution:** bundling simplifies offline use but introduces per-platform artifact, licensing, size, and signing work; first-run downloads require reliable verification, network handling, and an update policy.
3. **Format support:** output availability depends on the bundled `ffmpeg` encoders. Present only formats the installed build supports and explain any format-specific constraints.
4. **Progress semantics:** some sources do not provide reliable total size or duration. Support indeterminate progress and avoid presenting estimated values as exact.
5. **URL privacy and security:** treat URLs as untrusted input, avoid shell execution, minimize sensitive URL logging, and make clear that downloading content is subject to the user's rights and the source site's terms.
6. **Advanced options:** arbitrary `yt-dlp` flags can introduce unsafe behavior and complicate support. Start with a validated allowlist and expand only with explicit product requirements.
