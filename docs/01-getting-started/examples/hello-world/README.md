# Hello World VTable Example

This is the simplest possible VTable example. It creates a basic list with 10 items that you can navigate with arrow keys or j/k.

## What You'll See

```
Hello World VTable List (press 'q' to quit)

► Item 1
  Item 2
  Item 3
  Item 4
  Item 5

Use ↑/↓ or j/k to navigate
```

## How to Run

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down
- `q` or `Ctrl+C` - Quit

## What This Demonstrates

- Basic DataSource implementation
- Simple List creation  
- Bubble Tea integration
- Virtual rendering (only 5 visible items)
- Keyboard navigation with proper command handling

This is your starting point for all VTable components! 