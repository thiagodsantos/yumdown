# Yumdown

Yumdown is a Wails desktop app for extracting audio from video URLs. The initial implementation targets local Linux/amd64 development; Windows and macOS tool-name/path resolution is included, but release artifacts on those platforms are not yet validated.

## Requirements

- Go 1.24 or newer
- Node.js 22 or newer and PNPM 12.23.0 (the version pinned by the frontend package)
- Wails v2.10.2 CLI
- Linux development libraries required by Wails (GTK 3, WebKit2GTK, `pkg-config`, and a C toolchain)
- `yt-dlp` and `ffmpeg` matching your OS and architecture to perform downloads

Install dependencies and the pinned Wails CLI:

```sh
make install-deps
```

This installs Go/frontend modules and Wails, installs `yt-dlp` and `ffmpeg` through the Linux package manager, then links those tools into Yumdown's app-config tools directory. Linux package installation uses `sudo`. Fedora 40+ uses WebKitGTK 4.1; the Makefile detects it and adds Wails' `webkit2_41` build tag automatically. On Debian/Ubuntu it uses WebKitGTK 4.0. On macOS, install Xcode Command Line Tools and media tools first.

## Run and build the desktop app

```sh
make dev
make build
```

You can also run `wails dev -tags webkit2_41` and `wails build -tags webkit2_41` manually on Fedora 40+.

The frontend can also be checked separately with `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm format:check`, and `pnpm build` from `frontend/`. Run backend tests from the repository root with `go test ./...`.

## Releases

Push a semantic-version tag such as `v0.1.0` to start a release. The release workflow runs the CI checks first, then builds a Linux/amd64 app archive and publishes it to the repository's GitHub Releases page. Before tagging, add release notes under the matching heading in `CHANGELOG.md`, for example `## [0.1.0] - YYYY-MM-DD`. The workflow uses that version's section as the GitHub release notes.

## Media tool setup

Yumdown does not download or update third-party executables automatically. Put executable `yt-dlp` and `ffmpeg` binaries in one of these locations:

1. A `tools/` directory next to the Yumdown executable (packaged-resource location).
2. An absolute directory set by `YUMDOWN_TOOLS_DIR` for development/testing.
3. The app configuration directory's `tools/` folder (`os.UserConfigDir()/Yumdown/tools`) as the user-managed fallback.

On Windows, use `yt-dlp.exe` and `ffmpeg.exe`; other supported OSes use `yt-dlp` and `ffmpeg`. Linux/macOS binaries need their executable bit set. `ffprobe` is optional. Yumdown will not silently execute a system binary found on `PATH`.

The app probes the installed `ffmpeg` encoders and only presents formats it detects. If the encoder build, tool versions, or architecture do not match, check **Advanced → Media tools** for the diagnostic message. Obtain tools from official sources and review their current licenses and terms before redistribution or use.

## Data and privacy

Settings and download history are stored locally in an SQLite database under the app's configuration directory. Media files are saved only to the chosen output folder. Routine history display hides URL query strings, though the local database records the submitted URL to support troubleshooting and history search. Download only content you are authorized to save and follow the source's terms.

## Current scope and release notes

- The media-tool distribution policy for releases is not yet selected. The initial build uses user-managed binaries; it does not bundle or auto-update them.
- Linux/amd64 is the only platform available for local validation in the current development environment. Other OS/architecture mappings are implementation scaffolding and are not a support claim.
- App signing, notarization, installers, and clean-machine smoke tests remain release work.
