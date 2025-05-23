package main

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// Test data provider that exactly matches the example
type TestStringProvider struct {
	items     []string
	selection map[int]bool
}

func NewTestStringProvider(count int) *TestStringProvider {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	return &TestStringProvider{
		items:     items,
		selection: make(map[int]bool),
	}
}

func (p *TestStringProvider) GetTotal() int {
	return len(p.items)
}

func (p *TestStringProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []vtable.Data[string]{}, nil
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]vtable.Data[string], end-start)
	for i := start; i < end; i++ {
		result[i-start] = vtable.Data[string]{
			Item:     p.items[i],
			Selected: p.selection[i],
			Metadata: nil,
			Disabled: false,
			Hidden:   false,
		}
	}

	return result, nil
}

func (p *TestStringProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *TestStringProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *TestStringProvider) SelectAll() bool {
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *TestStringProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *TestStringProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *TestStringProvider) GetItemID(item *string) string {
	for i, itm := range p.items {
		if itm == *item {
			return strconv.Itoa(i)
		}
	}
	return ""
}

func (p *TestStringProvider) FindItemIndex(key string, value any) (int, bool) {
	if key != "id" {
		return -1, false
	}

	var id int
	switch v := value.(type) {
	case int:
		id = v
	case string:
		var err error
		id, err = strconv.Atoi(v)
		if err != nil {
			return -1, false
		}
	default:
		return -1, false
	}

	if id >= 0 && id < len(p.items) {
		return id, true
	}

	return -1, false
}

type TestModel struct {
	list   *vtable.TeaList[string]
	status string
}

func (m TestModel) Init() tea.Cmd {
	return nil
}

func (m TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle selection keys FIRST, before component processes them
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case " ", "space":
			if m.list.ToggleCurrentSelection() {
				selectionCount := m.list.GetSelectionCount()
				m.status = fmt.Sprintf("Toggled selection - count: %d", selectionCount)
			} else {
				m.status = "Toggle failed"
			}
			// Return early to prevent component processing
			return m, tea.Batch(cmds...)
		case "ctrl+a":
			m.list.SelectAll()
			count := m.list.GetSelectionCount()
			m.status = fmt.Sprintf("Selected all - count: %d", count)
			// Return early to prevent component processing
			return m, tea.Batch(cmds...)
		case "esc", "escape":
			m.list.ClearSelection()
			count := m.list.GetSelectionCount()
			m.status = fmt.Sprintf("Cleared selection - count: %d", count)
			// Return early to prevent component processing
			return m, tea.Batch(cmds...)
		}
	}

	// Update the list only if we didn't handle a selection key
	listModel, cmd := m.list.Update(msg)
	if listM, ok := listModel.(*vtable.TeaList[string]); ok {
		m.list = listM
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m TestModel) View() string {
	return m.list.View() + "\n\nStatus: " + m.status + "\n\nPress SPACE to toggle, CTRL+A to select all, ESC to clear, Q to quit"
}

func main() {
	// Create provider
	provider := NewTestStringProvider(20)

	// Create viewport config
	config := vtable.ViewportConfig{
		Height:               10,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 7,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create style config
	styleConfig := vtable.ThemeToStyleConfig(vtable.DefaultTheme())

	// Create the PROPER formatter
	listFormatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item

		var style lipgloss.Style
		if isCursor {
			if data.Selected {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("226")).Bold(true)
			} else {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("63"))
			}
		} else if data.Selected {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("46"))
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		result := fmt.Sprintf("%d: %s", index, item)

		var prefix string
		if isCursor && data.Selected {
			prefix = "✓>"
		} else if isCursor {
			prefix = "> "
		} else if data.Selected {
			prefix = "✓ "
		} else {
			prefix = "  "
		}

		result = fmt.Sprintf("%s %s", prefix, result)
		return style.Render(result)
	}

	// Create list with the ACTUAL formatter
	list, err := vtable.NewTeaList(config, provider, styleConfig, listFormatter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	model := TestModel{
		list:   list,
		status: "Ready",
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
