---
name: cross-platform-release
description: Validate and package Yumdown across its declared desktop OS/architecture targets, including Wails builds, media-tool assets, CI, signing constraints, and release smoke tests.
---

# Cross-Platform Validation and Release

Use this skill for CI, installer, packaging, release, or platform-support work. Read `implementation.md` and inspect the repository's actual target matrix and build setup first.

## Target and build matrix

- Confirm the supported OS/architecture matrix before changing packaging. Record which targets are release-supported versus experimental.
- Verify Wails, Go, Node/PNPM, and native build prerequisites for every target. Use only commands supported by the current scaffold and CI environment.
- Check platform-specific filesystem behavior, executable permissions, process termination, file-manager opening, path length/Unicode handling, and application data locations.
- Keep platform-specific code behind small, testable adapters and avoid claiming support based only on compilation on another OS.

## Media tool artifacts

- For every artifact, verify the expected OS/architecture, version, checksum/signature where available, executable bit, and startup/capability check.
- Review licensing and redistribution terms, artifact sources, update behavior, installer size, and signing/notarization implications before release.
- Test missing/corrupt/incompatible tools and network failure during first-run setup. Ensure failures are actionable and do not cause fallback to an arbitrary executable.
- Advertise only formats that the resolved `ffmpeg` build can encode.

## CI, test, and release evidence

- Run formatting, lint, unit/integration tests, and available platform builds as separate checks. Keep routine CI deterministic and offline; mark website-based downloads as explicit authorized manual smoke tests.
- Exercise install/launch, first-run tool setup, batch queue behavior, output verification, cancellation, history persistence across restart, upgrade, and uninstall on clean environments where possible.
- Include negative cases: invalid/unreachable URLs, unsupported formats, unwritable destination, existing filename, missing tools, interrupted process, and disk/write failures.
- Report results by command and platform. Separate confirmed blockers, untested platform behavior, and follow-up work; never claim a release check passed without evidence.
