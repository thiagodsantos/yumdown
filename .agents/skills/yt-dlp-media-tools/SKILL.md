---
name: yt-dlp-media-tools
description: Integrate yt-dlp and ffmpeg in Yumdown safely, including OS/architecture tool resolution, argument construction, progress parsing, output verification, and distribution checks.
---

# yt-dlp and Media Tool Integration

Use this skill for downloader, transcoding, tool discovery, or media-format work in Yumdown. Read `implementation.md` and inspect the actual selected tool versions before relying on flags or output formats.

## Tool resolution and distribution

- Resolve `yt-dlp`, `ffmpeg`, and `ffprobe` by the app's declared OS/architecture matrix. Prefer explicit packaged or app-managed locations; do not silently execute an unrelated system binary.
- Establish the distribution/update policy before adding binaries. Check upstream licensing and redistribution terms, artifact provenance, integrity verification, executable permissions, app signing/notarization, storage location, network failure behavior, and user-visible update policy.
- For managed downloads, use official sources over TLS and verify checksums/signatures where available. Store tools in an app-owned directory with restrictive, appropriate permissions; show clear missing-tool and retry errors.
- Probe capabilities from the resolved tools rather than assuming every build supports every format or encoder.

## Invocation and output

- Invoke tools as child processes using a context and separate arguments; never construct shell strings from user input.
- Treat URL, filename, destination, format, and advanced options as untrusted. Validate them in Go, use an allowlist for optional settings, and use safe output templates with overwrite behavior explicitly controlled.
- Do not implement access-control, authentication, DRM, or other content-protection circumvention. Make the product's terms/rights expectations clear and use only content the user is authorized to download.
- Verify output existence and format after successful process completion; derive output size from the resulting file. Distinguish download, conversion, and terminal failure states.

## Progress and errors

- Prefer a documented machine-readable progress template/structured output supported by the chosen yt-dlp version, not scraping human-oriented status lines.
- Test parser behavior against captured representative output from the actual supported version. Normalize percentage, byte counts, speed, and ETA only when present and valid.
- Treat unknown total size as indeterminate; avoid inventing a precise aggregate. Keep diagnostic logs useful while redacting sensitive URL query values where practical.
- Map common failures (missing tool, unsupported format, inaccessible destination, process cancellation, network/source failure) to actionable app errors without exposing arbitrary process output as UI markup.

## Testing

- Unit-test OS/architecture mapping, capability checks, argument construction, parser edge cases, and error mapping.
- Use a fake local executable to verify argument boundaries, cancellation, exit codes, and output handling. Keep normal automated tests independent of external websites and network access.
- Run permitted end-to-end tests only against a source/content the tester is authorized to access, and label them as network-dependent.
