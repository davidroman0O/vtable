# Checkbox List Example

This example shows how to add checkbox visual indicators to a styled list for clearer selection feedback.

## What it shows

- **Checkbox Indicators**: Visual "[ ]" for unselected and "[x]" for selected items
- **Same Styling**: All the colorful formatting from the styled list example
- **Clear Selection State**: Easy to see what's selected at a glance
- **All Features Work**: Navigation, selection, and status tracking work as before

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size
- `Home`/`End` or `g`/`G` - Jump to first/last item
- `Space` - Toggle selection (see checkbox change)
- `Ctrl+A` - Select all people (see all become [x])
- `Ctrl+D` - Clear all selections (see all become [ ])
- `q` or `Ctrl+C` - Quit

## What you'll see

Same colorful list as the styled example, but now with clear checkboxes:
- `[ ] Alice Johnson (28) 🌟 - UX Designer in San Francisco`
- `[x] Emma Wilson (52) 👑 - Tech Lead in Portland` (selected)
- `[ ] Bob Chen (34) - Software Engineer in New York`

## Key learnings

- How to add visual selection indicators to any formatter
- Building on existing formatters with minimal changes
- Using familiar UI patterns (checkboxes) for better usability
- The importance of visual feedback in list interfaces 