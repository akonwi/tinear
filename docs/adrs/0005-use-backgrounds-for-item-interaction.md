# 0005: Use Backgrounds for Item Interaction States

## Status

Accepted

## Context

ADR 0004 established accent selection rails to avoid reverse-video rows while
preserving the host terminal palette. In practice, the rail is too subtle for
current selection and consumes space inside compact list, card, and picker
items. Pointer hover also lacks a distinct visual state.

Tinear needs a shared interaction treatment that makes hover and selection
obvious across retained controls without introducing application-owned RGB
colors.

## Decision

Replace selection rails in Inbox rows, board cards, and picker options with
full item background states:

- Hover uses host ANSI bright black/gray (`8`) with bright host text (`15`).
- Selected/active uses the host accent blue (`4`) with bright host text (`15`).
- Selected state wins over hover and also uses bold primary text so color is not
  the only signal.
- Inactive items explicitly return to the terminal-default foreground and
  background.
- Keyboard navigation clears stale hover state. Reconciliation clears hover
  when retained item records change.
- Backgrounds fill the content surface but not the spacing row between items.

The semantic colors live in `tui/theme.ard`. Controllers continue to own their
retained interaction state and mutate Cooper controls directly.

This decision supersedes ADR 0004 only where it prescribes accent selection
rails. The rest of Quiet Structure remains active.

## Consequences

- Hover is visible only when terminal mouse reporting is available; keyboard
  selection remains independently visible through background and weight.
- Host ANSI palettes vary, so exact contrast is terminal-dependent. Bright host
  text is used consistently on both interactive backgrounds.
- Semantic priority colors are suppressed inside interactive board cards to
  avoid conflicting foreground colors on filled backgrounds; glyph shape still
  communicates priority.
- Tests must cover inactive, hovered, and selected surfaces and verify that
  selected state takes precedence.

## Related

- [0004: Adopt the Quiet Structure Design Language](0004-quiet-structure-design-language.md)
- [Tinear design language](../design-language.md)
