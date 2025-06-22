# VTable Advanced Tree Features Example

Output "advanced-features.gif"

Set FontSize 16
Set Width 800
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Expand entire tree
Type "E"
Sleep 2s

# Collapse entire tree
Type "C"
Sleep 2s

# Navigate and expand a subtree
Down@500ms
Down@500ms
Type "e"
Sleep 2s

# Navigate and select a folder to show cascading
Down@500ms
Down@500ms
Space
Sleep 2s

# Clear selection
Type "x"
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 