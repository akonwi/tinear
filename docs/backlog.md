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
- [x] **Comment threading** (read-only): replies render indented with
  `↳` under their parents. Posting replies still TODO (compose would
  need a reply-to target).
- [x] **Edit issue description** (`e` on an issue tab): prefilled
  modal, Ctrl+Enter saves via `issueUpdate`.
- [ ] **Edit your own comments** — blocked on a comment cursor: the
  comments section has no per-comment selection yet. Building one also
  unlocks posting threaded replies (reply-to target).

## Discovery & navigation

- [ ] **Filtered board views**. My Issues is fixed to "assigned to me";
  add an active-cycle filter and/or project switcher. Persist the
  filter in the existing cache shape.
- [x] **Open in browser** (`o`): inbox items, issue tabs, doc tabs
  (board deliberately unbound — Enter opens the tab there).
- [ ] ~~Recent documents list~~ dropped — search covers doc discovery.

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
