# VTable Enumerator System Demo

This example demonstrates VTable's powerful enumerator system with multiple approaches: convenience methods, direct configuration access, and custom enumerator functions.

## What it shows

- **Component-Based Rendering**: How enumerators fit into VTable's rendering pipeline
- **Convenience Methods**: Quick setup with `SetNumberedStyle()`, `SetBulletStyle()`, etc.
- **Direct Configuration**: Accessing `RenderConfig.EnumeratorConfig` for full control
- **Custom Functions**: Writing your own enumerator functions with data access
- **Interactive Demo**: Press 'e' to cycle through 6 different enumerator styles

## Formatter Setup

**Note**: For enumerators to work, set your formatter in the config:

```go
// Bypasses component system
vtableList := list.NewList(listConfig, dataSource, styledPersonFormatter)

// Recommended - Formatter in config (clean and explicit)
listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter
vtableList := list.NewList(listConfig, dataSource)

// Alternative - Set through render config after creation  
vtableList := list.NewList(listConfig, dataSource)
renderConfig := vtableList.GetRenderConfig()
renderConfig.ContentConfig.Formatter = styledPersonFormatter
vtableList.SetRenderConfig(renderConfig)
```

When you pass a formatter to `NewList()`, VTable uses it directly and bypasses the component-based rendering system that includes enumerators.

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size
- `Home`/`End` or `g`/`G` - Jump to first/last item
- `Space` - Toggle selection
- `Ctrl+A` - Select all people
- `Ctrl+D` - Clear all selections
- **`e`** - **Cycle through enumerator styles**
- `q` or `Ctrl+C` - Quit

## Enumerator Styles Demonstrated

### 1. Arabic Numbers (Convenience Method)
```go
vtableList.SetNumberedStyle()
```
**Output:** `1. Alice Johnson...`, `2. Bob Chen...`

### 2. Bullet Points (Convenience Method)
```go
vtableList.SetBulletStyle()
```
**Output:** `• Alice Johnson...`, `• Bob Chen...`

### 3. Checkboxes (Convenience Method)
```go
vtableList.SetChecklistStyle()
```
**Output:** `☐ Alice Johnson...`, `☑ Bob Chen...` (based on selection)

### 4. Custom Brackets (Direct Config)
```go
renderConfig := vtableList.GetRenderConfig()
renderConfig.EnumeratorConfig.Enumerator = customBracketEnumerator
renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
vtableList.SetRenderConfig(renderConfig)
```
**Output:** `[1] Alice Johnson...`, `[2] Bob Chen...`

### 5. Smart Conditional (Custom Function)
```go
func smartEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
    if item.Selected {
        return "✓ "  // Checkmark for selected
    }
    return fmt.Sprintf("%d. ", index+1) // Numbers for unselected
}
```
**Output:** `✓ Alice Johnson...` (selected), `2. Bob Chen...` (unselected)

### 6. Job-Aware Emojis (Data-Aware Custom)
```go
func jobAwareEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
    person := item.Item.(Person)
    if strings.Contains(person.Job, "Manager") {
        return "👑 "
    } else if strings.Contains(person.Job, "Engineer") {
        return "⚙️ "
    } else if strings.Contains(person.Job, "Designer") {
        return "🎨 "
    }
    return fmt.Sprintf("%d. ", index+1)
}
```
**Output:** `👑 Carol Rodriguez...` (Manager), `⚙️ Bob Chen...` (Engineer), `🎨 Alice Johnson...` (Designer)

## Key Technical Concepts

### Enumerator Function Signature
```go
type ListEnumerator func(item core.Data[any], index int, ctx core.RenderContext) string
```

### Direct Configuration Access
```go
renderConfig := vtableList.GetRenderConfig()
renderConfig.EnumeratorConfig.Enumerator = myCustomFunction
renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
renderConfig.EnumeratorConfig.MaxWidth = 5
renderConfig.EnumeratorConfig.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
vtableList.SetRenderConfig(renderConfig)
```

### Component-Based Pipeline
1. **Cursor Component** - Shows current position
2. **Enumerator Component** - Shows your enumeration
3. **Content Component** - Shows your formatted data

## Key learnings

- How VTable's component-based rendering pipeline works
- Difference between convenience methods and direct configuration
- Writing custom enumerator functions with access to item data
- Using alignment and styling for professional appearance
- Creating conditional enumerators that adapt to item state
- Leveraging the separation between enumeration and content formatting 