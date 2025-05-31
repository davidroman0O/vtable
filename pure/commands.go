package vtable

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// NAVIGATION COMMANDS
// ================================

// CursorUpCmd returns a command to move the cursor up
func CursorUpCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorUpMsg{}
	}
}

// CursorDownCmd returns a command to move the cursor down
func CursorDownCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorDownMsg{}
	}
}

// PageUpCmd returns a command to move up one page
func PageUpCmd() tea.Cmd {
	return func() tea.Msg {
		return PageUpMsg{}
	}
}

// PageDownCmd returns a command to move down one page
func PageDownCmd() tea.Cmd {
	return func() tea.Msg {
		return PageDownMsg{}
	}
}

// JumpToStartCmd returns a command to jump to the start
func JumpToStartCmd() tea.Cmd {
	return func() tea.Msg {
		return JumpToStartMsg{}
	}
}

// JumpToEndCmd returns a command to jump to the end
func JumpToEndCmd() tea.Cmd {
	return func() tea.Msg {
		return JumpToEndMsg{}
	}
}

// JumpToCmd returns a command to jump to a specific index
func JumpToCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return JumpToMsg{Index: index}
	}
}

// TreeJumpToIndexCmd returns a command to jump to a specific index in a tree with optional parent expansion
func TreeJumpToIndexCmd(index int, expandParents bool) tea.Cmd {
	return func() tea.Msg {
		return TreeJumpToIndexMsg{
			Index:         index,
			ExpandParents: expandParents,
		}
	}
}

// ================================
// DATA COMMANDS
// ================================

// DataRefreshCmd returns a command to refresh all data
func DataRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return DataRefreshMsg{}
	}
}

// DataChunksRefreshCmd returns a command to refresh chunks while preserving cursor position
func DataChunksRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return DataChunksRefreshMsg{}
	}
}

// DataChunkLoadedCmd returns a command indicating a chunk was loaded
func DataChunkLoadedCmd(startIndex int, items []Data[any], request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return DataChunkLoadedMsg{
			StartIndex: startIndex,
			Items:      items,
			Request:    request,
		}
	}
}

// DataChunkErrorCmd returns a command indicating a chunk failed to load
func DataChunkErrorCmd(startIndex int, err error, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return DataChunkErrorMsg{
			StartIndex: startIndex,
			Error:      err,
			Request:    request,
		}
	}
}

// DataTotalCmd returns a command with the total number of items
func DataTotalCmd(total int) tea.Cmd {
	return func() tea.Msg {
		return DataTotalMsg{Total: total}
	}
}

// DataTotalUpdateCmd returns a command with the total number of items while preserving cursor position
func DataTotalUpdateCmd(total int) tea.Cmd {
	return func() tea.Msg {
		return DataTotalUpdateMsg{Total: total}
	}
}

// DataLoadErrorCmd returns a command indicating a data loading error
func DataLoadErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return DataLoadErrorMsg{Error: err}
	}
}

// DataSourceSetCmd returns a command to set a new data source
func DataSourceSetCmd(dataSource DataSource[any]) tea.Cmd {
	return func() tea.Msg {
		return DataSourceSetMsg{DataSource: dataSource}
	}
}

// ChunkUnloadedCmd returns a command indicating a chunk was unloaded from memory
func ChunkUnloadedCmd(chunkStart int) tea.Cmd {
	return func() tea.Msg {
		return ChunkUnloadedMsg{ChunkStart: chunkStart}
	}
}

// ChunkLoadingStartedCmd returns a command indicating a chunk has started loading
func ChunkLoadingStartedCmd(chunkStart int, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return ChunkLoadingStartedMsg{
			ChunkStart: chunkStart,
			Request:    request,
		}
	}
}

// ChunkLoadingCompletedCmd returns a command indicating a chunk has finished loading
func ChunkLoadingCompletedCmd(chunkStart int, itemCount int, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return ChunkLoadingCompletedMsg{
			ChunkStart: chunkStart,
			ItemCount:  itemCount,
			Request:    request,
		}
	}
}

// DataTotalRequestCmd returns a command to request the total count of items
func DataTotalRequestCmd() tea.Cmd {
	return func() tea.Msg {
		return DataTotalRequestMsg{}
	}
}

// ================================
// SELECTION COMMANDS
// ================================

// SelectCurrentCmd returns a command to select the current item
func SelectCurrentCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectCurrentMsg{}
	}
}

// SelectToggleCmd returns a command to toggle selection of an item
func SelectToggleCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SelectToggleMsg{Index: index}
	}
}

// SelectAllCmd returns a command to select all items
func SelectAllCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectAllMsg{}
	}
}

// SelectClearCmd returns a command to clear all selections
func SelectClearCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectClearMsg{}
	}
}

// SelectRangeCmd returns a command to select a range of items
func SelectRangeCmd(startID, endID string) tea.Cmd {
	return func() tea.Msg {
		return SelectRangeMsg{
			StartID: startID,
			EndID:   endID,
		}
	}
}

// SelectionModeSetCmd returns a command to set the selection mode
func SelectionModeSetCmd(mode SelectionMode) tea.Cmd {
	return func() tea.Msg {
		return SelectionModeSetMsg{Mode: mode}
	}
}

// SelectionResponseCmd returns a command with the result of a selection operation
func SelectionResponseCmd(success bool, index int, id string, selected bool, operation string, err error, affectedIDs []string) tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{
			Success:     success,
			Index:       index,
			ID:          id,
			Selected:    selected,
			Operation:   operation,
			Error:       err,
			AffectedIDs: affectedIDs,
		}
	}
}

// SelectionChangedCmd returns a command indicating selection state has changed
func SelectionChangedCmd(selectedIndices []int, selectedIDs []string, totalSelected int) tea.Cmd {
	return func() tea.Msg {
		return SelectionChangedMsg{
			SelectedIndices: selectedIndices,
			SelectedIDs:     selectedIDs,
			TotalSelected:   totalSelected,
		}
	}
}

// ================================
// FILTER COMMANDS
// ================================

// FilterSetCmd returns a command to set a filter
func FilterSetCmd(field string, value any) tea.Cmd {
	return func() tea.Msg {
		return FilterSetMsg{
			Field: field,
			Value: value,
		}
	}
}

// FilterClearCmd returns a command to clear a filter
func FilterClearCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return FilterClearMsg{Field: field}
	}
}

// FiltersClearAllCmd returns a command to clear all filters
func FiltersClearAllCmd() tea.Cmd {
	return func() tea.Msg {
		return FiltersClearAllMsg{}
	}
}

// ================================
// SORT COMMANDS
// ================================

// SortToggleCmd returns a command to toggle sorting on a field
func SortToggleCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return SortToggleMsg{Field: field}
	}
}

// SortSetCmd returns a command to set sorting on a field
func SortSetCmd(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortSetMsg{
			Field:     field,
			Direction: direction,
		}
	}
}

// SortAddCmd returns a command to add a sort field
func SortAddCmd(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortAddMsg{
			Field:     field,
			Direction: direction,
		}
	}
}

// SortRemoveCmd returns a command to remove a sort field
func SortRemoveCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return SortRemoveMsg{Field: field}
	}
}

// SortsClearAllCmd returns a command to clear all sorting
func SortsClearAllCmd() tea.Cmd {
	return func() tea.Msg {
		return SortsClearAllMsg{}
	}
}

// ================================
// FOCUS COMMANDS
// ================================

// FocusCmd returns a command to give focus to the component
func FocusCmd() tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{}
	}
}

// BlurCmd returns a command to remove focus from the component
func BlurCmd() tea.Cmd {
	return func() tea.Msg {
		return BlurMsg{}
	}
}

// ================================
// ANIMATION COMMANDS
// ================================

// GlobalAnimationTickCmd returns a command for the global animation ticker
func GlobalAnimationTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return GlobalAnimationTickMsg{Timestamp: time.Now()}
	})
}

// AnimationUpdateCmd returns a command indicating animations were updated
func AnimationUpdateCmd(updatedAnimations []string) tea.Cmd {
	return func() tea.Msg {
		return AnimationUpdateMsg{UpdatedAnimations: updatedAnimations}
	}
}

// AnimationConfigCmd returns a command to set animation configuration
func AnimationConfigCmd(config AnimationConfig) tea.Cmd {
	return func() tea.Msg {
		return AnimationConfigMsg{Config: config}
	}
}

// AnimationStartCmd returns a command to start a specific animation
func AnimationStartCmd(animationID string) tea.Cmd {
	return func() tea.Msg {
		return AnimationStartMsg{AnimationID: animationID}
	}
}

// AnimationStopCmd returns a command to stop a specific animation
func AnimationStopCmd(animationID string) tea.Cmd {
	return func() tea.Msg {
		return AnimationStopMsg{AnimationID: animationID}
	}
}

// ================================
// THEME COMMANDS
// ================================

// ThemeSetCmd returns a command to set the theme
func ThemeSetCmd(theme interface{}) tea.Cmd {
	return func() tea.Msg {
		return ThemeSetMsg{Theme: theme}
	}
}

// ================================
// REAL-TIME UPDATE COMMANDS
// ================================

// RealTimeUpdateCmd returns a command for real-time updates
func RealTimeUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		return RealTimeUpdateMsg{}
	}
}

// RealTimeConfigCmd returns a command to configure real-time updates
func RealTimeConfigCmd(enabled bool, interval time.Duration) tea.Cmd {
	return func() tea.Msg {
		return RealTimeConfigMsg{
			Enabled:  enabled,
			Interval: interval,
		}
	}
}

// ================================
// VIEWPORT COMMANDS
// ================================

// ViewportResizeCmd returns a command indicating viewport resize
func ViewportResizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		return ViewportResizeMsg{
			Width:  width,
			Height: height,
		}
	}
}

// ViewportConfigCmd returns a command to set viewport configuration
func ViewportConfigCmd(config ViewportConfig) tea.Cmd {
	return func() tea.Msg {
		return ViewportConfigMsg{Config: config}
	}
}

// ================================
// TABLE-SPECIFIC COMMANDS
// ================================

// ColumnSetCmd returns a command to set table columns
func ColumnSetCmd(columns []TableColumn) tea.Cmd {
	return func() tea.Msg {
		return ColumnSetMsg{Columns: columns}
	}
}

// ColumnUpdateCmd returns a command to update a specific column
func ColumnUpdateCmd(index int, column TableColumn) tea.Cmd {
	return func() tea.Msg {
		return ColumnUpdateMsg{
			Index:  index,
			Column: column,
		}
	}
}

// HeaderVisibilityCmd returns a command to set header visibility
func HeaderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return HeaderVisibilityMsg{Visible: visible}
	}
}

// BorderVisibilityCmd returns a command to set border visibility
func BorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return BorderVisibilityMsg{Visible: visible}
	}
}

// TopBorderVisibilityCmd returns a command to set top border visibility
func TopBorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return TopBorderVisibilityMsg{Visible: visible}
	}
}

// BottomBorderVisibilityCmd returns a command to set bottom border visibility
func BottomBorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return BottomBorderVisibilityMsg{Visible: visible}
	}
}

// HeaderSeparatorVisibilityCmd returns a command to set header separator visibility
func HeaderSeparatorVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return HeaderSeparatorVisibilityMsg{Visible: visible}
	}
}

// TopBorderSpaceRemovalCmd returns a command to control top border space removal
func TopBorderSpaceRemovalCmd(remove bool) tea.Cmd {
	return func() tea.Msg {
		return TopBorderSpaceRemovalMsg{Remove: remove}
	}
}

// BottomBorderSpaceRemovalCmd returns a command to control bottom border space removal
func BottomBorderSpaceRemovalCmd(remove bool) tea.Cmd {
	return func() tea.Msg {
		return BottomBorderSpaceRemovalMsg{Remove: remove}
	}
}

// ActiveCellIndicationModeSetCmd returns a command to set the active cell indication mode
func ActiveCellIndicationModeSetCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return ActiveCellIndicationModeSetMsg{Enabled: enabled}
	}
}

// ActiveCellBackgroundColorSetCmd returns a command to set the active cell background color
func ActiveCellBackgroundColorSetCmd(color string) tea.Cmd {
	return func() tea.Msg {
		return ActiveCellBackgroundColorSetMsg{Color: color}
	}
}

// CellFormatterSetCmd returns a command to set a cell formatter
func CellFormatterSetCmd(columnIndex int, formatter SimpleCellFormatter) tea.Cmd {
	return func() tea.Msg {
		return CellFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// CellAnimatedFormatterSetCmd returns a command to set an animated cell formatter
func CellAnimatedFormatterSetCmd(columnIndex int, formatter CellFormatterAnimated) tea.Cmd {
	return func() tea.Msg {
		return CellAnimatedFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// RowFormatterSetCmd returns a command to set a loading row formatter
func RowFormatterSetCmd(formatter LoadingRowFormatter) tea.Cmd {
	return func() tea.Msg {
		return RowFormatterSetMsg{Formatter: formatter}
	}
}

// HeaderFormatterSetCmd returns a command to set a header formatter
func HeaderFormatterSetCmd(columnIndex int, formatter SimpleHeaderFormatter) tea.Cmd {
	return func() tea.Msg {
		return HeaderFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// LoadingFormatterSetCmd returns a command to set a loading row formatter (DEPRECATED)
func LoadingFormatterSetCmd(formatter LoadingRowFormatter) tea.Cmd {
	return func() tea.Msg {
		return LoadingFormatterSetMsg{Formatter: formatter}
	}
}

// HeaderCellFormatterSetCmd returns a command to set a header cell formatter (DEPRECATED)
func HeaderCellFormatterSetCmd(formatter HeaderCellFormatter) tea.Cmd {
	return func() tea.Msg {
		return HeaderCellFormatterSetMsg{Formatter: formatter}
	}
}

// ColumnConstraintsSetCmd returns a command to set column constraints
func ColumnConstraintsSetCmd(columnIndex int, constraints CellConstraint) tea.Cmd {
	return func() tea.Msg {
		return ColumnConstraintsSetMsg{
			ColumnIndex: columnIndex,
			Constraints: constraints,
		}
	}
}

// TableThemeSetCmd returns a command to set the table theme
func TableThemeSetCmd(theme Theme) tea.Cmd {
	return func() tea.Msg {
		return TableThemeSetMsg{Theme: theme}
	}
}

// ================================
// LIST-SPECIFIC COMMANDS
// ================================

// FormatterSetCmd returns a command to set the list item formatter
func FormatterSetCmd(formatter ItemFormatter[any]) tea.Cmd {
	return func() tea.Msg {
		return FormatterSetMsg{Formatter: formatter}
	}
}

// AnimatedFormatterSetCmd returns a command to set the animated list item formatter
func AnimatedFormatterSetCmd(formatter ItemFormatterAnimated[any]) tea.Cmd {
	return func() tea.Msg {
		return AnimatedFormatterSetMsg{Formatter: formatter}
	}
}

// ChunkSizeSetCmd returns a command to set the chunk size
func ChunkSizeSetCmd(size int) tea.Cmd {
	return func() tea.Msg {
		return ChunkSizeSetMsg{Size: size}
	}
}

// MaxWidthSetCmd returns a command to set the maximum width
func MaxWidthSetCmd(width int) tea.Cmd {
	return func() tea.Msg {
		return MaxWidthSetMsg{Width: width}
	}
}

// StyleConfigSetCmd returns a command to set the style configuration
func StyleConfigSetCmd(config StyleConfig) tea.Cmd {
	return func() tea.Msg {
		return StyleConfigSetMsg{Config: config}
	}
}

// ================================
// ANIMATION CONTROL COMMANDS
// ================================

// CellAnimationStartCmd returns a command to start a cell animation
func CellAnimationStartCmd(rowID string, columnIndex int, animation CellAnimation) tea.Cmd {
	return func() tea.Msg {
		return CellAnimationStartMsg{
			RowID:       rowID,
			ColumnIndex: columnIndex,
			Animation:   animation,
		}
	}
}

// CellAnimationStopCmd returns a command to stop a cell animation
func CellAnimationStopCmd(rowID string, columnIndex int) tea.Cmd {
	return func() tea.Msg {
		return CellAnimationStopMsg{
			RowID:       rowID,
			ColumnIndex: columnIndex,
		}
	}
}

// RowAnimationStartCmd returns a command to start a row animation
func RowAnimationStartCmd(rowID string, animation RowAnimation) tea.Cmd {
	return func() tea.Msg {
		return RowAnimationStartMsg{
			RowID:     rowID,
			Animation: animation,
		}
	}
}

// RowAnimationStopCmd returns a command to stop a row animation
func RowAnimationStopCmd(rowID string) tea.Cmd {
	return func() tea.Msg {
		return RowAnimationStopMsg{RowID: rowID}
	}
}

// ItemAnimationStartCmd returns a command to start an item animation
func ItemAnimationStartCmd(itemID string, animation ListAnimation) tea.Cmd {
	return func() tea.Msg {
		return ItemAnimationStartMsg{
			ItemID:    itemID,
			Animation: animation,
		}
	}
}

// ItemAnimationStopCmd returns a command to stop an item animation
func ItemAnimationStopCmd(itemID string) tea.Cmd {
	return func() tea.Msg {
		return ItemAnimationStopMsg{ItemID: itemID}
	}
}

// ================================
// CONFIGURATION COMMANDS
// ================================

// KeyMapSetCmd returns a command to set the key map
func KeyMapSetCmd(keyMap NavigationKeyMap) tea.Cmd {
	return func() tea.Msg {
		return KeyMapSetMsg{KeyMap: keyMap}
	}
}

// PerformanceConfigCmd returns a command to configure performance monitoring
func PerformanceConfigCmd(enabled, monitorMemory, monitorRenderTime bool, reportInterval time.Duration) tea.Cmd {
	return func() tea.Msg {
		return PerformanceConfigMsg{
			Enabled:           enabled,
			MonitorMemory:     monitorMemory,
			MonitorRenderTime: monitorRenderTime,
			ReportInterval:    reportInterval,
		}
	}
}

// DebugEnableCmd returns a command to enable/disable debugging
func DebugEnableCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return DebugEnableMsg{Enabled: enabled}
	}
}

// DebugLevelSetCmd returns a command to set the debug level
func DebugLevelSetCmd(level DebugLevel) tea.Cmd {
	return func() tea.Msg {
		return DebugLevelSetMsg{Level: level}
	}
}

// ================================
// ERROR COMMANDS
// ================================

// ErrorCmd returns a command representing an error
func ErrorCmd(err error, context string) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Error:   err,
			Context: context,
		}
	}
}

// ValidationErrorCmd returns a command representing a validation error
func ValidationErrorCmd(field string, value any, err error, context string) tea.Cmd {
	return func() tea.Msg {
		return ValidationErrorMsg{
			Field:   field,
			Value:   value,
			Error:   err,
			Context: context,
		}
	}
}

// ================================
// STATUS COMMANDS
// ================================

// StatusCmd returns a command with a status message
func StatusCmd(message string, statusType StatusType) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{
			Message: message,
			Type:    statusType,
		}
	}
}

// ================================
// SEARCH COMMANDS
// ================================

// SearchSetCmd returns a command to set a search query
func SearchSetCmd(query, field string) tea.Cmd {
	return func() tea.Msg {
		return SearchSetMsg{
			Query: query,
			Field: field,
		}
	}
}

// SearchClearCmd returns a command to clear the search
func SearchClearCmd() tea.Cmd {
	return func() tea.Msg {
		return SearchClearMsg{}
	}
}

// SearchResultCmd returns a command with search results
func SearchResultCmd(results []int, query string, total int) tea.Cmd {
	return func() tea.Msg {
		return SearchResultMsg{
			Results: results,
			Query:   query,
			Total:   total,
		}
	}
}

// ================================
// ACCESSIBILITY COMMANDS
// ================================

// AccessibilityConfigCmd returns a command to configure accessibility features
func AccessibilityConfigCmd(screenReader, highContrast, reducedMotion bool) tea.Cmd {
	return func() tea.Msg {
		return AccessibilityConfigMsg{
			ScreenReader:  screenReader,
			HighContrast:  highContrast,
			ReducedMotion: reducedMotion,
		}
	}
}

// AriaLabelSetCmd returns a command to set the ARIA label
func AriaLabelSetCmd(label string) tea.Cmd {
	return func() tea.Msg {
		return AriaLabelSetMsg{Label: label}
	}
}

// DescriptionSetCmd returns a command to set the description
func DescriptionSetCmd(description string) tea.Cmd {
	return func() tea.Msg {
		return DescriptionSetMsg{Description: description}
	}
}

// ================================
// BATCH COMMANDS
// ================================

// BatchCmd returns a command that sends multiple messages
func BatchCmd(messages ...interface{}) tea.Cmd {
	return func() tea.Msg {
		return BatchMsg{Messages: messages}
	}
}

// ================================
// LIFECYCLE COMMANDS
// ================================

// InitCmd returns a command to initialize the component
func InitCmd() tea.Cmd {
	return func() tea.Msg {
		return InitMsg{}
	}
}

// DestroyCmd returns a command to destroy the component
func DestroyCmd() tea.Cmd {
	return func() tea.Msg {
		return DestroyMsg{}
	}
}

// ResetCmd returns a command to reset the component
func ResetCmd() tea.Cmd {
	return func() tea.Msg {
		return ResetMsg{}
	}
}

// ================================
// UTILITY COMMANDS
// ================================

// DelayCmd returns a command that sends a message after a delay
func DelayCmd(duration time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return msg
	})
}

// NoOpCmd returns a command that does nothing
func NoOpCmd() tea.Cmd {
	return nil
}

// FullRowHighlightToggleCmd returns a command that toggles full row highlighting mode.
func FullRowHighlightToggleCmd() tea.Cmd {
	return func() tea.Msg {
		return FullRowHighlightToggleMsg{}
	}
}

// FullRowHighlightEnableCmd returns a command that enables full row highlighting mode.
func FullRowHighlightEnableCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return FullRowHighlightEnableMsg{Enabled: enabled}
	}
}

// FullRowHighlightToggleMsg toggles full row highlighting mode.
type FullRowHighlightToggleMsg struct{}

// FullRowHighlightEnableMsg enables or disables full row highlighting mode.
type FullRowHighlightEnableMsg struct {
	Enabled bool
}
