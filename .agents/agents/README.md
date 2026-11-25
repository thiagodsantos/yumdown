# Yumdown Agent Roles

These role prompts are reusable instructions for delegating work on Yumdown. They are intentionally stored as Markdown because this repository has no existing agent-profile configuration format. To use one, provide the role file and the relevant task/paths as context when starting a subagent; do not assume the editor automatically registers these files as selectable profiles.

## Roles

- [`implementation-coordinator.md`](implementation-coordinator.md) — break work into ordered, independent slices and coordinate integrations.
- [`go-desktop-backend.md`](go-desktop-backend.md) — implement Wails/Go services, job lifecycle, persistence, and backend contracts.
- [`react-desktop-ui.md`](react-desktop-ui.md) — implement the React/TypeScript desktop interface with Base UI and atomic design.
- [`qa-release.md`](qa-release.md) — own validation strategy, cross-platform packaging checks, and release readiness.

## Delegation guidance

- Give each agent one concrete task, the exact files it may edit, relevant acceptance criteria from `implementation.md`, and the validation command(s) expected.
- Keep write scopes disjoint when agents run in parallel. The coordinator should integrate contracts before assigning frontend/backend work that depends on them.
- Ask agents to report files changed, tests run, outcomes, and unresolved risks. Review generated code and run project validation before accepting it.
- Use the project skills in `.agents/skills/` as specialized guidance when a task matches their descriptions.
