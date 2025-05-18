# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Require:
#   Require <string>                Ensure a program is on the $PATH to proceed
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set LoopOffset <float>%         Set the starting frame offset for the GIF loop
#   Set Theme <json|string>         Set the theme of the terminal
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#   Set MarginFill <file|#000000>   Set the file or color the margin will be filled with.
#   Set Margin <number>             Set the size of the margin. Has no effect if MarginFill isn't set.
#   Set BorderRadius <number>       Set terminal border radius, in pixels.
#   Set WindowBar <string>          Set window bar type. (one of: Rings, RingsRight, Colorful, ColorfulRight)
#   Set WindowBarSize <number>      Set window bar size, in pixels. Default is 40.
#   Set TypingSpeed <time>          Set the typing speed of the terminal. Default is 50ms.
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Escape[@<time>] [number]        Press the Escape key
#   Backspace[@<time>] [number]     Press the Backspace key
#   Delete[@<time>] [number]        Press the Delete key
#   Insert[@<time>] [number]        Press the Insert key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   PageUp[@<time>] [number]        Press the Page Up key
#   PageDown[@<time>] [number]      Press the Page Down key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output demo.gif

Require go

Set FontSize 16
Set Width 1200
Set Height 600
Set Padding 20
Set BorderRadius 10
Set WindowBar Colorful
Set LoopOffset 50%
Set PlaybackSpeed 1.0
Set MarginFill "#2D3748"
Set Margin 20

# Start
Hide
Type "cd examples" Enter
Sleep 500ms
Show

# Run the example
Type "go run ." Enter
Sleep 2s

# Scroll down
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Sleep 1s

# Scroll down quickly to demonstrate virtualization
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Sleep 1.5s

# Page down a few times
Type "d" Sleep 1s
Type "d" Sleep 1s
Type "d" Sleep 1s

# Jump to end to show we can handle large datasets
Type "G" Sleep 2s

# Page up
Type "u" Sleep 1s
Type "u" Sleep 1s

# Jump to beginning
Type "g" Sleep 2s

# Switch to table view
Tab Sleep 2s

# Scroll down in table
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Type "j" Sleep 200ms
Sleep 1s

# Scroll down quickly in table
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Type "j" Sleep 50ms
Sleep 1.5s

# Page down in table
Type "d" Sleep 1s
Type "d" Sleep 1s

# Search for item 500
Type "f" Sleep 1s
Type "500" Sleep 1s
Enter Sleep 2s

# Change theme
Type "t" Sleep 1.5s
Type "t" Sleep 1.5s
Type "t" Sleep 1.5s

# Scroll up
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Type "k" Sleep 50ms
Sleep 1.5s

# Jump to start
Type "g" Sleep 2s

# End the demo
Type "q" Sleep 1s
