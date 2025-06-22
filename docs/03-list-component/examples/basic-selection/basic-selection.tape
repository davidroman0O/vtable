# VTable Basic Selection Example

Output "basic-selection.gif"

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

# Select an item
Space
Sleep 1s

# Navigate down further
Down@500ms
Down@500ms

# Select another item
Space
Sleep 1s

# Navigate up and deselect the first one
Up@500ms
Up@500ms
Up@500ms
Up@500ms
Space
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 