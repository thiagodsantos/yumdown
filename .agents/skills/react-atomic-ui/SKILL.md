---
name: react-atomic-ui
description: Build Yumdown's accessible React/TypeScript desktop interface with Base UI, Tailwind CSS, atomic design, typed Wails integration, and reliable download/history states.
---

# React, Base UI, and Atomic Design

Use this skill for Yumdown frontend components, pages, state, or styling. Follow `implementation.md` and inspect the installed Base UI, React, and Wails binding versions before using APIs.

## Component organization

- Group code by feature where practical (`downloads`, `history`, `settings`) and classify reusable UI by responsibility: `atoms`, `molecules`, `organisms`, `templates`, and `pages`.
- Keep atoms focused on presentation/accessibility; molecules combine small controls; organisms own cohesive sections; templates arrange page regions; pages connect app state and feature flows.
- Use Base UI primitives and documented composition patterns. Avoid wrapping every primitive without adding value and avoid introducing an additional component framework without a requirement.
- Use Tailwind consistently with existing project conventions. Keep component state and domain behavior out of styling-only atoms.

## Wails integration and state

- Treat generated Wails bindings and backend event/request types as authoritative. Do not call `yt-dlp`, `ffmpeg`, shell commands, or filesystem operations directly from the frontend.
- Subscribe/unsubscribe to Wails events using the project's actual runtime APIs. Match updates by stable job ID and query current backend state when needed after initial load or reconnect.
- Model queued, active, completed, failed, and cancelled states explicitly. Prevent duplicate batch submissions and handle event/listener cleanup.
- Present per-job progress and honest aggregate progress. Use indeterminate UI when totals are unknown; do not average per-job percentages and label the result byte-weighted.
- Avoid displaying sensitive full query strings in history. Present clear user-actionable errors without exposing raw process logs indiscriminately.

## UX and accessibility

- Provide required/optional labels and per-URL validation feedback for the multi-URL form; expose busy and disabled states while submitting.
- Include destination selection, selected path, output format, open-folder action, per-item queue status, overall queue status, and empty/loading/error states.
- Ensure keyboard access, visible focus, semantic labels, and screen-reader-friendly progress/status messages. Confirm destructive history actions.
- Do not claim a format is available unless the backend reports it as supported by the resolved media tools.

## Validation

- Add focused tests for URL/form validation, repeated-submit prevention, event-driven status/progress rendering, history filters, and empty/error states using the repository's existing test stack.
- Run the existing frontend lint, formatting check, tests, and build. Report commands and results; do not invent successful checks.
