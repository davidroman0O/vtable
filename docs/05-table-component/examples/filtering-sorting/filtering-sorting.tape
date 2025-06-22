# VTable Table Filtering and Sorting Example

Output "filtering-sorting.gif"

Set FontSize 16
Set Width 1700
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Apply filters
Type "1"
Sleep 1s
Type "6"
Sleep 1s
Type "8"
Sleep 2s

# Clear filters
Type "0"
Sleep 2s

# Navigate to a column and sort
Right@500ms
Right@500ms
Type "s"
Sleep 1s
Type "s"
Sleep 1s

# Clear sort
Type "S"
Sleep 2s

# Search
Type "/"
Sleep 1s
Type "Developer"
Enter
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 