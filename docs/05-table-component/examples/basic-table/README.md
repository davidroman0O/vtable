# Basic Table Example

A simple employee directory table demonstrating the fundamentals of VTable's table component.

## Features Demonstrated

- **Table structure** with multiple columns (Name, Department, Status, Salary)
- **Column configuration** with different alignments (left, center, right)
- **DataSource implementation** following VTable's async pattern
- **Keyboard navigation** (↑↓, j/k, g/G for start/end)
- **Component rendering** with clean cursor indicators (►)
- **Proper message handling** with status updates

## Running the Example

```bash
cd docs/05-table-component/examples/basic-table
go run .
```

## Controls

| Key | Action |
|-----|--------|
| `↑` `k` | Move up one row |
| `↓` `j` | Move down one row |
| `g` | Jump to first row |
| `G` | Jump to last row |
| `q` | Quit application |

## What You'll See

A table with 12 employees showing:
- Employee names (left-aligned)
- Departments (center-aligned) 
- Status (center-aligned)
- Salaries (right-aligned)

The cursor indicator (►) shows your current position, and the status bar shows position information as you navigate.

## Code Structure

- **DataSource**: `EmployeeDataSource` implements all required TableDataSource methods
- **Columns**: Configured with widths, alignments, and field mappings
- **Navigation**: Proper message handling with status updates
- **Component rendering**: Clean visual indicators without content pollution

This example serves as the foundation for all advanced table features covered in subsequent documentation sections. 