package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// SimpleDataSource - same as basic selection example
type SimpleDataSource struct {
	items    []string
	selected map[int]bool
}

func NewSimpleDataSource() *SimpleDataSource {
	items := make([]string, 50)
	for i := 0; i < 50; i++ {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	return &SimpleDataSource{
		items:    items,
		selected: make(map[int]bool),
	}
}

// GetSelectedCount returns the number of selected items
func (ds *SimpleDataSource) GetSelectedCount() int {
	return len(ds.selected)
}

func (ds *SimpleDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.items)}
	}
}

func (ds *SimpleDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *SimpleDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		for i := request.Start; i < request.Start+request.Count && i < len(ds.items); i++ {
			items = append(items, core.Data[any]{
				ID:       fmt.Sprintf("item-%d", i),
				Item:     ds.items[i],
				Selected: ds.selected[i],
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.items) {
			if selected {
				ds.selected[index] = true
			} else {
				delete(ds.selected, index)
			}
			return core.SelectionResponseMsg{
				Success:  true,
				Index:    index,
				ID:       fmt.Sprintf("item-%d", index),
				Selected: selected,
			}
		}
		return core.SelectionResponseMsg{Success: false}
	}
}

func (ds *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	var index int
	if _, err := fmt.Sscanf(id, "item-%d", &index); err == nil {
		return ds.SetSelected(index, selected)
	}
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: false}
	}
}

func (ds *SimpleDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < len(ds.items); i++ {
			ds.selected[i] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *SimpleDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		ds.selected = make(map[int]bool)
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		if startIndex > endIndex {
			startIndex, endIndex = endIndex, startIndex
		}
		for i := startIndex; i <= endIndex && i < len(ds.items); i++ {
			ds.selected[i] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *SimpleDataSource) GetItemID(item any) string {
	return fmt.Sprintf("%v", item)
}

// App - enhanced with selection tracking
type App struct {
	list           *list.List
	dataSource     *SimpleDataSource // NEW: Keep reference to DataSource
	selectionCount int               // NEW: Track selection count
	statusMessage  string            // NEW: Show status
}

func (app *App) Init() tea.Cmd {
	return app.list.Init()
}

// ENHANCED Update method with multiple selection
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit

		// Navigation (same as before)
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
		case "pgup", "h":
			return app, core.PageUpCmd()
		case "pgdown", "l":
			return app, core.PageDownCmd()
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()

		// Basic selection (same as before)
		case " ": // Spacebar
			return app, core.SelectCurrentCmd()

		// NEW: Multiple selection
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "ctrl+d":
			return app, core.SelectClearCmd()
		}

	// NEW: Handle selection response messages
	case core.SelectionResponseMsg:
		if msg.Success {
			app.updateSelectionCount()
			app.updateStatusMessage()
		}
	}

	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

// NEW: Update selection count from DataSource
func (app *App) updateSelectionCount() {
	app.selectionCount = app.dataSource.GetSelectedCount()
}

// NEW: Update status message
func (app *App) updateStatusMessage() {
	if app.selectionCount == 0 {
		app.statusMessage = "No items selected"
	} else if app.selectionCount == 1 {
		app.statusMessage = "1 item selected"
	} else {
		app.statusMessage = fmt.Sprintf("%d items selected", app.selectionCount)
	}
}

// Enhanced View with status
func (app *App) View() string {
	return fmt.Sprintf(
		"Multiple Selection List (press 'q' to quit)\n\n%s\n\n%s\n%s",
		app.list.View(),
		"Navigate: ↑/↓ j/k • Page: h/l • Jump: g/G • Select: Space • Multi: Ctrl+A/D",
		app.statusMessage,
	)
}

// `03-list-component/examples/multiple-selection/main.go`
func main() {
	dataSource := NewSimpleDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.SelectionMode = core.SelectionMultiple

	vtableList := list.NewList(listConfig, dataSource)

	app := &App{
		list:           vtableList,
		dataSource:     dataSource,          // NEW: Keep reference
		selectionCount: 0,                   // NEW: Initialize
		statusMessage:  "No items selected", // NEW: Initialize
	}

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
