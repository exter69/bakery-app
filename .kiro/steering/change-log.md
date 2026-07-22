---
inclusion: always
---

# Change Log Requirements

Every task that modifies code MUST end by updating `docs/CHANGELOG.md` (create it if missing).

## Entry format

Add a new entry at the TOP of the file:

```markdown
## [YYYY-MM-DD] <short title of the change>

- **Module/App**: <backend service, frontend app, or package touched>
- **Purpose**: <why the change was made, 1-2 sentences>
- **Features/Areas**: <feature names or domains affected>
- **Summary**: <recap of what was done — key files, new endpoints/components, config changes>
- **Tests**: <tests added or updated>
```

## Rules

- Summarize, don't enumerate. Recap intent and scope, not every line changed.
- One entry per prompt/task, even if multiple files changed.
- Group related changes under a single entry.
- Never delete or rewrite past entries; only prepend new ones.
- If the change is trivial (typo, comment), a one-line entry is enough.
