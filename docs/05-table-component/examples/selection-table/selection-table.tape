# VTable Table Selection Example

Output "selection-table.gif"

Set FontSize 16
Set Width 1000
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate and select a few items
Down@500ms
Down@500ms
Space
Sleep 1s
Down@500ms
Down@500ms
Space
Sleep 1s

# Show selection info
Type "s"
Sleep 2s

# Select All
Ctrl+A
Sleep 1s
Type "s"
Sleep 2s

# Clear Selection
Type "c"
Sleep 1s
Type "s"
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 