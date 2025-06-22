# VTable Basic Tree Example

Output "basic-tree.gif"

Set FontSize 16
Set Width 800
Set Height 400
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate down
Down@500ms
Down@500ms
Down@500ms
Sleep 1s

# Expand a node
Enter
Sleep 1s

# Navigate into the expanded node
Down@500ms
Down@500ms
Enter
Sleep 2s

# Collapse the parent node
Up@500ms
Up@500ms
Enter
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 