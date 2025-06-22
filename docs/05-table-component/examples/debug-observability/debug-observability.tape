# VTable Debug and Observability Example

Output "debug-observability.gif"

Set FontSize 14
Set Width 1200
Set Height 700
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate down to trigger some chunk loading
Down@200ms
Down@200ms
Down@200ms
Down@200ms
Down@200ms
Down@200ms
Down@200ms
Down@200ms
Sleep 1s

# Navigate a bit more with the overlay on
PageDown
Sleep 1s

# Quit the application
Type "q"
Sleep 1s 