# Enhanced Navigation Example

This example builds on the basic list by adding page and jump navigation options.

## What it shows

- **Page navigation**: PgUp/PgDn and h/l for jumping by viewport size
- **Jump navigation**: Home/End and g/G for instant navigation to edges
- **Larger dataset**: 50 items to better appreciate the new navigation
- **Taller viewport**: 8 items visible to demonstrate page jumps

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size (8 items)
- `Home`/`End` or `g`/`G` - Jump to first/last item instantly
- `q` or `Ctrl+C` - Quit

## Key learnings

- How to add more navigation commands to your key handler
- The difference between line, page, and jump movement
- Why page navigation becomes essential with larger datasets
- Vi-style key bindings alongside standard keys 