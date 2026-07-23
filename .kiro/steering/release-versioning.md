---
inclusion: always
---

# Release, Versioning, Commit & Documentation Workflow

Semantic versioning `MAJOR.MINOR.PATCH`. Sources of truth:

- **`VERSION`** (repo root) — the current version. Single authoritative value.
- **Annotated git tags** `vX.Y.Z` — shipped versions only.
- **`docs/CHANGELOG.md`** — per-change log (see change-log.md steering).
- **`docs/CHANGES-BY-FEATURE.md`** — per-feature history (see §5).
- Notion "17 · Deployment Documentation" mirrors the release table (sync on tag).

## 1. Current phase: pre-launch (IMPORTANT)

The app is **not deployed yet**. Until the first production deployment:

- `VERSION` stays **`1.0.0-dev`**. Do NOT bump it, do NOT create sub-versions (no v0.x), do NOT tag.
- All work is simply "pre-1.0"; record it in the changelog and changes-by-feature pages without version numbers.
- The **first production deployment becomes `v1.0.0`** (set `VERSION`, tag, record the deploy date). Only from that point on do the bump rules in §2 apply.

## 2. Versioning rules (from v1.0.0 onward)

A version increase is a **release decision, not a commit count**. One commit is NEVER automatically one version increase — many tickets and commits normally ship together under a single version.

On receiving a ticket:

1. **Gather the current version**: read `VERSION`, cross-check with `git tag --sort=-v:refname | head -1` and the latest entries in `docs/CHANGES-BY-FEATURE.md`. If they disagree, reconcile first (flag it in the ticket).
2. **Check whether the required change is big enough to increase the number**, and pick the increment:
   - **PATCH** — bug fixes, copy/styling, config, docs, refactors; no new capability.
   - **MINOR** — new feature, endpoint, page or integration; backward compatible.
   - **MAJOR** — breaking API/DB-schema change, incompatible behavior, portal-wide redesign.
   - **No bump** (the common case) — the change joins the release currently being prepared; its version was already decided.
3. If a bump is warranted, updating `VERSION` is part of the ticket's scope. If in doubt between two increments, pick the smaller and note the doubt in the ticket.

## 3. Branching

- **Ticket specifies a target release** (e.g. "Release: v1.1.0" in the Linear ticket or label) → work on branch `release/v1.1.0`. Create it from up-to-date `master` if missing. Commit inside the correct release branch — never mix tickets from different releases.
- **No release specified** → commit ahead of the latest changes on `master` (pull latest, commit on top).
- Release branches merge into `master` when the release ships, then get tagged and deleted.
- Note: this repo's default branch is `master` (not `main`); `production` is the deploy branch.

## 4. Once the change is implemented — closing checklist (in order)

1. Tests green + mandatory review pass (review-process.md).
2. Update `docs/CHANGELOG.md` (top entry, format from change-log.md, always include the Linear ID).
3. Update `docs/CHANGES-BY-FEATURE.md` (see §5).
4. Update `VERSION` only if §2 decided a bump (never pre-launch, see §1).
5. **Commit** with message: `<type>: <summary> (MA-XX)` — types: feat/fix/docs/refactor/test/chore. Append ` [vX.Y.Z]` only on the commit that bumps `VERSION`.
6. **Tag** only when a version actually ships to production: `git tag -a vX.Y.Z -m "<highlights>"` on the merge commit into `master`. Never tag work-in-progress, never tag pre-launch.
7. On tag: sync the release table in Notion "17 · Deployment Documentation" (or leave a `TODO(sync-notion)` line in the changelog entry).

## 5. Changes-by-feature page

`docs/CHANGES-BY-FEATURE.md` groups history **by feature**, not by date. Every ticket that adds or changes a feature MUST add a row under that feature's section (create the section if the feature is new):

```markdown
## <Feature name>

| Date | Ticket | Version | Change |
|------|--------|---------|--------|
| YYYY-MM-DD | MA-XX | pre-1.0 | One-line description of what changed |
```

Use `pre-1.0` in the Version column until launch; real versions afterwards. Rules: append only — never rewrite or delete past rows; one row per ticket per feature; if a ticket touches several features, one row in each affected section.
