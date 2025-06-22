# VTable Virtualized Table Example

Output "virtualized-table.gif"

Set FontSize 16
Set Width 1000
Set Height 500
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Page down rapidly to show fast scrolling
PageDown@200ms
PageDown@200ms
PageDown@200ms
PageDown@200ms
PageDown@200ms
Sleep 1s

# Jump to a specific index
Type "J"
Sleep 500ms
Type "5000"
Sleep 1s
Enter
Sleep 2s

# Jump to the end
Type "G"
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 