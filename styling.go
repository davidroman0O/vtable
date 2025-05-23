package vtable

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines a complete color and style theme for the table
type Theme struct {
	Name string

	// Basic colors
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Foreground string

	// State colors
	Selected string
	Hover    string
	Active   string
	Disabled string
	Error    string
	Warning  string
	Success  string
	Info     string

	// Border styles
	BorderStyle       lipgloss.Border
	BorderColor       string
	HeaderBorder      lipgloss.Border
	HeaderBorderColor string
	BorderChars       BorderCharacters

	// Text styles
	HeaderStyle   lipgloss.Style
	CellStyle     lipgloss.Style
	SelectedStyle lipgloss.Style
	CursorStyle   lipgloss.Style
	DisabledStyle lipgloss.Style

	// Table-specific styles
	SelectedRowStyle  lipgloss.Style
	RowStyle          lipgloss.Style
	RowEvenStyle      lipgloss.Style
	RowOddStyle       lipgloss.Style
	HeaderBorderStyle lipgloss.Style

	// Status indicator styles
	LoadingStyle lipgloss.Style
	ErrorStyle   lipgloss.Style

	// Animation preferences
	AnimationDuration string
	ReducedMotion     bool
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

// DefaultTheme returns a sensible default theme
func DefaultTheme() *Theme {
	return &Theme{
		Name:              "default",
		Primary:           "#007ACC",
		Secondary:         "#666666",
		Accent:            "#FF6B35",
		Background:        "#FFFFFF",
		Foreground:        "#000000",
		Selected:          "#E6F3FF",
		Hover:             "#F0F8FF",
		Active:            "#B8DFFF",
		Disabled:          "#CCCCCC",
		Error:             "#FF4444",
		Warning:           "#FFA500",
		Success:           "#00AA00",
		Info:              "#0088CC",
		BorderStyle:       lipgloss.NormalBorder(),
		BorderColor:       "#CCCCCC",
		HeaderBorder:      lipgloss.NormalBorder(),
		HeaderBorderColor: "#999999",
		BorderChars:       DefaultBorderCharacters(),
		HeaderStyle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#333333")),
		CellStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")),
		SelectedStyle:     lipgloss.NewStyle().Background(lipgloss.Color("#E6F3FF")),
		CursorStyle:       lipgloss.NewStyle().Background(lipgloss.Color("#007ACC")).Foreground(lipgloss.Color("#FFFFFF")),
		DisabledStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		SelectedRowStyle:  lipgloss.NewStyle().Background(lipgloss.Color("#E6F3FF")).Bold(true),
		RowStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")),
		RowEvenStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")),
		RowOddStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
		HeaderBorderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")),
		LoadingStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true),
		ErrorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")),
		AnimationDuration: "200ms",
		ReducedMotion:     false,
	}
}

// DarkTheme returns a dark theme variant
func DarkTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "dark"
	theme.Background = "#1E1E1E"
	theme.Foreground = "#FFFFFF"
	theme.Primary = "#4A9EFF"
	theme.Secondary = "#999999"
	theme.Selected = "#2D3748"
	theme.Hover = "#374151"
	theme.Active = "#4A5568"
	theme.BorderColor = "#4A5568"
	theme.HeaderBorderColor = "#6B7280"
	theme.HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5E5E5"))
	theme.CellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	theme.SelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#2D3748"))
	theme.CursorStyle = lipgloss.NewStyle().Background(lipgloss.Color("#4A9EFF")).Foreground(lipgloss.Color("#000000"))
	theme.DisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	return theme
}

// HighContrastTheme returns a high contrast theme for accessibility
func HighContrastTheme() *Theme {
	theme := DefaultTheme()
	theme.Name = "high-contrast"
	theme.Background = "#FFFFFF"
	theme.Foreground = "#000000"
	theme.Primary = "#000000"
	theme.Secondary = "#000000"
	theme.Selected = "#000000"
	theme.Hover = "#EEEEEE"
	theme.Active = "#CCCCCC"
	theme.BorderColor = "#000000"
	theme.HeaderBorderColor = "#000000"
	theme.HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFFFF"))
	theme.CellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	theme.SelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#000000")).Foreground(lipgloss.Color("#FFFFFF"))
	theme.CursorStyle = lipgloss.NewStyle().Background(lipgloss.Color("#000000")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	theme.DisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	return theme
}

// StyleState contains rendering state information for applying themes
type StyleState struct {
	Selected  bool
	IsCursor  bool
	IsHeader  bool
	Disabled  bool
	Loading   bool
	Error     error
	IsHovered bool
	IsActive  bool
}

// ApplyTheme applies a theme to create a style for a specific state
func (t *Theme) ApplyTheme(state StyleState) lipgloss.Style {
	base := t.CellStyle

	switch {
	case state.Error != nil:
		base = base.Copy().Inherit(t.ErrorStyle)
	case state.Loading:
		base = base.Copy().Inherit(t.LoadingStyle)
	case state.Disabled:
		base = base.Copy().Inherit(t.DisabledStyle)
	case state.IsCursor:
		base = base.Copy().Inherit(t.CursorStyle)
	case state.Selected:
		base = base.Copy().Inherit(t.SelectedStyle)
	case state.IsHeader:
		base = base.Copy().Inherit(t.HeaderStyle)
	}

	return base
}

// ToStyleConfig converts a theme to a legacy StyleConfig
func (theme *Theme) ToStyleConfig() StyleConfig {
	return StyleConfig{
		BorderStyle:      theme.BorderColor,
		HeaderStyle:      theme.HeaderStyle.String(),
		RowStyle:         theme.CellStyle.String(),
		SelectedRowStyle: theme.SelectedStyle.String(),
	}
}

// ThemeToStyleConfig converts a theme to a legacy StyleConfig (global function)
func ThemeToStyleConfig(theme *Theme) StyleConfig {
	return theme.ToStyleConfig()
}

// FromStyleConfig creates a Theme from a legacy StyleConfig
func FromStyleConfig(style StyleConfig) *Theme {
	theme := DefaultTheme()

	// We'll use simple conversion since StyleConfig is limited
	theme.BorderStyle = lipgloss.NormalBorder()
	theme.HeaderBorder = lipgloss.NormalBorder()
	theme.HeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.HeaderStyle))
	theme.CellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.RowStyle))
	theme.SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(style.SelectedRowStyle))

	return theme
}
