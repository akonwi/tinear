# 0004: Adopt the Quiet Structure Design Language

## Status

Accepted

## Context

The Cooper rewrite restored the shell, Inbox, and initial My Issues behavior
before investing in visual polish. The resulting screens are functional but
flatten structured information into mostly monochrome text and rely too heavily
on reverse video for selection. Meanwhile, the modal and toast surfaces use
strong dark fills and borders that feel disconnected from the otherwise
terminal-native screens.

Tinear needs a coherent visual baseline before more screens and mutations are
implemented. The design must work with Cooper's retained controls, preserve
comfortable information density, remain understandable without color, and
avoid introducing a widget abstraction.

## Decision

Adopt the **Quiet Structure** design language specified in
[`docs/design-language.md`](../design-language.md).

- Preserve terminal-default foreground and background throughout the app; the
  default theme contains no hard-coded RGB colors.
- Build hierarchy with spacing, alignment, styled text spans, and dim rules
  before adding borders or filled surfaces.
- Map semantic accents to the host terminal's ANSI palette through Cooper
  indexed colors. Use accent only for active navigation, focus, selection rails,
  and important interactive context.
- Replace full-row reverse-video selection with an accent `▌` rail and bold
  primary text.
- Use comfortable density: three-row Inbox records and approximately five-row
  board cards with wrapped titles and visible metadata.
- Render detail panes as retained structured controls rather than flat text
  blobs.
- Reserve rounded borders for modal separation and make toasts visually light.
- Keep semantic styles in the theme and pure screen-local style/content
  functions; continue constructing and mutating Cooper controls directly.

## Consequences

- Shell, Inbox, board, picker, modal, and toast controls need a visual-baseline
  pass before feature work continues.
- Existing feature documentation that prescribes reverse video is superseded by
  this decision and should be updated as those screens are restyled.
- Comfortable spacing displays fewer records than the compact parity UI, so
  narrow and short terminal coverage is required.
- Terminal defaults and indexed semantic colors integrate with light and dark
  host themes. Cooper still cannot query luminance or derive a complete
  contrast-adjusted semantic palette, so default foreground plus attributes
  must carry most hierarchy.
- Selection remains accessible without color through glyph and weight changes.

## Related

- [0003: Adopt Cooper Retained UI](0003-adopt-cooper-retained-ui.md)
- [Tinear design language](../design-language.md)
- [Cooper rewrite parity roadmap](../cooper-rewrite-roadmap.md)
