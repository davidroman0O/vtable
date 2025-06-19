// Package render provides a collection of utility functions for rendering vtable
// components. It encapsulates common rendering logic, such as applying styles,
// formatting content, and handling different item states (e.g., loading, error,
// selected). This package promotes consistency and simplifies the rendering
// process within individual components like List and Table.
package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/mattn/go-runewidth"
)

// MeasureText calculates the visible width of a string, correctly handling
// East Asian wide characters. This is crucial for accurate layout and alignment
// in terminal UIs.
func MeasureText(text string) int {
	return runewidth.StringWidth(text)
}

// TruncateText shortens a string to a specified maximum width, appending an
// ellipsis (...) if the string is cut. It is aware of wide characters, ensuring
// the final string does not exceed the visual width limit.
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

// PadText adjusts a string to an exact width by adding padding. It supports
// left, right, and center alignment and is aware of wide characters to ensure
// correct visual alignment. If the text exceeds the width, it is truncated.
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
	case core.AlignLeft:
		return text + strings.Repeat(" ", padding)
	case core.AlignRight:
		return strings.Repeat(" ", padding) + text
	case core.AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

// ProgressBarConfig holds the configuration for rendering a progress bar.
type ProgressBarConfig struct {
	Width       int            // The total width of the progress bar in characters.
	ShowPercent bool           // Whether to display the percentage text next to the bar.
	EmptyChar   string         // The character used for the empty part of the bar.
	FilledChar  string         // The character used for the filled part of the bar.
	Style       lipgloss.Style // The style to apply to the entire progress bar string.
}

// DefaultProgressBarConfig returns a `ProgressBarConfig` with sensible default values.
func DefaultProgressBarConfig() ProgressBarConfig {
	return ProgressBarConfig{
		Width:       20,
		ShowPercent: true,
		EmptyChar:   "░",
		FilledChar:  "█",
		Style:       lipgloss.NewStyle(),
	}
}

// RenderProgressBar generates a string representation of a progress bar based on
// a given value (from 0.0 to 1.0) and a configuration.
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

// BadgeConfig holds the configuration for rendering a badge.
type BadgeConfig struct {
	Style    lipgloss.Style // The style for the badge's background and text.
	Padding  int            // The number of padding characters on each side of the text.
	MinWidth int            // The minimum width of the badge.
	MaxWidth int            // The maximum width of the badge before truncation.
	Truncate bool           // Whether to truncate the text if it exceeds MaxWidth.
}

// DefaultBadgeConfig returns a `BadgeConfig` with sensible default values.
func DefaultBadgeConfig() BadgeConfig {
	return BadgeConfig{
		Style:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		Padding:  1,
		MinWidth: 0,
		MaxWidth: 20,
		Truncate: true,
	}
}

// RenderBadge generates a styled badge string with the given text and configuration.
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

// StatusConfig holds the configuration for rendering a status indicator.
type StatusConfig struct {
	ActiveStyle   lipgloss.Style // Style for the 'Active' state.
	InactiveStyle lipgloss.Style // Style for the 'Inactive' state.
	ErrorStyle    lipgloss.Style // Style for the 'Error' state.
	LoadingStyle  lipgloss.Style // Style for the 'Loading' state.
	ActiveText    string         // Text/icon for the 'Active' state.
	InactiveText  string         // Text/icon for the 'Inactive' state.
	ErrorText     string         // Text/icon for the 'Error' state.
	LoadingText   string         // Text/icon for the 'Loading' state.
}

// DefaultStatusConfig returns a `StatusConfig` with sensible default values.
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

// Additional status constants for shared rendering, extending core.StatusType.
const (
	StatusActive core.StatusType = iota + 100 // Start after existing constants
	StatusInactive
	StatusLoading
)

// RenderStatus generates a styled status indicator string based on the given
// status and configuration.
func RenderStatus(status core.StatusType, config StatusConfig) string {
	switch status {
	case StatusActive:
		return config.ActiveStyle.Render(config.ActiveText)
	case StatusInactive:
		return config.InactiveStyle.Render(config.InactiveText)
	case core.StatusError: // Use existing StatusError from messages.go
		return config.ErrorStyle.Render(config.ErrorText)
	case StatusLoading:
		return config.LoadingStyle.Render(config.LoadingText)
	default:
		return config.InactiveStyle.Render(config.InactiveText)
	}
}

// NumericConfig holds the configuration for rendering a numeric value.
type NumericConfig struct {
	DecimalPlaces  int            // The number of decimal places to display.
	ThousandsSep   string         // The character to use as a thousands separator.
	CurrencySymbol string         // The currency symbol to prepend to the value.
	ShowPositive   bool           // Whether to show a '+' sign for positive numbers.
	Style          lipgloss.Style // The style to apply to the formatted number.
}

// DefaultNumericConfig returns a `NumericConfig` with sensible default values.
func DefaultNumericConfig() NumericConfig {
	return NumericConfig{
		DecimalPlaces:  2,
		ThousandsSep:   ",",
		CurrencySymbol: "",
		ShowPositive:   false,
		Style:          lipgloss.NewStyle(),
	}
}

// RenderNumeric generates a styled string representation of a numeric value
// based on the given configuration. It handles decimal places, thousands
// separators, currency symbols, and optional positive signs.
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
	if config.ShowPositive && value > 0 {
		text = "+" + text
	}

	return config.Style.Render(text)
}

// addThousandsSeparator is a helper function to insert thousands separators
// into a numeric string.
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

// CellRenderOptions defines the constraints and styles for rendering content
// within a table cell.
type CellRenderOptions struct {
	Width     int                // The exact width of the cell.
	Height    int                // The exact height of the cell (usually 1).
	Alignment int                // Text alignment (AlignLeft, AlignCenter, AlignRight).
	Padding   core.PaddingConfig // Padding within the cell.
	Style     lipgloss.Style     // The base style for the cell content.
	Truncate  bool               // Whether to truncate content that overflows the width.
	Wrap      bool               // Whether to wrap content that overflows the width.
}

// RenderInCell renders content within the constraints of a cell, handling
// padding, alignment, truncation, and styling. It returns a `CellRenderResult`
// containing the rendered content and metadata about the rendering process.
func RenderInCell(content string, options CellRenderOptions) core.CellRenderResult {
	// Calculate available width after padding
	availableWidth := options.Width - options.Padding.Left - options.Padding.Right
	if availableWidth <= 0 {
		return core.CellRenderResult{
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

	return core.CellRenderResult{
		Content:      finalContent,
		ActualWidth:  options.Width,
		ActualHeight: 1,
		Overflow:     overflow,
	}
}

// ComponentFactory provides a convenient way to create common UI components
// (like progress bars and badges) with a consistent, pre-configured style.
type ComponentFactory struct {
	ProgressBarConfig ProgressBarConfig // Configuration for progress bars created by this factory.
	BadgeConfig       BadgeConfig       // Configuration for badges created by this factory.
	StatusConfig      StatusConfig      // Configuration for status indicators created by this factory.
	NumericConfig     NumericConfig     // Configuration for numeric values created by this factory.
}

// NewComponentFactory creates a new `ComponentFactory` with default configurations.
func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{
		ProgressBarConfig: DefaultProgressBarConfig(),
		BadgeConfig:       DefaultBadgeConfig(),
		StatusConfig:      DefaultStatusConfig(),
		NumericConfig:     DefaultNumericConfig(),
	}
}

// ProgressBar creates a progress bar string using the factory's configuration.
func (cf *ComponentFactory) ProgressBar(value float64) string {
	return RenderProgressBar(value, cf.ProgressBarConfig)
}

// Badge creates a badge string using the factory's configuration.
func (cf *ComponentFactory) Badge(text string) string {
	return RenderBadge(text, cf.BadgeConfig)
}

// Status creates a status indicator string using the factory's configuration.
func (cf *ComponentFactory) Status(status core.StatusType) string {
	return RenderStatus(status, cf.StatusConfig)
}

// Numeric creates a formatted numeric string using the factory's configuration.
func (cf *ComponentFactory) Numeric(value float64) string {
	return RenderNumeric(value, cf.NumericConfig)
}
