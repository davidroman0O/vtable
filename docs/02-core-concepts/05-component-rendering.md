# Core Concepts: Component Rendering

VTable components handle their own visual output using **Lipgloss** for styling and a flexible **component-based system** for layout control. This section explains how VTable renders its UI and how you can customize it.

## The `View()` Method

Every VTable component (List, Table, Tree) has a `View()` method that returns a single, fully-styled string ready for terminal display.

```go
// In your Bubble Tea app's View method:
func (m MyApp) View() string {
    // You don't need to loop or style items manually.
    // The VTable component handles all its own rendering.
    return m.list.View()
}
```

You never need to manually render rows or cells. You configure the styles and layout, and VTable does the rest.

## Styling with Lipgloss

VTable uses [Lipgloss](https://github.com/charmbracelet/lipgloss) from Charm for all styling. You define `lipgloss.Style` objects and provide them in the component's configuration. VTable then applies these styles automatically based on the item's state (e.g., cursor, selected, normal).

```go
import "github.com/charmbracelet/lipgloss"

// Define a style for the cursor.
cursorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF6B35")). // Orange text
    Background(lipgloss.Color("#1A1A1A")). // Dark background
    Bold(true)

// Assign it in the configuration.
listConfig.StyleConfig.CursorStyle = cursorStyle
```
VTable will now automatically apply this style to the item currently under the cursor.

## Default Themes and Styles

To get you started quickly, VTable provides sensible defaults in the `config` package.

#### For Lists and Trees: `StyleConfig`
A `core.StyleConfig` defines styles for different item states.

```go
import "github.com/davidroman0O/vtable/config"

// Get the default styles.
styleConfig := config.DefaultStyleConfig()

// styleConfig now contains:
// - CursorStyle
// - SelectedStyle
// - DefaultStyle
// - LoadingStyle
// - ErrorStyle
// ...and more.
```

#### For Tables: `Theme`
A `core.Theme` defines a more comprehensive set of styles for tables, including borders.

```go
import "github.com/davidroman0O/vtable/config"

theme := config.DefaultTheme()

// theme now contains:
// - HeaderStyle
// - CellStyle
// - BorderChars (for drawing lines)
// - BorderColor
// ...and more.
```
You can use these defaults as a starting point and customize only what you need.

## Component-Based Rendering

For maximum layout flexibility, VTable uses a **component-based rendering pipeline**. Each part of a rendered row (like the cursor indicator or the item content) is a separate, configurable component.

#### List and Tree Components
A list or tree item is composed of several pieces that are rendered in a specific order:
`[Cursor] [Enumerator] [Content]`

-   **Cursor**: The `►` symbol or spacing that indicates the current line.
-   **Enumerator**: A prefix like a number (`1.`), bullet (`•`), or checkbox (`[x]`).
-   **Content**: Your actual formatted item data.

You can change the order of these components, customize their appearance, or even disable them entirely through the `RenderConfig` for lists and trees. This allows you to create a wide variety of layouts, from simple text lists to complex, numbered checklists.

#### Table Components
Tables use a similar system but are structured around columns. You can control the styling of headers, cells, and borders independently to create clean, readable tabular data displays.

## Color Support

Lipgloss (and therefore VTable) automatically adapts to the user's terminal capabilities. You can use any color format, and it will gracefully degrade on less capable terminals.

-   **ANSI (16 colors):** `lipgloss.Color("1")` // Red
-   **256 Colors:** `lipgloss.Color("196")` // Bright Red
-   **True Color (Hex):** `lipgloss.Color("#FF5733")`
-   **Adaptive Colors:** Define different colors for light and dark terminal themes.

```go
lipgloss.AdaptiveColor{
    Light: "#333333", // Dark text on light background
    Dark:  "#FFFFFF", // Light text on dark background
}
```

## What's Next?

You now have a solid understanding of VTable's core concepts! You know how it handles data, navigation, and rendering. You're ready to start building with the components themselves.

**Next:** [The List Component →](../03-list-component/01-basic-usage.md) 