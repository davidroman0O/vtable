# VTable Styled List Example

Output "styled-list.gif"

Set FontSize 16
Set Width 800
Set Height 500
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate down to show cursor styling
Down@500ms
Down@500ms
Sleep 1s

# Select an item to show selection styling
Space
Sleep 1s

# Navigate further to show both styles
Down@500ms
Down@500ms
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 