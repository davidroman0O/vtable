# Formatted Items Example

This example shows how to display rich, structured data with custom formatting instead of simple strings.

## What it shows

- **Rich Data Structures**: Using Person structs with Name, Age, City, and Job fields
- **Custom Formatter Function**: Creating a personFormatter that formats data for display
- **Type Assertion**: Casting `data.Item.(Person)` to access structured data fields
- **Conditional Formatting**: Different age displays with emojis (🌟 for young, 👑 for senior)
- **Rich Display**: Showing multiple fields in a formatted layout: "Name (Age) - Job in City"
- **All Features Work**: Navigation, selection, and status tracking work with formatted data

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

## What you'll see

Instead of "Item 1", "Item 2", you'll see rich formatted entries like:
- `Alice Johnson (28) 🌟 - UX Designer in San Francisco`
- `Emma Wilson (52) 👑 - Tech Lead in Portland`
- `Bob Chen (34) - Software Engineer in New York`

## Key learnings

- How to pass structured data through the DataSource
- Creating and applying custom formatter functions
- Type assertion to access your data structure fields
- Conditional formatting based on data values
- Integrating formatting with selection and navigation features
- The formatter function signature and parameters
- How VTable handles rich data while maintaining performance 