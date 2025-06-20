# Navigation and Keys: Enhanced Movement

Let's enhance our basic list by adding more navigation options. Same list, better movement!

## What We're Adding

Taking our "Item 1, Item 2, Item 3..." list and adding:
- **Page navigation**: Page Up/Down to jump by viewport size
- **Jump navigation**: Home/End to go to first/last item  
- **More key options**: Additional key bindings for better UX

## Key Changes

We only need to modify the key handling in our `Update()` method:

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
			
		// Basic movement (same as before)
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
			
		// NEW: Page navigation
		case "pgup", "h":
			return app, core.PageUpCmd()
		case "pgdown", "l":
			return app, core.PageDownCmd()
			
		// NEW: Jump navigation  
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()
		}
	}

	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}
```

**That's it!** Just add those extra key cases.

## Navigation Commands

### Page Movement  
```go
core.PageUpCmd()      // Move up by viewport height
core.PageDownCmd()    // Move down by viewport height
```
Perfect for quickly scanning through larger lists.

### Jump Movement
```go
core.JumpToStartCmd() // Go to first item
core.JumpToEndCmd()   // Go to last item
```
Instant navigation to list boundaries.

## Key Binding Choices

| Action | Primary | Alternative | Why |
|--------|---------|-------------|-----|
| Up/Down | `↑`/`↓` | `k`/`j` | Arrow keys + vi-style |
| Page | `PgUp`/`PgDn` | `h`/`l` | Standard + vi left/right |
| Jump | `Home`/`End` | `g`/`G` | Standard + vi "go to" |

## What You'll Experience

With 50 items and viewport height 8:

1. **Press `l` (Page Down)**: Jump 8 items forward
2. **Press `G` (End)**: Jump to the last item instantly
3. **Press `g` (Home)**: Jump back to first item
4. **Press `h` (Page Up)**: Jump 8 items backward

The viewport smoothly follows your navigation, loading data as needed.

## Complete Example

See the enhanced navigation example: [`examples/enhanced-navigation/`](examples/enhanced-navigation/)

Run it:
```bash
cd docs/03-list-component/examples/enhanced-navigation
go run main.go
```

## Try It Yourself

1. **Test page navigation**: Use `l` and `h` to jump through a larger list quickly
2. **Try vi-style keys**: Use `j`, `k`, `G`, `g` if you're a vi user  
3. **Expand the dataset**: Change to 1000 items and see how page navigation helps

## What's Next

Our list now has professional-grade navigation! Next, we'll add the ability to select items with the spacebar.

**Next:** [Basic Selection →](03-basic-selection.md) 