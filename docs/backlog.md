# Backlog

Feature ideas, roughly ordered by value-to-effort within each group.
Not commitments — a parking lot for what to build next.

## High value, fits existing patterns

- [x] **Issue creation** (`c` from the board). The biggest workflow gap:
  tinear can triage, read, and comment, but not capture a new bug/idea.
  Compose-style modal (title + markdown body, team/project pickers);
  reuses `compose_view` and `picker` machinery. `issueCreate` mutation.
- [x] **Edit more issue fields**: priority (`p`), assignee (`a`),
  project (`P`).
- [ ] **Labels** — deliberately deferred: tinear doesn't display labels
  anywhere yet, so display (detail header, maybe board cards) would
  have to come first, and editing needs a multi-select picker widget.
  Low priority until labels prove useful to see.
- [ ] **Comment threading**. Linear comments have `parent`/`children`;
  indent replies in `comment_card`. Reading value even before posting
  replies is supported.
- [ ] **Edit your own comments / issue description** (`e`). Compose modal
  prefilled with the existing text; `commentUpdate` / `issueUpdate`.

## Discovery & navigation

- [ ] **Filtered board views**. My Issues is fixed to "assigned to me";
  add an active-cycle filter and/or project switcher. Persist the
  filter in the existing cache shape.
- [ ] **Recent documents list**. `documents::fetch_recent` already exists
  and is unused — a docs section or tab listing recently updated docs
  gives discovery without searching.
- [ ] **Open in browser everywhere** (`o`). Doc tabs have it; issue tabs
  and inbox items should too (issue `url` is already fetched).

## Polish that compounds

- [ ] **Syntax highlighting in code blocks**. The deferred chroma
  follow-up; slots into `tui/markdown.ard`'s `code_widget` branch.
  Weigh binary size (chroma embeds ~250 lexers).
- [ ] **Inbox: mark-all-read and snooze**
  (`notificationUpdate.snoozedUntilAt`) — delete exists; snooze is the
  other half of triage.
- [ ] **Global manual refresh** (`R`): refresh every mounted view now,
  alongside the 5-minute background refresh.
- [ ] **Issue relations display**: blocked-by / blocks / duplicates on
  the detail header. Read-only via `issue.relations` is easy.

## Bigger bets

- [ ] **Create/edit documents**. Possible with a TextArea, but markdown
  editing in a terminal field is rough UX — needs design thought.
- [ ] **Offline-ish cache**: persist last-fetched board/inbox so startup
  paints instantly before the first refresh. Extends `models/cache`,
  plays to the existing hydrate pattern.

## Suggested next

**Issue creation**, then **priority/assignee pickers** — together they
make tinear a complete daily driver rather than a viewer.
