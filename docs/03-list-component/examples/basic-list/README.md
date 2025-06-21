# Basic List Example

This example demonstrates the simplest possible VTable list with just navigation.

## What it shows

- **DataSource**: Simple string data provider
- **Navigation**: Arrow keys and j/k movement
- **Viewport**: 5 items visible at a time
- **Data virtualization**: Efficient rendering of 20 items

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down
- `q` or `Ctrl+C` - Quit

## Key learnings

- How to implement a basic DataSource
- List configuration and creation
- Mapping keys to VTable commands
- The foundation for all other list features 