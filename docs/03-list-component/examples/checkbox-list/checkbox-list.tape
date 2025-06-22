# VTable Checkbox List Example

Output "checkbox-list.gif"

Set FontSize 16
Set Width 800
Set Height 500
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Select a few items
Down@500ms
Space
Sleep 500ms
Down@500ms
Down@500ms
Space
Sleep 1s

# Select all
Ctrl+A
Sleep 2s

# Clear all
Ctrl+D
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 