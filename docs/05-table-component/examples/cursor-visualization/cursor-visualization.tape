# VTable Cursor Visualization Example

Output "cursor-visualization.gif"

Set FontSize 16
Set Width 1200
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Start with Full Row Highlighting
Down@500ms
Sleep 1s

# Switch to Active Cell indication
Type "c"
Sleep 1s
Right@500ms
Right@500ms
Sleep 1s

# Cycle through active cell colors
Type "C"
Sleep 500ms
Type "C"
Sleep 500ms
Type "C"
Sleep 500ms
Type "C"
Sleep 1s

# Switch to Mixed Mode
Type "m"
Sleep 1s
Down@500ms
Left@500ms
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 