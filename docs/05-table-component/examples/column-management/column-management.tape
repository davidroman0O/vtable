# VTable Column Management Example

Output "column-management.gif"

Set FontSize 16
Set Width 1700
Set Height 500
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Move a column right using raw ANSI escape code for Ctrl+Right
Type "\x1b[1;5C"
Sleep 500ms
Type "\x1b[1;5C"
Sleep 1s

# Move a column left using raw ANSI escape code for Ctrl+Left
Type "\x1b[1;5D"
Sleep 1s

# Add a column
Type "+"
Sleep 1s

# Add another column
Type "+"
Sleep 1s

# Remove a column
Type "-"
Sleep 1s

# Adjust width of current column
Type "W"
Sleep 500ms
Type "W"
Sleep 500ms
Type "w"
Sleep 1s

# Cycle alignment
Type "A"
Sleep 500ms
Type "A"
Sleep 500ms
Type "A"
Sleep 1s

# Reset to default
Type "R"
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 