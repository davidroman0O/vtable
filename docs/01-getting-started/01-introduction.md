# Introduction to VTable

## What is VTable?

VTable is a high-performance library for building **virtual tables, lists, and trees** in terminal user interfaces. Built for [Bubble Tea](https://github.com/charmbracelet/bubbletea), VTable handles large datasets efficiently by only rendering what's visible on screen and loading data in chunks as needed.

## Why Use VTable?

### 🚀 **Handle Massive Datasets**
Display lists with millions of items without performance degradation or memory issues. VTable only keeps a small portion of data in memory at any time.

```
Your dataset: 1,000,000 items
VTable loads: ~100 items at a time
Memory usage: Constant, regardless of dataset size
```

### ⚡ **Instant Responsiveness**
Navigate through any size dataset with consistent performance. Scrolling, selection, and filtering remain snappy whether you have 100 or 100,000 items.

### 🎨 **Highly Customizable**
- **Lists**: Bullets, numbers, checkboxes, custom formatting
- **Trees**: Expandable hierarchies with custom symbols and indentation  
- **Tables**: Column formatting, sorting, borders, horizontal scrolling
- **Theming**: Complete visual control with colors, styles, and layouts

### 🔧 **Developer Friendly**
- Simple interfaces that follow Bubble Tea patterns
- Progressive complexity - start simple, add features as needed
- Extensive customization without complexity
- Complete TypeScript-like type safety with Go generics

## When Should You Use VTable?

### ✅ **Perfect For:**
- **Large datasets**: Log viewers, database browsers, file managers
- **Interactive data**: Selection, filtering, sorting capabilities needed
- **Professional TUIs**: Admin dashboards, monitoring tools, developer utilities
- **Performance-critical apps**: When responsiveness matters more than simplicity

### ❌ **Consider Alternatives When:**
- **Small, static lists** (< 100 items that never change)
- **Simple display only** (no interaction needed)
- **Learning Bubble Tea** (start with basic components first)

## What You'll Build

By the end of this guide, you'll know how to create:

**Lists:**
```
► Item 1
  Item 2          →    [x] Buy groceries     →    1. First task
  Item 3               [ ] Walk the dog           2. Second task  
  Item 4               [x] Read documentation     3. Third task
```

**Trees:**
```
📁 Project
├── 📁 src                →    ▼ Project               →    📁 Project
│   ├── 📄 main.go             ├── ▼ src                    ├── ▶ src (3 files)
│   └── 📄 utils.go            │   ├── 📄 main.go           ├── ▶ docs (2 files)
└── 📁 docs                    │   └── 📄 utils.go         └── ▶ tests (5 files)
    └── 📄 README.md           └── ▶ docs
```

**Tables:**
```
│ Name     │ Age │ City         │    →    ┌─────────┬─────┬─────────────┐
│ Alice    │ 28  │ New York     │         │ Name ↑  │ Age │ City        │
│ Bob      │ 34  │ Los Angeles  │         ├─────────┼─────┼─────────────┤
│ Charlie  │ 31  │ Chicago      │         │ Alice   │ 28  │ New York    │
                                          │ Bob     │ 34  │ Los Angeles │
                                          │ Charlie │ 31  │ Chicago     │
                                          └─────────┴─────┴─────────────┘
```

## Architecture Overview

VTable uses a **data virtualization** approach:

1. **DataSource**: Provides data chunks on demand
2. **Viewport**: Manages what's currently visible  
3. **Components**: Handle rendering, selection, styling
4. **Messages**: Control behavior via Bubble Tea commands

```
[Your Data] → [DataSource] → [Viewport] → [Component] → [Terminal]
   1M items      ~100 items     ~10 items    Rendered     What you see
```

## Quick Installation

**Prerequisites:** Go 1.19+ and basic familiarity with [Bubble Tea](https://github.com/charmbracelet/bubbletea)

```bash
go mod init your-project
go get github.com/davidroman0O/vtable
go get github.com/charmbracelet/bubbletea
```

That's it! You're ready to build your first component.

## What's Next?

Ready to build your first VTable component in 5 minutes?

**Next:** [Quick Start →](02-quick-start.md) 