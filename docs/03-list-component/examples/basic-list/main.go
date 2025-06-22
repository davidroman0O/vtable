package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// SimpleDataSource provides our basic data
type SimpleDataSource struct {
	items []string
}

func NewSimpleDataSource() *SimpleDataSource {
	// Generate 20 simple items
	items := make([]string, 20)
	for i := 0; i < 20; i++ {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	return &SimpleDataSource{
		items: items,
	}
}

// GetTotal returns total number of items
func (ds *SimpleDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.items)}
	}
}

// RefreshTotal reloads the total (same as GetTotal for this simple case)
func (ds *SimpleDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a specific range of items for the viewport
func (ds *SimpleDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		// Load the requested range
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

// Selection methods (required but not used in basic example)
func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *SimpleDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *SimpleDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *SimpleDataSource) GetItemID(item any) string                        { return fmt.Sprintf("%v", item) }

// App holds our application state
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
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
		}
	}

	// Pass all messages to the list
	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

func (app *App) View() string {
	return fmt.Sprintf(
		"Basic VTable List (press 'q' to quit)\n\n%s\n\nUse ↑/↓ or j/k to navigate",
		app.list.View(),
	)
}

// `03-list-component/examples/basic-list/main.go`
func main() {
	// Create data source
	dataSource := NewSimpleDataSource()

	// Create list configuration
	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 5 // Show 5 items at a time

	// Create the list
	vtableList := list.NewList(listConfig, dataSource)

	// Create and run the app
	app := &App{list: vtableList}

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
