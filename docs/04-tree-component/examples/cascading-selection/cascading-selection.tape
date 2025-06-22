# VTable Cascading Selection Example

Output "cascading-selection.gif"

Set FontSize 16
Set Width 800
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Expand the first project
Enter
Sleep 1s

# Navigate down to the 'src' folder and expand it
Down@500ms
Enter
Sleep 1s

# Navigate back up to the parent
Up@500ms
Sleep 1s

# Select the parent to trigger cascading selection
Space
Sleep 2s

# Collapse the project to show selection is maintained
Enter
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 