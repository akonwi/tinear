# Tinear design language

Tinear uses a **Quiet Structure** visual language: terminal-native, calm, and
information-dense without feeling cramped. It borrows Linear's restrained
hierarchy rather than reproducing a web interface inside the terminal.

## Principles

1. **Respect the terminal.** Keep the user's default foreground and background
   across primary screens. Do not paint a full application canvas.
2. **Structure before chrome.** Create hierarchy with alignment, spacing,
   styled text, and dim rules. Borders and filled surfaces are reserved for
   overlays that need separation.
3. **One accent, used deliberately.** Accent marks the active tab, current
   selection rail, focus, and primary interactive context. It is not decoration.
4. **Selection is not inversion.** A selected row or card uses an accent `▌`
   rail plus bold primary text. Do not reverse the whole item.
5. **Secondary information recedes.** Metadata uses the terminal's default
   color with `dim` where possible. Semantic colors communicate priority,
   warning, success, and failure and never carry meaning alone.
6. **Comfortable density.** Preserve breathing room between records, while
   keeping the most important context above the fold.
7. **Behavior and visuals share state.** Focus, loading, errors, stale data,
   and disabled actions must have explicit visual treatments.

## Semantic roles

`Theme` owns semantic roles, not component-specific colors. It is a host-theme
adapter, not an application color scheme:

| Role | Use |
|---|---|
| terminal default | canvas, surfaces, primary text, normal content |
| default + `dim` | metadata, inactive tabs, separators, timestamps |
| host ANSI blue | active tab, selection rail, focus, notification context |
| host ANSI red | destructive actions, failures, urgent priority |
| host ANSI yellow | warnings, high priority |
| host ANSI green | successful completion and confirmations |
| host ANSI bright black/gray | modal boundaries and rare structural separators |

The default theme must not contain hard-coded RGB values. Missing foreground
and background values preserve the terminal defaults; semantic colors use
Cooper's indexed-color support so ANSI slots resolve through the user's terminal
palette. Prefer `dim` over a color for muted content. Every colored state also
uses a glyph, label, weight, or placement so color is never the only signal.

## Shell

- The tab bar is one line followed by a dim horizontal rule.
- The active tab is accent-colored, bold, and underlined. Inactive tabs are
  dim. Tabs do not use reverse-video chips.
- Permanent tabs remain visually stable; dynamic tabs add a dim `×` suffix.
- The footer begins with a dim rule. Its left side provides context/status and
  its right side provides current key hints. Keys are accent/bold; labels are
  dim.
- Primary body regions use terminal-default background without boxes.

## Selection and focus

- Lists, cards, and picker options reserve a one-cell selection rail.
- Current focused selection: accent `▌` plus bold primary text.
- Remembered selection in an unfocused region: dim `▌`, without bold.
- Hover does not exist as a separate state. A click moves the cursor and invokes
  the control's documented action.
- Focus must remain understandable without color through the rail and weight.

## Inbox

- Use a comfortable three-row rhythm per notification: title, secondary line,
  then separation.
- Primary line: issue title, one line, ellipsized.
- Secondary line: actor + action, followed by identifier/team context, dim.
- The list/detail split uses a dim vertical separator, not bordered panels.
- Detail content is structured rather than rendered as one text blob:
  accent/bold notification context, dim identifier/team, wrapped bold title,
  aligned metadata labels, and labeled dim section rules.
- Loading, empty, and error states occupy the content region without adding a
  new panel style.

## My Issues

- Columns are approximately 42 cells wide with two cells between columns.
- Column headers use bold state name, dim count, and a dim rule. The selected
  column can strengthen the header but does not gain a box border.
- Cards use a comfortable five-row rhythm: priority + bold identifier, title up
  to two lines, dim cycle/project metadata, and separation.
- Priority combines glyph and semantic color:
  urgent = danger `!`, high = warning `↑`, medium = default `~`, low = dim `↓`.
- Selected cards use the shared accent rail and bold identifier treatment.
- Horizontal and vertical scrolling must keep the full selected card visible.

## Pickers and search

- Reuse the same selection rail and typography as screen lists.
- The filter input appears first, followed by a dim rule and the result list.
- Option label is primary; description is dim and may occupy a second line.
- Empty and loading states are inline. Do not add nested borders.

## Overlays

- Modals use a rounded border and title because they must separate from the
  terminal-default canvas. Their body keeps the terminal-default background;
  the active modal border uses accent, while destructive confirmation uses
  danger.
- Toasts are lightweight, short-lived notices. Prefer a semantic glyph, bold
  title, and dim detail over a heavily filled card.
- Overlay styling must not introduce a second design language into the app.

## Implementation constraints

- Continue using Cooper controls directly; this guide does not authorize a
  widget or binding abstraction.
- Put reusable values in `tui/theme.ard` and keep screen-specific style/content
  functions pure where practical.
- Use styled text spans rather than extra controls when only typography or
  color differs.
- Retained controls are constructed once and restyled through setters.
- Test visual state changes at narrow, standard, and wide terminal sizes.
- Cooper's indexed colors preserve host ANSI palette choices without querying
  or owning the terminal canvas. Default foreground plus attributes should do
  most hierarchy work; indexed semantic colors are reserved for meaning.
- Cooper does not yet query foreground/background luminance or derive
  contrast-adjusted roles. Do not compensate with a dark-only RGB palette.
