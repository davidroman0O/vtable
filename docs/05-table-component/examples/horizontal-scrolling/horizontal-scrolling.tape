# VTable Horizontal Scrolling Example

Output "horizontal-scrolling.gif"

Set FontSize 16
Set Width 1700
Set Height 600
Set Shell "bash"

Type "go run ."
Enter
Sleep 2s

# Navigate columns with arrow keys
Right@500ms
Right@500ms
Right@500ms
Right@500ms
Right@500ms
Sleep 1s

# Scroll right within the Description column using '.'
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Type ","
Sleep 150ms
Sleep 1s

# Scroll left within the Description column using ','
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Type "."
Sleep 150ms
Sleep 1s

# Reset scroll
Backspace
Sleep 2s

# Quit the application
Type "q"
Sleep 1s 