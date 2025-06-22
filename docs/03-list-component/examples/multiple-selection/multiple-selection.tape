# VTable Multiple Selection Example

Output "multiple-selection.gif"

Set FontSize 16
Set Width 800
Set Height 400
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Select a few items individually
Down@500ms
Space
Sleep 500ms
Down@500ms
Down@500ms
Space
Sleep 1s

# Select All
Ctrl+A
Sleep 2s

# Clear Selection
Ctrl+D
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 