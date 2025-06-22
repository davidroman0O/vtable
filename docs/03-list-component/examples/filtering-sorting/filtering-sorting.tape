# VTable Filtering and Sorting Example

Output "filtering-sorting.gif"

Set FontSize 16
Set Width 1000
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Apply some filters
Type "1"
Sleep 1s
Type "6"
Sleep 1s
Type "8"
Sleep 2s

# Clear filters
Type "0"
Sleep 2s

# Apply a sort
Type "!"
Sleep 1s
Type "!"
Sleep 1s

# Sort by a different column
Type "@"
Sleep 2s

# Clear sorting
Type "S"
Sleep 1s

# Use search
Type "/"
Sleep 1s
Type "Developer"
Enter
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 