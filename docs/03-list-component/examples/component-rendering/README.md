# VTable Component Rendering Demo

This example demonstrates VTable's component-based rendering system. Learn how each list item is built from individual components that you can rearrange and style.

## What it shows

- **Component building blocks**: How cursor, enumerator, and content combine to make list items
- **Component reordering**: Moving pieces around to create different layouts
- **Component removal**: Creating minimal layouts with fewer components
- **Spacing components**: Adding space before and after content
- **Component styling**: Individual background colors and styling for each component
- **Dynamic controls**: Adjusting spacing width and toggling backgrounds

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down
- `Space` - Toggle selection
- **`c`** - **Cycle through component layouts**
- **`b`** - **Toggle component backgrounds on/off**
- **`+`/`-`** - **Increase/decrease spacing width**
- `q` - Quit

## Layout Options

### 1. Default: [Cursor][Enumerator][Content]
```
►  1. Alice Johnson (28) - UX Designer in San Francisco
   2. Bob Chen (34) - Software Engineer in New York
```
The standard VTable layout with cursor, numbers, and content.

### 2. Numbers at End: [Cursor][Content][Enumerator]
```
► Alice Johnson (28) - UX Designer in San Francisco  1.
  Bob Chen (34) - Software Engineer in New York      2.
```
Same components, but numbers moved to the end of each line.

### 3. Content Only: [Content]
```
Alice Johnson (28) - UX Designer in San Francisco
Bob Chen (34) - Software Engineer in New York
```
Minimal layout with just the formatted content - no cursor or numbers.

### 4. With Spacing: [Spacing][Cursor][Enumerator][Content][Spacing]
```
  ►  1. Alice Johnson (28) - UX Designer in San Francisco 
     2. Bob Chen (34) - Software Engineer in New York    
```
Adds spacing before and after content. Use +/- keys to adjust spacing width (0-10 spaces).

## Additional Features

### Background Styling (Press 'b' to toggle)
When enabled, each component gets its own background color:
- **Cursor**: Gray background (darker when selected)
- **Enumerator**: Green background with white text
- **Content**: Dark background (different color when selected)

This works with any layout to help you visualize the individual components.

### Dynamic Spacing Width (Press '+'/'-')
In layout 4 (With Spacing), adjust the spacing width from 0 to 10 spaces:
- Current width shown in status line
- Affects both PreSpacing and PostSpacing components
- See real-time changes as you adjust

## Key Learning Points

**Components**: Each list item = individual pieces you can rearrange

**Component Order**: Control which pieces appear and in what sequence

**Spacing**: Add space before and after with PreSpacing/PostSpacing components

**Individual Styling**: Each component can have its own colors and backgrounds

**Dynamic Control**: Adjust spacing and styling without restarting

## What You Can Do

- Press 'c' to cycle through all 4 layouts
- Press 'b' to toggle backgrounds on/off in any layout
- In layout 4, use +/- to adjust spacing width and see the changes
- Select items (Space) to see different background colors for selected items
- Navigate around to see how cursor positioning works in each layout
- Compare how the same data looks completely different with different component arrangements 