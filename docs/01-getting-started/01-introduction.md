# Introduction to VTable

## What is VTable?

VTable is a high-performance Go library for building terminal applications that can handle massive amounts of data. Built for the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, it provides powerful, easy-to-use **virtualized lists, tables, and trees**.

"Virtualized" means that VTable is incredibly efficient. It only renders what's visible on screen and intelligently loads data in chunks as you scroll. This allows you to build TUIs that can handle hundreds, thousands, or even millions of items without breaking a sweat.

## The Core Benefits: Why Use VTable?

#### 🚀 Handle Massive Datasets with Ease
Display lists with millions of items without performance degradation or memory issues. VTable only keeps a small portion of your data in memory at any time.

- **Your dataset:** 1,000,000 items
- **VTable loads:** ~100 items at a time
- **Memory usage:** Constant, regardless of dataset size

#### ⚡ Instant Responsiveness
Navigate through any size dataset with consistent, snappy performance. Scrolling, selection, and filtering remain fast whether you have 100 or 100,000 items.

#### 🎨 Highly Customizable
- **Lists**: Create bulleted, numbered, or checkbox-style lists with custom formatting.
- **Trees**: Build expandable hierarchies with custom symbols and indentation.
- **Tables**: Design tables with custom column formatting, sorting, borders, and horizontal scrolling.
- **Theming**: Take complete visual control with custom colors, styles, and layouts.

#### 🔧 Developer-Friendly API
- **Simple interfaces** that follow familiar Bubble Tea patterns.
- **Progressive complexity**—start simple and add advanced features as you need them.
- **Complete type-safety** with Go generics.

## What You'll Build

By the end of this guide, you'll know how to create professional, data-intensive components:

**Lists:** From simple navigation to complex, styled checklists.
```
► Item 1                   ► [x] Buy groceries
  Item 2                     [ ] Walk the dog
  Item 3                     [x] Read documentation
```

**Trees:** For hierarchical data like file systems or nested categories.
```
▼ Project                    ► Project (3 files)
  ├── 📄 main.go              ► docs (2 files)
  └── 📁 internal             ► tests (5 files)
```

**Tables:** For structured, columnar data with sorting and custom rendering.
```
┌───────────┬─────┬─────────────┐
│ Name ↑    │ Age │ City        │
├───────────┼─────┼─────────────┤
│ Alice     │ 28  │ New York    │
│ Bob       │ 34  │ Los Angeles │
└───────────┴─────┴─────────────┘
```

## How It Works: The Core Concepts

VTable uses a **data virtualization** architecture to stay fast and efficient.

1.  **DataSource**: Your data provider. It can be an in-memory slice, a database connection, or an API. It provides data in small, manageable chunks on demand.
2.  **Viewport**: A "moving window" that represents the visible portion of your data on screen. When you scroll, the viewport moves, not the data itself.
3.  **Component**: The UI component (List, Table, or Tree) that requests data from the viewport and renders it to the terminal.

This architecture ensures that only the data needed for the current view is ever processed, making VTable incredibly scalable.

## Quick Installation

**Prerequisites:** Go 1.19+ and basic familiarity with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```bash
go get github.com/davidroman0O/vtable
```

You'll also need Bubble Tea:
```bash
go get github.com/charmbracelet/bubbletea
```

That's it! You're ready to build your first VTable component.

## What's Next?

Ready to get your hands dirty? Let's build your first virtualized list in under 5 minutes.

**Next:** [Quick Start: Your First List →](02-quick-start.md) 