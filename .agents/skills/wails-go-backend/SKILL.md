---
name: wails-go-backend
description: Implement and test Yumdown's Wails/Go backend, including typed bindings and events, download job lifecycle, cancellation, local persistence, and safe process invocation.
---

# Wails and Go Backend

Use this skill when implementing backend services or Wails integration for Yumdown. Follow the repository's `implementation.md` and inspect the existing scaffold before applying these patterns.

## Architecture

- Keep Wails-bound methods in a thin application layer. Put validation, media execution, queue logic, and persistence in Go packages that can be tested without Wails.
- Define explicit request, response, event, job-state, settings, and history types. Use stable job IDs and make state transitions deliberate.
- Use repository/service interfaces where they make persistence or process behavior testable; avoid unnecessary abstraction for trivial code.
- Keep generated Wails bindings at the frontend integration boundary and update/regenerate them using the repository's actual scripts.

## Process and queue safety

- Start external tools with `exec.CommandContext` and an argument slice. Never invoke a shell or interpolate user-provided URLs, paths, or filenames into command text.
- Validate request fields on the backend regardless of frontend validation. Restrict formats and advanced options to supported allowlisted values.
- Resolve the expected packaged or app-managed executable for the current OS/architecture; do not silently run an arbitrary `PATH` binary.
- Bound concurrent jobs, track states such as queued/downloading/converting/completed/failed/cancelled, and make cancellation and application shutdown terminate child processes predictably.
- Prefer machine-readable downloader progress. Normalize known values and represent missing totals as indeterminate. Do not report an unweighted average as byte-weighted total progress.
- Ensure output paths are deterministic and safe, avoid accidental overwrite by default, verify the final file, and calculate its size only after successful completion.
- Limit routine process logging and redact sensitive URL query values where practical.

## Persistence and events

- Persist settings and terminal history locally; keep media files out of the database. Keep migrations and repository operations testable.
- Record URL, destination, timestamp, status, selected format, output size when available, and useful error summary. Do not report failed/cancelled work as successful.
- Emit events with stable job IDs and typed payloads. Avoid making event delivery the only source of truth; provide a way to query current jobs after frontend reconnect/reload.

## Validation

- Add deterministic tests for validation, state transitions, queue limits, progress parsing, cancellation, tool resolution, output naming, persistence, and error mapping as relevant.
- Use fake executables or injected process interfaces for process tests. Keep tests offline by default.
- Run focused Go tests and project lint/build commands that actually exist; report exact results and platform limitations.
