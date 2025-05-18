package vtable

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete visual theme for vtable components
type Theme struct {
	// Border styling
	BorderStyle       lipgloss.Style
	HeaderBorderStyle lipgloss.Style

	// Header styling
	HeaderStyle lipgloss.Style

	// Row styling
	RowStyle     lipgloss.Style
	RowEvenStyle lipgloss.Style // Optional, for alternating rows
	RowOddStyle  lipgloss.Style // Optional, for alternating rows

	// Cursor styling
	SelectedRowStyle lipgloss.Style

	// Special row indicators
	TopThresholdStyle    lipgloss.Style // Optional, for highlighting top threshold
	BottomThresholdStyle lipgloss.Style // Optional, for highlighting bottom threshold

	// Border characters
	BorderChars BorderCharacters
}

// BorderCharacters defines the characters used for table borders
type BorderCharacters struct {
	Horizontal  string // ─
	Vertical    string // │
	TopLeft     string // ┌
	TopRight    string // ┐
	BottomLeft  string // └
	BottomRight string // ┘
	LeftT       string // ├
	RightT      string // ┤
	TopT        string // ┬
	BottomT     string // ┴
	Cross       string // ┼
}

// DefaultBorderCharacters returns standard box drawing characters for borders
func DefaultBorderCharacters() BorderCharacters {
	return BorderCharacters{
		Horizontal:  "─",
		Vertical:    "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		LeftT:       "├",
		RightT:      "┤",
		TopT:        "┬",
		BottomT:     "┴",
		Cross:       "┼",
	}
}

// RoundedBorderCharacters returns rounded box drawing characters for borders
func RoundedBorderCharacters() BorderCharacters {
	return BorderCharacters{
		Horizontal:  "─",
		Vertical:    "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		LeftT:       "├",
		RightT:      "┤",
		TopT:        "┬",
		BottomT:     "┴",
		Cross:       "┼",
	}
}

// ThickBorderCharacters returns thick box drawing characters for borders
func ThickBorderCharacters() BorderCharacters {
	return BorderCharacters{
		Horizontal:  "━",
		Vertical:    "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		LeftT:       "┣",
		RightT:      "┫",
		TopT:        "┳",
		BottomT:     "┻",
		Cross:       "╋",
	}
}

// DoubleBorderCharacters returns double line box drawing characters for borders
func DoubleBorderCharacters() BorderCharacters {
	return BorderCharacters{
		Horizontal:  "═",
		Vertical:    "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		LeftT:       "╠",
		RightT:      "╣",
		TopT:        "╦",
		BottomT:     "╩",
		Cross:       "╬",
	}
}

// AsciiBoxCharacters returns ASCII characters for borders, useful for terminals that don't support Unicode
func AsciiBoxCharacters() BorderCharacters {
	return BorderCharacters{
		Horizontal:  "-",
		Vertical:    "|",
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		LeftT:       "+",
		RightT:      "+",
		TopT:        "+",
		BottomT:     "+",
		Cross:       "+",
	}
}

// DefaultTheme returns a basic default theme for vtable components
func DefaultTheme() Theme {
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	return Theme{
		BorderStyle:          borderStyle,
		HeaderBorderStyle:    borderStyle.Copy(),
		HeaderStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("238")),
		RowStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		RowEvenStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		RowOddStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("249")),
		SelectedRowStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("63")),
		TopThresholdStyle:    lipgloss.NewStyle(), // Default: no special styling
		BottomThresholdStyle: lipgloss.NewStyle(), // Default: no special styling
		BorderChars:          DefaultBorderCharacters(),
	}
}

// DarkTheme returns a dark-colored theme for vtable components
func DarkTheme() Theme {
	theme := DefaultTheme()
	theme.BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	theme.HeaderBorderStyle = theme.BorderStyle.Copy()
	theme.HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
	theme.RowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	theme.RowEvenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	theme.RowOddStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	theme.SelectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("25"))
	return theme
}

// LightTheme returns a light-colored theme for vtable components
func LightTheme() Theme {
	theme := DefaultTheme()
	theme.BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	theme.HeaderBorderStyle = theme.BorderStyle.Copy()
	theme.HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("234")).Background(lipgloss.Color("252"))
	theme.RowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("234"))
	theme.RowEvenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("234"))
	theme.RowOddStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	theme.SelectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("25"))
	return theme
}

// ColorfulTheme returns a colorful theme for vtable components
func ColorfulTheme() Theme {
	theme := DefaultTheme()
	theme.BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("105"))
	theme.HeaderBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	theme.HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("57"))
	theme.RowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	theme.RowEvenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	theme.RowOddStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("38"))
	theme.SelectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("161"))
	theme.TopThresholdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	theme.BottomThresholdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	return theme
}

// ThemeToStyleConfig converts a Theme to a StyleConfig (for backward compatibility)
func ThemeToStyleConfig(theme Theme) StyleConfig {
	return StyleConfig{
		BorderStyle:      theme.BorderStyle.String(),
		HeaderStyle:      theme.HeaderStyle.String(),
		RowStyle:         theme.RowStyle.String(),
		SelectedRowStyle: theme.SelectedRowStyle.String(),
	}
}

// StyleToTheme converts a StyleConfig to a Theme (for backward compatibility)
func StyleToTheme(style StyleConfig) Theme {
	theme := DefaultTheme()

	// We'll use simple conversion since StyleConfig is limited
	theme.BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.BorderStyle))
	theme.HeaderBorderStyle = theme.BorderStyle.Copy()
	theme.HeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.HeaderStyle))
	theme.SelectedRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.SelectedRowStyle))
	theme.RowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.RowStyle))

	return theme
}
