// Package core provides the fundamental types, interfaces, and messages for the
// vtable library. It defines the shared data structures and contracts used by
// different components like List and Table, ensuring a consistent and
// interoperable architecture. This package is the foundation upon which all other
// vtable modules are built.
package core

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CursorUpCmd creates a command that sends a CursorUpMsg to move the cursor up.
func CursorUpCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorUpMsg{}
	}
}

// CursorDownCmd creates a command that sends a CursorDownMsg to move the cursor down.
func CursorDownCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorDownMsg{}
	}
}

// CursorLeftCmd creates a command that sends a CursorLeftMsg to move the cursor left.
func CursorLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorLeftMsg{}
	}
}

// CursorRightCmd creates a command that sends a CursorRightMsg to move the cursor right.
func CursorRightCmd() tea.Cmd {
	return func() tea.Msg {
		return CursorRightMsg{}
	}
}

// PageUpCmd creates a command that sends a PageUpMsg to move the cursor up one page.
func PageUpCmd() tea.Cmd {
	return func() tea.Msg {
		return PageUpMsg{}
	}
}

// PageDownCmd creates a command that sends a PageDownMsg to move the cursor down one page.
func PageDownCmd() tea.Cmd {
	return func() tea.Msg {
		return PageDownMsg{}
	}
}

// PageLeftCmd creates a command that sends a PageLeftMsg to move the cursor left one page.
func PageLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return PageLeftMsg{}
	}
}

// PageRightCmd creates a command that sends a PageRightMsg to move the cursor right one page.
func PageRightCmd() tea.Cmd {
	return func() tea.Msg {
		return PageRightMsg{}
	}
}

// === HORIZONTAL SCROLLING COMMANDS ===

// HorizontalScrollLeftCmd creates a command that sends a HorizontalScrollLeftMsg to scroll horizontally left within the current column.
func HorizontalScrollLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollLeftMsg{}
	}
}

// HorizontalScrollRightCmd creates a command that sends a HorizontalScrollRightMsg to scroll horizontally right within the current column.
func HorizontalScrollRightCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollRightMsg{}
	}
}

// HorizontalScrollWordLeftCmd creates a command that sends a HorizontalScrollWordLeftMsg to scroll horizontally left by word boundaries.
func HorizontalScrollWordLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollWordLeftMsg{}
	}
}

// HorizontalScrollWordRightCmd creates a command that sends a HorizontalScrollWordRightMsg to scroll horizontally right by word boundaries.
func HorizontalScrollWordRightCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollWordRightMsg{}
	}
}

// HorizontalScrollSmartLeftCmd creates a command that sends a HorizontalScrollSmartLeftMsg to scroll horizontally left using smart boundaries.
func HorizontalScrollSmartLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollSmartLeftMsg{}
	}
}

// HorizontalScrollSmartRightCmd creates a command that sends a HorizontalScrollSmartRightMsg to scroll horizontally right using smart boundaries.
func HorizontalScrollSmartRightCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollSmartRightMsg{}
	}
}

// HorizontalScrollPageLeftCmd creates a command that sends a HorizontalScrollPageLeftMsg to scroll horizontally left by a page amount.
func HorizontalScrollPageLeftCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollPageLeftMsg{}
	}
}

// HorizontalScrollPageRightCmd creates a command that sends a HorizontalScrollPageRightMsg to scroll horizontally right by a page amount.
func HorizontalScrollPageRightCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollPageRightMsg{}
	}
}

// HorizontalScrollModeToggleCmd creates a command that sends a HorizontalScrollModeToggleMsg to cycle through horizontal scroll modes.
func HorizontalScrollModeToggleCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollModeToggleMsg{}
	}
}

// HorizontalScrollScopeToggleCmd creates a command that sends a HorizontalScrollScopeToggleMsg to toggle horizontal scroll scope.
func HorizontalScrollScopeToggleCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollScopeToggleMsg{}
	}
}

// HorizontalScrollResetCmd creates a command that sends a HorizontalScrollResetMsg to reset all horizontal scroll offsets.
func HorizontalScrollResetCmd() tea.Cmd {
	return func() tea.Msg {
		return HorizontalScrollResetMsg{}
	}
}

// === COLUMN NAVIGATION COMMANDS ===

// NextColumnCmd creates a command that sends a NextColumnMsg to move to the next column for horizontal navigation/scrolling focus.
func NextColumnCmd() tea.Cmd {
	return func() tea.Msg {
		return NextColumnMsg{}
	}
}

// PrevColumnCmd creates a command that sends a PrevColumnMsg to move to the previous column for horizontal navigation/scrolling focus.
func PrevColumnCmd() tea.Cmd {
	return func() tea.Msg {
		return PrevColumnMsg{}
	}
}

// JumpToStartCmd creates a command that sends a JumpToStartMsg to move the cursor
// to the first item.
func JumpToStartCmd() tea.Cmd {
	return func() tea.Msg {
		return JumpToStartMsg{}
	}
}

// JumpToEndCmd creates a command that sends a JumpToEndMsg to move the cursor to
// the last item.
func JumpToEndCmd() tea.Cmd {
	return func() tea.Msg {
		return JumpToEndMsg{}
	}
}

// JumpToCmd creates a command that sends a JumpToMsg to move the cursor to a
// specific index.
func JumpToCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return JumpToMsg{Index: index}
	}
}

// TreeJumpToIndexCmd creates a command that sends a TreeJumpToIndexMsg to move
// the cursor to a specific index in a tree, with an option to expand parent nodes.
func TreeJumpToIndexCmd(index int, expandParents bool) tea.Cmd {
	return func() tea.Msg {
		return TreeJumpToIndexMsg{
			Index:         index,
			ExpandParents: expandParents,
		}
	}
}

// DataRefreshCmd creates a command that sends a DataRefreshMsg to trigger a
// full data reload.
func DataRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return DataRefreshMsg{}
	}
}

// DataChunksRefreshCmd creates a command that sends a DataChunksRefreshMsg to
// refresh currently loaded chunks while preserving cursor position.
func DataChunksRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return DataChunksRefreshMsg{}
	}
}

// DataChunkLoadedCmd creates a command that sends a DataChunkLoadedMsg,
// indicating a data chunk has been successfully loaded.
func DataChunkLoadedCmd(startIndex int, items []Data[any], request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return DataChunkLoadedMsg{
			StartIndex: startIndex,
			Items:      items,
			Request:    request,
		}
	}
}

// DataChunkErrorCmd creates a command that sends a DataChunkErrorMsg, indicating
// a data chunk failed to load.
func DataChunkErrorCmd(startIndex int, err error, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return DataChunkErrorMsg{
			StartIndex: startIndex,
			Error:      err,
			Request:    request,
		}
	}
}

// DataTotalCmd creates a command that sends a DataTotalMsg, providing the total
// number of items in the dataset.
func DataTotalCmd(total int) tea.Cmd {
	return func() tea.Msg {
		return DataTotalMsg{Total: total}
	}
}

// DataTotalUpdateCmd creates a command that sends a DataTotalUpdateMsg to
// update the total item count while preserving the cursor position.
func DataTotalUpdateCmd(total int) tea.Cmd {
	return func() tea.Msg {
		return DataTotalUpdateMsg{Total: total}
	}
}

// DataLoadErrorCmd creates a command that sends a DataLoadErrorMsg, indicating a
// general data loading error.
func DataLoadErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return DataLoadErrorMsg{Error: err}
	}
}

// DataSourceSetCmd creates a command that sends a DataSourceSetMsg to replace
// the component's data source.
func DataSourceSetCmd(dataSource DataSource[any]) tea.Cmd {
	return func() tea.Msg {
		return DataSourceSetMsg{DataSource: dataSource}
	}
}

// ChunkUnloadedCmd creates a command that sends a ChunkUnloadedMsg, indicating a
// chunk has been unloaded from memory.
func ChunkUnloadedCmd(chunkStart int) tea.Cmd {
	return func() tea.Msg {
		return ChunkUnloadedMsg{ChunkStart: chunkStart}
	}
}

// ChunkLoadingStartedCmd creates a command that sends a ChunkLoadingStartedMsg,
// indicating a chunk has started to load.
func ChunkLoadingStartedCmd(chunkStart int, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return ChunkLoadingStartedMsg{
			ChunkStart: chunkStart,
			Request:    request,
		}
	}
}

// ChunkLoadingCompletedCmd creates a command that sends a ChunkLoadingCompletedMsg,
// indicating a chunk has finished loading.
func ChunkLoadingCompletedCmd(chunkStart int, itemCount int, request DataRequest) tea.Cmd {
	return func() tea.Msg {
		return ChunkLoadingCompletedMsg{
			ChunkStart: chunkStart,
			ItemCount:  itemCount,
			Request:    request,
		}
	}
}

// DataTotalRequestCmd creates a command that sends a DataTotalRequestMsg to
// explicitly request the total item count from the data source.
func DataTotalRequestCmd() tea.Cmd {
	return func() tea.Msg {
		return DataTotalRequestMsg{}
	}
}

// SelectCurrentCmd creates a command that sends a SelectCurrentMsg to select the
// item currently under the cursor.
func SelectCurrentCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectCurrentMsg{}
	}
}

// SelectToggleCmd creates a command that sends a SelectToggleMsg to toggle the
// selection state of an item at a specific index.
func SelectToggleCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return SelectToggleMsg{Index: index}
	}
}

// SelectAllCmd creates a command that sends a SelectAllMsg to select all items.
func SelectAllCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectAllMsg{}
	}
}

// SelectClearCmd creates a command that sends a SelectClearMsg to clear all selections.
func SelectClearCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectClearMsg{}
	}
}

// SelectRangeCmd creates a command that sends a SelectRangeMsg to select a range
// of items between two item IDs.
func SelectRangeCmd(startID, endID string) tea.Cmd {
	return func() tea.Msg {
		return SelectRangeMsg{
			StartID: startID,
			EndID:   endID,
		}
	}
}

// SelectionModeSetCmd creates a command that sends a SelectionModeSetMsg to change
// the component's selection mode.
func SelectionModeSetCmd(mode SelectionMode) tea.Cmd {
	return func() tea.Msg {
		return SelectionModeSetMsg{Mode: mode}
	}
}

// SelectionResponseCmd creates a command that sends a SelectionResponseMsg,
// typically from a data source, to report the result of a selection operation.
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

// SelectionChangedCmd creates a command that sends a SelectionChangedMsg to
// indicate that the selection state has changed within the data source.
func SelectionChangedCmd(selectedIndices []int, selectedIDs []string, totalSelected int) tea.Cmd {
	return func() tea.Msg {
		return SelectionChangedMsg{
			SelectedIndices: selectedIndices,
			SelectedIDs:     selectedIDs,
			TotalSelected:   totalSelected,
		}
	}
}

// FilterSetCmd creates a command that sends a FilterSetMsg to apply a filter to the data.
func FilterSetCmd(field string, value any) tea.Cmd {
	return func() tea.Msg {
		return FilterSetMsg{
			Field: field,
			Value: value,
		}
	}
}

// FilterClearCmd creates a command that sends a FilterClearMsg to remove a filter.
func FilterClearCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return FilterClearMsg{Field: field}
	}
}

// FiltersClearAllCmd creates a command that sends a FiltersClearAllMsg to remove
// all active filters.
func FiltersClearAllCmd() tea.Cmd {
	return func() tea.Msg {
		return FiltersClearAllMsg{}
	}
}

// SortToggleCmd creates a command that sends a SortToggleMsg to toggle the sort
// order of a field.
func SortToggleCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return SortToggleMsg{Field: field}
	}
}

// SortSetCmd creates a command that sends a SortSetMsg to apply a specific
// sort order to a field.
func SortSetCmd(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortSetMsg{
			Field:     field,
			Direction: direction,
		}
	}
}

// SortAddCmd creates a command that sends a SortAddMsg to add a field to a
// multi-field sort configuration.
func SortAddCmd(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortAddMsg{
			Field:     field,
			Direction: direction,
		}
	}
}

// SortRemoveCmd creates a command that sends a SortRemoveMsg to remove a field
// from the sort configuration.
func SortRemoveCmd(field string) tea.Cmd {
	return func() tea.Msg {
		return SortRemoveMsg{Field: field}
	}
}

// SortsClearAllCmd creates a command that sends a SortsClearAllMsg to clear all
// sorting configurations.
func SortsClearAllCmd() tea.Cmd {
	return func() tea.Msg {
		return SortsClearAllMsg{}
	}
}

// FocusCmd creates a command that sends a FocusMsg to give focus to the component.
func FocusCmd() tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{}
	}
}

// BlurCmd creates a command that sends a BlurMsg to remove focus from the component.
func BlurCmd() tea.Cmd {
	return func() tea.Msg {
		return BlurMsg{}
	}
}

// GlobalAnimationTickCmd creates a command that produces a GlobalAnimationTickMsg
// at a regular interval, driving the animation engine.
func GlobalAnimationTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return GlobalAnimationTickMsg{Timestamp: time.Now()}
	})
}

// AnimationUpdateCmd creates a command that sends an AnimationUpdateMsg to
// indicate that specific animations have updated.
func AnimationUpdateCmd(updatedAnimations []string) tea.Cmd {
	return func() tea.Msg {
		return AnimationUpdateMsg{UpdatedAnimations: updatedAnimations}
	}
}

// AnimationConfigCmd creates a command that sends an AnimationConfigMsg to apply
// a new configuration to the animation engine.
func AnimationConfigCmd(config AnimationConfig) tea.Cmd {
	return func() tea.Msg {
		return AnimationConfigMsg{Config: config}
	}
}

// AnimationStartCmd creates a command that sends an AnimationStartMsg to start
// a specific animation.
func AnimationStartCmd(animationID string) tea.Cmd {
	return func() tea.Msg {
		return AnimationStartMsg{AnimationID: animationID}
	}
}

// AnimationStopCmd creates a command that sends an AnimationStopMsg to stop a
// specific animation.
func AnimationStopCmd(animationID string) tea.Cmd {
	return func() tea.Msg {
		return AnimationStopMsg{AnimationID: animationID}
	}
}

// ThemeSetCmd creates a command that sends a ThemeSetMsg to apply a new theme.
func ThemeSetCmd(theme interface{}) tea.Cmd {
	return func() tea.Msg {
		return ThemeSetMsg{Theme: theme}
	}
}

// RealTimeUpdateCmd creates a command that sends a RealTimeUpdateMsg to trigger
// a real-time data refresh.
func RealTimeUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		return RealTimeUpdateMsg{}
	}
}

// RealTimeConfigCmd creates a command that sends a RealTimeConfigMsg to
// configure real-time updates.
func RealTimeConfigCmd(enabled bool, interval time.Duration) tea.Cmd {
	return func() tea.Msg {
		return RealTimeConfigMsg{
			Enabled:  enabled,
			Interval: interval,
		}
	}
}

// ViewportResizeCmd creates a command that sends a ViewportResizeMsg to notify a
// component that its available size has changed.
func ViewportResizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		return ViewportResizeMsg{
			Width:  width,
			Height: height,
		}
	}
}

// ViewportConfigCmd creates a command that sends a ViewportConfigMsg to apply a
// new viewport configuration.
func ViewportConfigCmd(config ViewportConfig) tea.Cmd {
	return func() tea.Msg {
		return ViewportConfigMsg{Config: config}
	}
}

// ColumnSetCmd creates a command that sends a ColumnSetMsg to define the columns
// for a table component.
func ColumnSetCmd(columns []TableColumn) tea.Cmd {
	return func() tea.Msg {
		return ColumnSetMsg{Columns: columns}
	}
}

// ColumnUpdateCmd creates a command that sends a ColumnUpdateMsg to update the
// configuration of a single table column.
func ColumnUpdateCmd(index int, column TableColumn) tea.Cmd {
	return func() tea.Msg {
		return ColumnUpdateMsg{
			Index:  index,
			Column: column,
		}
	}
}

// HeaderVisibilityCmd creates a command that sends a HeaderVisibilityMsg to set
// the visibility of the table header.
func HeaderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return HeaderVisibilityMsg{Visible: visible}
	}
}

// BorderVisibilityCmd creates a command that sends a BorderVisibilityMsg to set
// the visibility of the table borders.
func BorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return BorderVisibilityMsg{Visible: visible}
	}
}

// TopBorderVisibilityCmd creates a command that sends a TopBorderVisibilityMsg to
// set the visibility of the top border.
func TopBorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return TopBorderVisibilityMsg{Visible: visible}
	}
}

// BottomBorderVisibilityCmd creates a command that sends a BottomBorderVisibilityMsg
// to set the visibility of the bottom border.
func BottomBorderVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return BottomBorderVisibilityMsg{Visible: visible}
	}
}

// HeaderSeparatorVisibilityCmd creates a command that sends a
// HeaderSeparatorVisibilityMsg to set the visibility of the header separator line.
func HeaderSeparatorVisibilityCmd(visible bool) tea.Cmd {
	return func() tea.Msg {
		return HeaderSeparatorVisibilityMsg{Visible: visible}
	}
}

// TopBorderSpaceRemovalCmd creates a command that sends a TopBorderSpaceRemovalMsg
// to control the removal of space for the top border.
func TopBorderSpaceRemovalCmd(remove bool) tea.Cmd {
	return func() tea.Msg {
		return TopBorderSpaceRemovalMsg{Remove: remove}
	}
}

// BottomBorderSpaceRemovalCmd creates a command that sends a BottomBorderSpaceRemovalMsg
// to control the removal of space for the bottom border.
func BottomBorderSpaceRemovalCmd(remove bool) tea.Cmd {
	return func() tea.Msg {
		return BottomBorderSpaceRemovalMsg{Remove: remove}
	}
}

// ActiveCellIndicationModeSetCmd creates a command that sends an
// ActiveCellIndicationModeSetMsg to control the active cell background indicator.
func ActiveCellIndicationModeSetCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return ActiveCellIndicationModeSetMsg{Enabled: enabled}
	}
}

// ActiveCellBackgroundColorSetCmd creates a command that sends an
// ActiveCellBackgroundColorSetMsg to set the active cell's background color.
func ActiveCellBackgroundColorSetCmd(color string) tea.Cmd {
	return func() tea.Msg {
		return ActiveCellBackgroundColorSetMsg{Color: color}
	}
}

// CellFormatterSetCmd creates a command that sends a CellFormatterSetMsg to apply
// a custom formatter to a table column.
func CellFormatterSetCmd(columnIndex int, formatter SimpleCellFormatter) tea.Cmd {
	return func() tea.Msg {
		return CellFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// CellAnimatedFormatterSetCmd creates a command that sends a
// CellAnimatedFormatterSetMsg to apply a custom animated formatter to a table column.
func CellAnimatedFormatterSetCmd(columnIndex int, formatter CellFormatterAnimated) tea.Cmd {
	return func() tea.Msg {
		return CellAnimatedFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// RowFormatterSetCmd creates a command that sends a RowFormatterSetMsg to apply a
// custom formatter for loading placeholder rows.
func RowFormatterSetCmd(formatter LoadingRowFormatter) tea.Cmd {
	return func() tea.Msg {
		return RowFormatterSetMsg{Formatter: formatter}
	}
}

// HeaderFormatterSetCmd creates a command that sends a HeaderFormatterSetMsg to
// apply a custom formatter to a table header cell.
func HeaderFormatterSetCmd(columnIndex int, formatter SimpleHeaderFormatter) tea.Cmd {
	return func() tea.Msg {
		return HeaderFormatterSetMsg{
			ColumnIndex: columnIndex,
			Formatter:   formatter,
		}
	}
}

// LoadingFormatterSetCmd creates a command that sends a LoadingFormatterSetMsg.
//
// Deprecated: Use RowFormatterSetCmd instead.
func LoadingFormatterSetCmd(formatter LoadingRowFormatter) tea.Cmd {
	return func() tea.Msg {
		return LoadingFormatterSetMsg{Formatter: formatter}
	}
}

// HeaderCellFormatterSetCmd creates a command that sends a HeaderCellFormatterSetMsg.
//
// Deprecated: Use HeaderFormatterSetCmd instead.
func HeaderCellFormatterSetCmd(formatter HeaderCellFormatter) tea.Cmd {
	return func() tea.Msg {
		return HeaderCellFormatterSetMsg{Formatter: formatter}
	}
}

// ColumnConstraintsSetCmd creates a command that sends a ColumnConstraintsSetMsg to
// apply layout constraints to a table column.
func ColumnConstraintsSetCmd(columnIndex int, constraints CellConstraint) tea.Cmd {
	return func() tea.Msg {
		return ColumnConstraintsSetMsg{
			ColumnIndex: columnIndex,
			Constraints: constraints,
		}
	}
}

// TableThemeSetCmd creates a command that sends a TableThemeSetMsg to apply a new
// theme to a table.
func TableThemeSetCmd(theme Theme) tea.Cmd {
	return func() tea.Msg {
		return TableThemeSetMsg{Theme: theme}
	}
}

// FormatterSetCmd creates a command that sends a FormatterSetMsg to apply a
// custom item formatter to a list.
func FormatterSetCmd(formatter ItemFormatter[any]) tea.Cmd {
	return func() tea.Msg {
		return FormatterSetMsg{Formatter: formatter}
	}
}

// AnimatedFormatterSetCmd creates a command that sends an AnimatedFormatterSetMsg
// to apply a custom animated item formatter to a list.
func AnimatedFormatterSetCmd(formatter ItemFormatterAnimated[any]) tea.Cmd {
	return func() tea.Msg {
		return AnimatedFormatterSetMsg{Formatter: formatter}
	}
}

// ChunkSizeSetCmd creates a command that sends a ChunkSizeSetMsg to change the
// data loading chunk size.
func ChunkSizeSetCmd(size int) tea.Cmd {
	return func() tea.Msg {
		return ChunkSizeSetMsg{Size: size}
	}
}

// MaxWidthSetCmd creates a command that sends a MaxWidthSetMsg to set the
// maximum width of a list.
func MaxWidthSetCmd(width int) tea.Cmd {
	return func() tea.Msg {
		return MaxWidthSetMsg{Width: width}
	}
}

// StyleConfigSetCmd creates a command that sends a StyleConfigSetMsg to apply a
// new style configuration to a list.
func StyleConfigSetCmd(config StyleConfig) tea.Cmd {
	return func() tea.Msg {
		return StyleConfigSetMsg{Config: config}
	}
}

// CellAnimationStartCmd creates a command that sends a CellAnimationStartMsg to
// start an animation on a specific table cell.
func CellAnimationStartCmd(rowID string, columnIndex int, animation CellAnimation) tea.Cmd {
	return func() tea.Msg {
		return CellAnimationStartMsg{
			RowID:       rowID,
			ColumnIndex: columnIndex,
			Animation:   animation,
		}
	}
}

// CellAnimationStopCmd creates a command that sends a CellAnimationStopMsg to
// stop an animation on a specific table cell.
func CellAnimationStopCmd(rowID string, columnIndex int) tea.Cmd {
	return func() tea.Msg {
		return CellAnimationStopMsg{
			RowID:       rowID,
			ColumnIndex: columnIndex,
		}
	}
}

// RowAnimationStartCmd creates a command that sends a RowAnimationStartMsg to
// start an animation on a specific table row.
func RowAnimationStartCmd(rowID string, animation RowAnimation) tea.Cmd {
	return func() tea.Msg {
		return RowAnimationStartMsg{
			RowID:     rowID,
			Animation: animation,
		}
	}
}

// RowAnimationStopCmd creates a command that sends a RowAnimationStopMsg to stop
// an animation on a specific table row.
func RowAnimationStopCmd(rowID string) tea.Cmd {
	return func() tea.Msg {
		return RowAnimationStopMsg{RowID: rowID}
	}
}

// ItemAnimationStartCmd creates a command that sends an ItemAnimationStartMsg to
// start an animation on a specific list item.
func ItemAnimationStartCmd(itemID string, animation ListAnimation) tea.Cmd {
	return func() tea.Msg {
		return ItemAnimationStartMsg{
			ItemID:    itemID,
			Animation: animation,
		}
	}
}

// ItemAnimationStopCmd creates a command that sends an ItemAnimationStopMsg to
// stop an animation on a specific list item.
func ItemAnimationStopCmd(itemID string) tea.Cmd {
	return func() tea.Msg {
		return ItemAnimationStopMsg{ItemID: itemID}
	}
}

// KeyMapSetCmd creates a command that sends a KeyMapSetMsg to apply a new key map.
func KeyMapSetCmd(keyMap NavigationKeyMap) tea.Cmd {
	return func() tea.Msg {
		return KeyMapSetMsg{KeyMap: keyMap}
	}
}

// PerformanceConfigCmd creates a command that sends a PerformanceConfigMsg to
// configure performance monitoring.
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

// DebugEnableCmd creates a command that sends a DebugEnableMsg to enable or
// disable debugging features.
func DebugEnableCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return DebugEnableMsg{Enabled: enabled}
	}
}

// DebugLevelSetCmd creates a command that sends a DebugLevelSetMsg to set the
// debug verbosity level.
func DebugLevelSetCmd(level DebugLevel) tea.Cmd {
	return func() tea.Msg {
		return DebugLevelSetMsg{Level: level}
	}
}

// ErrorCmd creates a command that sends an ErrorMsg to report a generic error.
func ErrorCmd(err error, context string) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Error:   err,
			Context: context,
		}
	}
}

// ValidationErrorCmd creates a command that sends a ValidationErrorMsg to report
// a validation error.
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

// StatusCmd creates a command that sends a StatusMsg to display a status message
// to the user.
func StatusCmd(message string, statusType StatusType) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{
			Message: message,
			Type:    statusType,
		}
	}
}

// SearchSetCmd creates a command that sends a SearchSetMsg to initiate a search.
func SearchSetCmd(query, field string) tea.Cmd {
	return func() tea.Msg {
		return SearchSetMsg{
			Query: query,
			Field: field,
		}
	}
}

// SearchClearCmd creates a command that sends a SearchClearMsg to clear the
// current search.
func SearchClearCmd() tea.Cmd {
	return func() tea.Msg {
		return SearchClearMsg{}
	}
}

// SearchResultCmd creates a command that sends a SearchResultMsg containing the
// results of a search.
func SearchResultCmd(results []int, query string, total int) tea.Cmd {
	return func() tea.Msg {
		return SearchResultMsg{
			Results: results,
			Query:   query,
			Total:   total,
		}
	}
}

// AccessibilityConfigCmd creates a command that sends an AccessibilityConfigMsg
// to configure accessibility features.
func AccessibilityConfigCmd(screenReader, highContrast, reducedMotion bool) tea.Cmd {
	return func() tea.Msg {
		return AccessibilityConfigMsg{
			ScreenReader:  screenReader,
			HighContrast:  highContrast,
			ReducedMotion: reducedMotion,
		}
	}
}

// AriaLabelSetCmd creates a command that sends an AriaLabelSetMsg to set the
// ARIA label for a component.
func AriaLabelSetCmd(label string) tea.Cmd {
	return func() tea.Msg {
		return AriaLabelSetMsg{Label: label}
	}
}

// DescriptionSetCmd creates a command that sends a DescriptionSetMsg to set the
// accessible description for a component.
func DescriptionSetCmd(description string) tea.Cmd {
	return func() tea.Msg {
		return DescriptionSetMsg{Description: description}
	}
}

// BatchCmd creates a command that sends a BatchMsg, allowing multiple messages
// to be dispatched at once.
func BatchCmd(messages ...interface{}) tea.Cmd {
	return func() tea.Msg {
		return BatchMsg{Messages: messages}
	}
}

// InitCmd creates a command that sends an InitMsg to initialize a component.
func InitCmd() tea.Cmd {
	return func() tea.Msg {
		return InitMsg{}
	}
}

// DestroyCmd creates a command that sends a DestroyMsg to clean up a component.
func DestroyCmd() tea.Cmd {
	return func() tea.Msg {
		return DestroyMsg{}
	}
}

// ResetCmd creates a command that sends a ResetMsg to reset a component to its
// initial state.
func ResetCmd() tea.Cmd {
	return func() tea.Msg {
		return ResetMsg{}
	}
}

// DelayCmd creates a command that sends a given message after a specified duration.
func DelayCmd(duration time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return msg
	})
}

// NoOpCmd returns a command that does nothing. It is useful for satisfying
// function signatures that require a command to be returned, but no action is needed.
func NoOpCmd() tea.Cmd {
	return nil
}

// FullRowHighlightToggleCmd creates a command that sends a FullRowHighlightToggleMsg
// to toggle full row highlighting mode.
func FullRowHighlightToggleCmd() tea.Cmd {
	return func() tea.Msg {
		return FullRowHighlightToggleMsg{}
	}
}

// FullRowHighlightEnableCmd creates a command that sends a FullRowHighlightEnableMsg
// to enable or disable full row highlighting mode.
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

// AriaLabelCmd returns a command to set the ARIA label for accessibility.
func AriaLabelCmd(label string) tea.Cmd {
	return func() tea.Msg {
		return AriaLabelSetMsg{Label: label}
	}
}

// SetFullRowSelectionCmd returns a command to enable/disable full row selection styling
func SetFullRowSelectionCmd(enabled bool, background lipgloss.Style) tea.Cmd {
	return func() tea.Msg {
		return SetFullRowSelectionMsg{
			Enabled:    enabled,
			Background: background,
		}
	}
}

// SetCursorRowStylingCmd returns a command to enable/disable full row cursor styling
func SetCursorRowStylingCmd(enabled bool, background lipgloss.Style) tea.Cmd {
	return func() tea.Msg {
		return SetCursorRowStylingMsg{
			Enabled:    enabled,
			Background: background,
		}
	}
}

// SetComponentBackgroundCmd returns a command to configure background styling for a specific component
func SetComponentBackgroundCmd(componentType ListComponentType, cursorBg, selectedBg, normalBg lipgloss.Style, applyCursor, applySelected, applyNormal bool) tea.Cmd {
	return func() tea.Msg {
		return SetComponentBackgroundMsg{
			ComponentType: componentType,
			CursorBg:      cursorBg,
			SelectedBg:    selectedBg,
			NormalBg:      normalBg,
			ApplyCursor:   applyCursor,
			ApplySelected: applySelected,
			ApplyNormal:   applyNormal,
		}
	}
}
