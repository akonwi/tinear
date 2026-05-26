# 0001: Record Architecture Decisions

## Status

Accepted

## Context

The project needs a lightweight way to document significant architectural
and design decisions — choices about structure, dependencies, conventions,
and UX patterns that future contributors (including AI sessions) should
understand without rediscovering them.

We want something more durable than chat history and more structured than
free-form notes. ADRs (Architecture Decision Records) are a well-known
pattern for this.

## Decision

We will use Architecture Decision Records stored in `docs/adrs/`.

Each ADR follows this template:

```markdown
# NNNN: Title

## Status

Proposed | Accepted | Deprecated | Superseded

(If Superseded, list the replacing ADR.)

## Context

What's the problem or opportunity? What constraints or trade-offs are at play?

## Decision

What we decided to do.

## Consequences

What changes as a result — positively or negatively. What needs to be
maintained, updated, or watched.

## Related

Optional links to other ADRs, PRs, issues, or external references.
```

- ADRs are numbered sequentially (`NNNN`).
- A short `docs/README.md` explains the system and how to add a new ADR.
- ADRs can be updated with new statuses as decisions evolve.
- Keep ADRs concise — a few paragraphs per section is enough.

## Consequences

- Decisions are documented in a discoverable, reviewable format.
- New contributors (human or AI) can catch up on past reasoning by reading
  the ADRs directory.
- Writing an ADR adds a small overhead to significant decisions, but avoids
  repeated re-argumentation.
- The `docs/adrs/` directory will grow over time; pruning or deprecating
  outdated ADRs keeps it useful.

## Related

- [docs/README.md](../README.md) — docs index and how-to.
- [ADR GitHub organization](https://adr.github.io/) — the wider ADR practice.
