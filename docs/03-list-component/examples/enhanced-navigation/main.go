package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// SimpleDataSource - same as basic example but with more items
type SimpleDataSource struct {
	items []string
}

func NewSimpleDataSource() *SimpleDataSource {
	// Generate 50 items to better appreciate page navigation
	items := make([]string, 50)
	for i := 0; i < 50; i++ {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	return &SimpleDataSource{
		items: items,
	}
}

// All DataSource methods remain exactly the same...
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

// Selection methods (still not used)
func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *SimpleDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *SimpleDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *SimpleDataSource) GetItemID(item any) string                        { return fmt.Sprintf("%v", item) }

// App - same structure
type App struct {
	list *list.List
}

func (app *App) Init() tea.Cmd {
	return app.list.Init()
}

// ENHANCED Update method with more navigation
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit

		// Basic movement (same as before)
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()

		// NEW: Page navigation
		case "pgup", "h":
			return app, core.PageUpCmd()
		case "pgdown", "l":
			return app, core.PageDownCmd()

		// NEW: Jump navigation
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()
		}
	}

	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

// Enhanced View with better help text
func (app *App) View() string {
	return fmt.Sprintf(
		"Enhanced Navigation List (press 'q' to quit)\n\n%s\n\n%s",
		app.list.View(),
		"Navigate: ↑/↓ j/k (line) • PgUp/PgDn h/l (page) • Home/End g/G (jump)",
	)
}

func main() {
	dataSource := NewSimpleDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8 // Taller viewport to better show page navigation

	vtableList := list.NewList(listConfig, dataSource)

	app := &App{list: vtableList}

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
