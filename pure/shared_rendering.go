package vtable

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// ================================
// SHARED RENDERING COMPONENTS
// ================================
// These components can be used across List, TreeList, and Table

// ================================
// RUNEWIDTH UTILITIES
// ================================

// MeasureText returns the actual display width of text, accounting for wide characters
func MeasureText(text string) int {
	return runewidth.StringWidth(text)
}

// TruncateText truncates text to fit within maxWidth, accounting for wide characters
func TruncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	textWidth := MeasureText(text)
	if textWidth <= maxWidth {
		return text
	}

	// Be much less aggressive - only truncate if significantly over
	if textWidth <= maxWidth+5 {
		return text
	}

	if maxWidth < 4 {
		return runewidth.Truncate(text, maxWidth, "")
	}

	// Use single ellipsis character instead of "..."
	return runewidth.Truncate(text, maxWidth-1, "") + "…"
}

// PadText pads text to exact width, accounting for wide characters
func PadText(text string, width int, alignment int) string {
	textWidth := MeasureText(text)

	// Allow overflow for small amounts - be much more lenient
	if textWidth > width && textWidth <= width+8 {
		return text
	}

	if textWidth > width {
		return TruncateText(text, width)
	}

	padding := width - textWidth
	if padding <= 0 {
		return text
	}

	switch alignment {
	case AlignLeft:
		return text + strings.Repeat(" ", padding)
	case AlignRight:
		return strings.Repeat(" ", padding) + text
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

// ================================
// PROGRESS BAR COMPONENT
// ================================

// ProgressBarConfig configures progress bar rendering
type ProgressBarConfig struct {
	Width       int
	ShowPercent bool
	EmptyChar   string
	FilledChar  string
	Style       lipgloss.Style
}

// DefaultProgressBarConfig returns sensible defaults
func DefaultProgressBarConfig() ProgressBarConfig {
	return ProgressBarConfig{
		Width:       20,
		ShowPercent: true,
		EmptyChar:   "░",
		FilledChar:  "█",
		Style:       lipgloss.NewStyle(),
	}
}

// RenderProgressBar renders a progress bar for any numeric value (0.0 to 1.0)
func RenderProgressBar(value float64, config ProgressBarConfig) string {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	// Calculate filled width
	filledWidth := int(math.Round(value * float64(config.Width)))
	emptyWidth := config.Width - filledWidth

	// Build progress bar
	var bar strings.Builder
	bar.WriteString(strings.Repeat(config.FilledChar, filledWidth))
	bar.WriteString(strings.Repeat(config.EmptyChar, emptyWidth))

	result := bar.String()

	// Add percentage if requested
	if config.ShowPercent {
		percent := fmt.Sprintf(" %3.0f%%", value*100)
		result += percent
	}

	return config.Style.Render(result)
}

// ================================
// BADGE COMPONENT
// ================================

// BadgeConfig configures badge rendering
type BadgeConfig struct {
	Style    lipgloss.Style
	Padding  int
	MinWidth int
	MaxWidth int
	Truncate bool
}

// DefaultBadgeConfig returns sensible defaults
func DefaultBadgeConfig() BadgeConfig {
	return BadgeConfig{
		Style:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		Padding:  1,
		MinWidth: 0,
		MaxWidth: 20,
		Truncate: true,
	}
}

// RenderBadge renders a styled badge with text
func RenderBadge(text string, config BadgeConfig) string {
	if config.Truncate && config.MaxWidth > 0 {
		text = TruncateText(text, config.MaxWidth-config.Padding*2)
	}

	// Apply minimum width if specified
	if config.MinWidth > 0 {
		textWidth := MeasureText(text) + config.Padding*2
		if textWidth < config.MinWidth {
			padding := config.MinWidth - textWidth
			text = text + strings.Repeat(" ", padding)
		}
	}

	return config.Style.Render(text)
}

// ================================
// STATUS INDICATOR COMPONENT
// ================================

// StatusConfig configures status indicator rendering
type StatusConfig struct {
	ActiveStyle   lipgloss.Style
	InactiveStyle lipgloss.Style
	ErrorStyle    lipgloss.Style
	LoadingStyle  lipgloss.Style
	ActiveText    string
	InactiveText  string
	ErrorText     string
	LoadingText   string
}

// DefaultStatusConfig returns sensible defaults
func DefaultStatusConfig() StatusConfig {
	return StatusConfig{
		ActiveStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		InactiveStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		ErrorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		LoadingStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
		ActiveText:    "●",
		InactiveText:  "○",
		ErrorText:     "✗",
		LoadingText:   "⟳",
	}
}

// Additional status constants for shared rendering (extending existing StatusType)
const (
	StatusActive StatusType = iota + 100 // Start after existing constants
	StatusInactive
	StatusLoading
)

// RenderStatus renders a status indicator
func RenderStatus(status StatusType, config StatusConfig) string {
	switch status {
	case StatusActive:
		return config.ActiveStyle.Render(config.ActiveText)
	case StatusInactive:
		return config.InactiveStyle.Render(config.InactiveText)
	case StatusError: // Use existing StatusError from messages.go
		return config.ErrorStyle.Render(config.ErrorText)
	case StatusLoading:
		return config.LoadingStyle.Render(config.LoadingText)
	default:
		return config.InactiveStyle.Render(config.InactiveText)
	}
}

// ================================
// NUMERIC FORMATTING COMPONENT
// ================================

// NumericConfig configures numeric value rendering
type NumericConfig struct {
	DecimalPlaces  int
	ThousandsSep   string
	CurrencySymbol string
	ShowPositive   bool
	Style          lipgloss.Style
}

// DefaultNumericConfig returns sensible defaults
func DefaultNumericConfig() NumericConfig {
	return NumericConfig{
		DecimalPlaces:  2,
		ThousandsSep:   ",",
		CurrencySymbol: "",
		ShowPositive:   false,
		Style:          lipgloss.NewStyle(),
	}
}

// RenderNumeric renders a numeric value with formatting
func RenderNumeric(value float64, config NumericConfig) string {
	// Format the number
	format := fmt.Sprintf("%%.%df", config.DecimalPlaces)
	text := fmt.Sprintf(format, value)

	// Add thousands separator if specified
	if config.ThousandsSep != "" {
		text = addThousandsSeparator(text, config.ThousandsSep)
	}

	// Add currency symbol if specified
	if config.CurrencySymbol != "" {
		if value >= 0 {
			text = config.CurrencySymbol + text
		} else {
			text = "-" + config.CurrencySymbol + text[1:] // Remove negative sign and add currency
		}
	}

	// Add positive sign if requested
	if config.ShowPositive && value > 0 && config.CurrencySymbol == "" {
		text = "+" + text
	}

	return config.Style.Render(text)
}

// addThousandsSeparator adds thousands separators to a numeric string
func addThousandsSeparator(text, separator string) string {
	// Find decimal point
	parts := strings.Split(text, ".")
	intPart := parts[0]

	// Handle negative numbers
	negative := strings.HasPrefix(intPart, "-")
	if negative {
		intPart = intPart[1:]
	}

	// Add separators
	if len(intPart) > 3 {
		var result strings.Builder
		for i, char := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteString(separator)
			}
			result.WriteRune(char)
		}
		intPart = result.String()
	}

	// Reconstruct number
	if negative {
		intPart = "-" + intPart
	}

	if len(parts) > 1 {
		return intPart + "." + parts[1]
	}
	return intPart
}

// ================================
// CELL CONSTRAINT UTILITIES
// ================================

// CellRenderOptions defines options for rendering content within cell constraints
type CellRenderOptions struct {
	Width     int
	Height    int
	Alignment int
	Padding   PaddingConfig
	Style     lipgloss.Style
	Truncate  bool
	Wrap      bool
}

// RenderInCell renders content within cell constraints with proper width handling
func RenderInCell(content string, options CellRenderOptions) CellRenderResult {
	// Calculate available width after padding
	availableWidth := options.Width - options.Padding.Left - options.Padding.Right
	if availableWidth <= 0 {
		return CellRenderResult{
			Content:      "",
			ActualWidth:  options.Width,
			ActualHeight: 1,
			Overflow:     true,
		}
	}

	// Handle content based on options
	var processedContent string
	overflow := false

	if options.Wrap && options.Height > 1 {
		// Multi-line wrapping (future feature)
		processedContent = content
		if MeasureText(content) > availableWidth {
			processedContent = TruncateText(content, availableWidth)
			overflow = true
		}
	} else {
		// Single line with optional truncation
		if MeasureText(content) > availableWidth {
			if options.Truncate {
				processedContent = TruncateText(content, availableWidth)
				overflow = true
			} else {
				processedContent = content
				overflow = true
			}
		} else {
			processedContent = content
		}
	}

	// Apply alignment and padding
	paddedContent := PadText(processedContent, availableWidth, options.Alignment)

	// Add left and right padding
	if options.Padding.Left > 0 {
		paddedContent = strings.Repeat(" ", options.Padding.Left) + paddedContent
	}
	if options.Padding.Right > 0 {
		paddedContent = paddedContent + strings.Repeat(" ", options.Padding.Right)
	}

	// Apply styling
	finalContent := options.Style.Render(paddedContent)

	return CellRenderResult{
		Content:      finalContent,
		ActualWidth:  options.Width,
		ActualHeight: 1,
		Overflow:     overflow,
	}
}

// ================================
// SHARED COMPONENT FACTORY
// ================================

// ComponentFactory provides easy access to shared rendering components
type ComponentFactory struct {
	ProgressBarConfig ProgressBarConfig
	BadgeConfig       BadgeConfig
	StatusConfig      StatusConfig
	NumericConfig     NumericConfig
}

// NewComponentFactory creates a new component factory with default configs
func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{
		ProgressBarConfig: DefaultProgressBarConfig(),
		BadgeConfig:       DefaultBadgeConfig(),
		StatusConfig:      DefaultStatusConfig(),
		NumericConfig:     DefaultNumericConfig(),
	}
}

// ProgressBar renders a progress bar using the factory's config
func (cf *ComponentFactory) ProgressBar(value float64) string {
	return RenderProgressBar(value, cf.ProgressBarConfig)
}

// Badge renders a badge using the factory's config
func (cf *ComponentFactory) Badge(text string) string {
	return RenderBadge(text, cf.BadgeConfig)
}

// Status renders a status indicator using the factory's config
func (cf *ComponentFactory) Status(status StatusType) string {
	return RenderStatus(status, cf.StatusConfig)
}

// Numeric renders a numeric value using the factory's config
func (cf *ComponentFactory) Numeric(value float64) string {
	return RenderNumeric(value, cf.NumericConfig)
}
