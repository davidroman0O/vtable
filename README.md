<p align="center">
  <img src="./demo.gif?raw=true" alt="VTable Demo" width="800">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/davidroman0O/vtable"><img src="https://pkg.go.dev/badge/github.com/davidroman0O/vtable.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/davidroman0O/vtable"><img src="https://goreportcard.com/badge/github.com/davidroman0O/vtable" alt="Go Report Card"></a>
  <a href="https://github.com/davidroman0O/vtable/blob/main/LICENSE"><img src="https://img.shields.io/github/license/davidroman0O/vtable" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.18-blue" alt="Go Version">
</p>

# VTable

A high-performance, feature-rich library of virtualized `List`, `Table`, and `Tree` components for [Bubble Tea](https://github.com/charmbracelet/bubbletea). It's designed to handle millions of items efficiently through intelligent virtualization and asynchronous, chunk-based data loading, while offering extensive customization options.

## ✨ Features

### 🚀 Core Engine
- **High Performance**: Built for speed, using viewport virtualization (`Chunking`) to handle massive datasets without breaking a sweat.
- **Asynchronous by Design**: All data operations (loading, sorting, filtering) are non-blocking and handled via `tea.Cmd`s, ensuring a perfectly responsive UI.
- **Pure Go & Bubble Tea Native**: Implemented as standard `tea.Model`s for seamless integration into any Bubble Tea application.
- **Stateful & Predictable**: Manages its own internal state, updated immutably through messages, making it easy to reason about.

### 📦 Components
- **`List`**: A powerful and customizable vertical list for homogenous items.
- **`Table`**: A multi-column table with headers, borders, advanced formatters, and multiple horizontal scrolling modes.
- **`TreeList`**: A hierarchical list for displaying tree-like data structures with node expansion and collapsing.

### 🎨 Rendering & Styling
- **Component-Based Rendering**: A highly flexible rendering pipeline. Build custom item/row layouts by assembling components like `Cursor`, `Enumerator`, `Content`, `Background`, and more.
- **Advanced Formatters**: Full control over item/cell rendering with simple (`ItemFormatter`) or animated (`ItemFormatterAnimated`) formatters.
- **Extensive Theming**: Easily configure `lipgloss` styles for every part of your component, from cursor and selection to borders and alternating rows.
- **Granular Border Control**: Independently control visibility of top, bottom, and header-separator borders in tables.
- **Advanced Horizontal Scrolling**: Sophisticated per-cell or global scrolling for tables with `character`, `word`, and `smart` modes.
- **Active Cell Indication**: Highlight the currently active cell in a table with a background color or custom formatter logic.

### 📊 Data Management
- **Asynchronous `DataSource`**: Your data source is completely decoupled from the UI. The components request data chunks as needed via commands.
- **Multi-Column Sorting**: Sort by multiple fields with priority.
- **Dynamic Filtering**: Apply complex filters to your data on the fly.
- **Selection Modes**: Supports `SelectionSingle`, `SelectionMultiple`, and `SelectionNone`.

## 📦 Installation

```bash
go get github.com/davidroman0O/vtable
```

## 🚀 Quick Start

`vtable` components are `tea.Model`s. You embed one in your own model and delegate `Update` calls to it. Interaction is done by sending `tea.Msg`s, which are created by `vtable`'s command functions (e.g., `vtable.CursorUpCmd()`).

For complete working examples, see the [Getting Started Guide](docs/01-getting-started/) and [Table Component Documentation](docs/05-table-component/).

## 📚 Core Concepts

For complete understanding of VTable's architecture, data management, and virtualization system, see the [Core Concepts documentation](docs/02-core-concepts/).

### Key Topics Covered:
- **DataSource Interface**: Asynchronous data loading and management
- **Viewport Virtualization**: High-performance rendering of large datasets  
- **Command & Message API**: Bubble Tea-native interaction patterns

## 🎨 Component-Based Rendering

VTable includes an advanced component-based rendering system for maximum flexibility. See the [Enhanced List Example](examples/enhanced-list/) for a complete demonstration of building custom layouts with render components.

## ⌨️ API Reference

Complete API documentation with all commands, messages, and configuration options is available in the source code:

- **Commands & Messages**: [`core/commands.go`](core/commands.go)
- **Configuration**: [`config/config.go`](config/config.go) and [`core/types.go`](core/types.go)
- **Components**: Individual component documentation in the component directories

## 📁 Examples & Documentation

### Progressive Documentation
Complete documentation with step-by-step examples:

1. [Getting Started](docs/01-getting-started/) - Your first VTable application
2. [Core Concepts](docs/02-core-concepts/) - Architecture and data management
3. [Table Component](docs/05-table-component/) - Multi-column tables with rich features

### Table Features (Progressive Examples)
1. [Basic Table](docs/05-table-component/01-basic-table.md) - Simple multi-column display
2. [Data Virtualization](docs/05-table-component/02-data-virtualization.md) - Handling large datasets
3. [Selection](docs/05-table-component/03-selection-table.md) - Single and multi-select
4. [Cell Constraints](docs/05-table-component/04-cell-constraints.md) - Layout and alignment control
5. [Column Formatting](docs/05-table-component/05-column-formatting.md) - Custom cell formatters and styling
6. [Table Styling](docs/05-table-component/06-table-styling.md) - Themes, borders, and visual customization

### Live Examples
The [`examples/`](examples/) directory contains comprehensive working examples:

- **`basic-list`**: Simple list with asynchronous data loading
- **`basic-table`**: Complete table features demonstration  
- **`basic-tree-list`**: Hierarchical data with TreeList component
- **`enhanced-list`**: Advanced component-based rendering system

```bash
# Run any example
cd examples/basic-table && go run .
```

## 📄 License

[MIT License](LICENSE)
