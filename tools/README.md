# User-managed media tools

This folder is the packaged-resource location checked first by Yumdown. Do not commit third-party executable binaries here unless the repository has an explicit redistribution decision, verified upstream license review, provenance/integrity checks, and platform packaging approval.

For local development, put the matching `yt-dlp` and `ffmpeg` executables in this directory, in the app configuration directory's `Yumdown/tools` subdirectory, or set `YUMDOWN_TOOLS_DIR` to an absolute tools directory. Yumdown does not fall back to `PATH` and does not auto-download or auto-update these tools.
