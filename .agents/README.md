# Yumdown Agent Resources

Use [`../implementation.md`](../implementation.md) as the source of product scope and acceptance criteria.

- [`agents/`](agents/README.md) contains reusable role prompts for coordinator, backend, frontend, and QA/release delegation. These Markdown prompts are not automatically registered as editor profiles; provide the selected prompt when delegating.
- [`skills/`](skills/README.md) contains project-local Zed skills for Wails/Go backend work, yt-dlp/media tools, React atomic UI, and cross-platform release validation.

Keep agent tasks bounded, specify file ownership and validation commands, and avoid overlapping write scopes when delegating in parallel.
