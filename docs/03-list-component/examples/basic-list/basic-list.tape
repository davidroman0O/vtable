# VTable Basic List Example

Output "basic-list.gif"

Set FontSize 16
Set Width 800
Set Height 400
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate down with arrow keys to show scrolling
Down@500ms
Down@500ms
Down@500ms
Down@500ms
Down@500ms
Sleep 1s

# Navigate up with k
Type "k"
Sleep 500ms
Type "k"
Sleep 500ms
Type "k"
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 