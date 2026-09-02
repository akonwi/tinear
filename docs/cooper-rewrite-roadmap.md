# Cooper Rewrite Roadmap

> Behavioral parity reference: `main` at `84ff69a`.
>
> Cooper rewrite branch: `rewrite/cooper`.

The retained Cooper rewrite is complete and ready for review/release preparation.
Completed behavior is documented in [features.md](features.md), architecture in
[ADR 0003](adrs/0003-adopt-cooper-retained-ui.md), and visual decisions in
[design-language.md](design-language.md).

## Release status

- Cooper is pinned to remote revision `848bac2`.
- Native Linux/macOS amd64/arm64 CI and release workflows are restored.
- Latest validation: `ard check main.ard`, `ard build main.ard`, 136 Ard tests,
  Go Markdown parser tests, and the live PTY suite pass.
- The PTY suite covers startup, terminal/cache restoration, resize, normalized
  keys, mouse/wheel input, multiline paste, OSC 52 copy, and clean quit.

## Explicitly out of parity scope

These are future product ideas, not Cooper rewrite requirements:

- [ ] Projects as a permanent tab
- [ ] `:` command palette
- [ ] Separate `login` CLI command / old `commands/*` architecture
- [ ] Full Inbox/board/offline cache from `data-cache-plan.md`
- [ ] Recent-documents list
- [ ] Labels display/editing
- [ ] Editing comments
- [ ] Filtered boards
- [ ] Syntax highlighting
- [ ] Mark-all-read and snooze
- [ ] Global `R` refresh
- [ ] Issue relations
- [ ] Document creation/editing
