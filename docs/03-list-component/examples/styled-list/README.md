# Styled List Example

This example demonstrates beautiful visual styling with colors, highlighting, and professional terminal UI design using Lipgloss.

## What it shows

- **Rich Color Coding**: Different colors for names (white bold), ages (gold/orange/purple), jobs (job-specific colors), and cities (green)
- **Age-Based Styling**: Young professionals (🌟 gold), mid-career (orange), senior professionals (👑 purple)
- **Job-Specific Colors**: Engineers (turquoise), Managers (tomato), Designers (orchid), Leads (lime), etc.
- **Cursor Highlighting**: Gray background with white text for current item
- **Selection Highlighting**: Green background for selected items
- **Styled UI Elements**: Beautiful title, help text, and status messages
- **Professional Design**: Cohesive color scheme and visual hierarchy

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size
- `Home`/`End` or `g`/`G` - Jump to first/last item
- `Space` - Toggle selection of current item
- `Ctrl+A` - Select all people
- `Ctrl+D` - Clear all selections
- `q` or `Ctrl+C` - Quit

## Visual Features

### Color Coding by Data Type
- **Names**: White bold for emphasis
- **Ages**: Gold for young (< 30), orange for mid-career (30-45), purple for senior (> 45)
- **Jobs**: Dynamic colors based on role type (engineers, managers, designers, leads, etc.)
- **Cities**: Light green for locations

### State-Based Highlighting
- **Cursor**: Gray background highlighting for current position
- **Selection**: Green background for selected items
- **Status**: Gold bold text for selection count

### Age Categories with Emojis
- **Young professionals (< 30)**: `Alice Johnson (28) 🌟` in gold
- **Mid-career (30-45)**: `Bob Chen (34)` in orange  
- **Senior professionals (> 45)**: `Emma Wilson (52) 👑` in purple

### Job-Specific Colors
- **Engineers**: Turquoise (`#00CED1`)
- **Managers**: Tomato (`#FF6347`)
- **Designers**: Orchid (`#DA70D6`)
- **Leads/Principals**: Lime Green (`#32CD32`)
- **Developers**: Sky Blue (`#87CEEB`)
- **Architects**: Plum (`#DDA0DD`)

## Key learnings

- How to use Lipgloss for terminal styling
- Creating color-coded data displays
- Conditional styling based on data values
- Cursor and selection highlighting techniques
- Building professional-looking terminal UIs
- Color coordination and visual hierarchy
- Performance impact of styling (minimal with VTable) 