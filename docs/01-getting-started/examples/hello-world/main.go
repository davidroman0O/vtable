package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// SimpleDataSource provides basic string data
type SimpleDataSource struct {
	items []string
}

// GetTotal returns the total number of items
func (ds *SimpleDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.items)}
	}
}

// RefreshTotal reloads the total count
func (ds *SimpleDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of data for the viewport
func (ds *SimpleDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		// Load requested items
		for i := request.Start; i < request.Start+request.Count && i < len(ds.items); i++ {
			items = append(items, core.Data[any]{
				ID:   fmt.Sprintf("item-%d", i),
				Item: ds.items[i],
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

// Selection methods (required by interface, but not used in this example)
func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *SimpleDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *SimpleDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *SimpleDataSource) GetItemID(item any) string                        { return fmt.Sprintf("%v", item) }

// App is your main Bubble Tea model
type App struct {
	list *list.List
}

func (app *App) Init() tea.Cmd {
	return app.list.Init()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		case "up", "k":
			// Move cursor up
			return app, core.CursorUpCmd()
		case "down", "j":
			// Move cursor down
			return app, core.CursorDownCmd()
		}
	}

	// Pass all other messages to the list
	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

func (app *App) View() string {
	return fmt.Sprintf(
		"Hello World VTable List (press 'q' to quit)\n\n%s\n\nUse ↑/↓ or j/k to navigate",
		app.list.View(),
	)
}

func main() {
	// Create sample data
	dataSource := &SimpleDataSource{
		items: []string{
			"Item 1",
			"Item 2",
			"Item 3",
			"Item 4",
			"Item 5",
			"Item 6",
			"Item 7",
			"Item 8",
			"Item 9",
			"Item 10",
		},
	}

	// Create list configuration
	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 5 // Show 5 items at a time

	// Create the list
	vtableList := list.NewList(listConfig, dataSource)

	// Create the app
	app := &App{list: vtableList}

	// Run the app
	program := tea.NewProgram(app)
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
