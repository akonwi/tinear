# tinear docs

Architecture Decision Records (ADRs) and project documentation.

## Reference

- [features.md](features.md) — feature reference for the TUI: startup, the
  logged-in shell, inbox/board/issue-detail flows, modals, persistence.
- [design-language.md](design-language.md) — the Quiet Structure visual system
  for layout, hierarchy, selection, color, density, and overlays.
- [cooper-rewrite-roadmap.md](cooper-rewrite-roadmap.md) — parity ledger and
  release-hardening checklist for the retained Cooper implementation.

## ADRs

[Architecture Decision Records](adrs/) capture important decisions about
the project's architecture, tooling, conventions, and design. Each ADR is a
short, lightweight document explaining the context, the decision, and its
consequences.

### How to add an ADR

```bash
# Copy the template
cp docs/adrs/0001-record-architecture-decisions.md docs/adrs/NNNN-title-of-decision.md

# Edit, filling in Status, Context, Decision, Consequences
# Use the next sequential number
```

See [0001: Record Architecture Decisions](adrs/0001-record-architecture-decisions.md)
for the full template and conventions.
