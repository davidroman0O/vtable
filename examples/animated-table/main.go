package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// ServerMetrics represents server monitoring data
type ServerMetrics struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	Memory   float64 `json:"memory"`
	Network  float64 `json:"network"`
	Uptime   int     `json:"uptime"` // hours
	LastSeen time.Time
}

// ServerDataProvider provides server metrics with live updates
type ServerDataProvider struct {
	servers         []ServerMetrics
	selection       map[int]bool
	lastUpdate      time.Time
	serverIDCounter int
}

func NewServerDataProvider() *ServerDataProvider {
	return &ServerDataProvider{
		servers:         generateServers(),
		selection:       make(map[int]bool),
		lastUpdate:      time.Now(),
		serverIDCounter: 10,
	}
}

func generateServers() []ServerMetrics {
	servers := []ServerMetrics{
		{ID: 1, Name: "web-01", Status: "healthy", CPU: 45.2, Memory: 62.1, Network: 1.5, Uptime: 720, LastSeen: time.Now()},
		{ID: 2, Name: "web-02", Status: "healthy", CPU: 38.7, Memory: 58.3, Network: 1.2, Uptime: 720, LastSeen: time.Now()},
		{ID: 3, Name: "api-01", Status: "warning", CPU: 78.9, Memory: 85.4, Network: 2.8, Uptime: 168, LastSeen: time.Now()},
		{ID: 4, Name: "api-02", Status: "healthy", CPU: 52.3, Memory: 71.2, Network: 2.1, Uptime: 168, LastSeen: time.Now()},
		{ID: 5, Name: "db-01", Status: "critical", CPU: 91.5, Memory: 94.7, Network: 5.2, Uptime: 72, LastSeen: time.Now()},
		{ID: 6, Name: "db-02", Status: "healthy", CPU: 42.1, Memory: 67.8, Network: 3.9, Uptime: 720, LastSeen: time.Now()},
		{ID: 7, Name: "cache-01", Status: "healthy", CPU: 23.4, Memory: 41.2, Network: 0.8, Uptime: 360, LastSeen: time.Now()},
		{ID: 8, Name: "queue-01", Status: "warning", CPU: 67.8, Memory: 78.5, Network: 1.9, Uptime: 240, LastSeen: time.Now()},
		{ID: 9, Name: "lb-01", Status: "healthy", CPU: 28.9, Memory: 45.6, Network: 12.3, Uptime: 720, LastSeen: time.Now()},
		{ID: 10, Name: "monitor-01", Status: "healthy", CPU: 15.2, Memory: 32.1, Network: 0.4, Uptime: 720, LastSeen: time.Now()},
	}

	return servers
}

func (p *ServerDataProvider) GetTotal() int {
	return len(p.servers)
}

func (p *ServerDataProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	// Apply filters
	filteredServers := make([]ServerMetrics, 0, len(p.servers))
	for _, server := range p.servers {
		// Status filter
		if statusFilter, exists := request.Filters["status"]; exists {
			if statusStr, ok := statusFilter.(string); ok && statusStr != server.Status {
				continue
			}
		}

		// High CPU filter (> 80%)
		if cpuFilter, exists := request.Filters["high_cpu"]; exists {
			if cpuBool, ok := cpuFilter.(bool); ok && cpuBool && server.CPU <= 80.0 {
				continue
			}
		}

		filteredServers = append(filteredServers, server)
	}

	// Apply sorting
	if len(request.SortFields) > 0 {
		// Simple sorting by first field
		sortField := request.SortFields[0]
		sortDirection := request.SortDirections[0]

		// Sort the filtered servers based on the field
		for i := 0; i < len(filteredServers)-1; i++ {
			for j := i + 1; j < len(filteredServers); j++ {
				shouldSwap := false

				switch sortField {
				case "name":
					if sortDirection == "asc" {
						shouldSwap = filteredServers[i].Name > filteredServers[j].Name
					} else {
						shouldSwap = filteredServers[i].Name < filteredServers[j].Name
					}
				case "cpu":
					if sortDirection == "asc" {
						shouldSwap = filteredServers[i].CPU > filteredServers[j].CPU
					} else {
						shouldSwap = filteredServers[i].CPU < filteredServers[j].CPU
					}
				case "memory":
					if sortDirection == "asc" {
						shouldSwap = filteredServers[i].Memory > filteredServers[j].Memory
					} else {
						shouldSwap = filteredServers[i].Memory < filteredServers[j].Memory
					}
				case "status":
					if sortDirection == "asc" {
						shouldSwap = filteredServers[i].Status > filteredServers[j].Status
					} else {
						shouldSwap = filteredServers[i].Status < filteredServers[j].Status
					}
				}

				if shouldSwap {
					filteredServers[i], filteredServers[j] = filteredServers[j], filteredServers[i]
				}
			}
		}
	}

	if start >= len(filteredServers) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	end := start + count
	if end > len(filteredServers) {
		end = len(filteredServers)
	}

	result := make([]vtable.Data[vtable.TableRow], end-start)
	for i := start; i < end; i++ {
		server := filteredServers[i]

		// Find original index for selection state
		originalIndex := -1
		for j, original := range p.servers {
			if original.ID == server.ID {
				originalIndex = j
				break
			}
		}

		// Convert server metrics to table row
		tableRow := vtable.TableRow{
			Cells: []string{
				server.Name,
				server.Status,
				fmt.Sprintf("%.1f%%", server.CPU),
				fmt.Sprintf("%.1f%%", server.Memory),
				fmt.Sprintf("%.1f MB/s", server.Network),
				fmt.Sprintf("%dh", server.Uptime),
			},
		}

		result[i-start] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("server-%d", server.ID),
			Item:     tableRow,
			Selected: originalIndex >= 0 && p.selection[originalIndex],
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement DataProvider interface
func (p *ServerDataProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *ServerDataProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.servers) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *ServerDataProvider) SelectAll() bool {
	for i := 0; i < len(p.servers); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *ServerDataProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *ServerDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *ServerDataProvider) GetItemID(item *vtable.TableRow) string {
	// Extract server name from first cell
	if len(item.Cells) > 0 {
		return "server-" + item.Cells[0]
	}
	return "unknown"
}

func (p *ServerDataProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.servers) {
			ids = append(ids, fmt.Sprintf("server-%d", p.servers[idx].ID))
		}
	}
	return ids
}

func (p *ServerDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		for i, server := range p.servers {
			if fmt.Sprintf("server-%d", server.ID) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *ServerDataProvider) SelectRange(startID, endID string) bool {
	return true
}

// UpdateMetrics simulates live server metrics updates
func (p *ServerDataProvider) UpdateMetrics() {
	now := time.Now()
	elapsed := now.Sub(p.lastUpdate)

	// Update every 300ms for smooth animations
	if elapsed < 300*time.Millisecond {
		return
	}

	for i := range p.servers {
		server := &p.servers[i]
		server.LastSeen = now

		// Simulate metric fluctuations
		p.updateServerMetrics(server)

		// Update uptime
		server.Uptime += int(elapsed.Hours())
	}

	p.lastUpdate = now
}

func (p *ServerDataProvider) updateServerMetrics(server *ServerMetrics) {
	// CPU fluctuation based on status
	switch server.Status {
	case "healthy":
		// Small fluctuations for healthy servers
		server.CPU += (rand.Float64() - 0.5) * 10
		if server.CPU < 10 {
			server.CPU = 10 + rand.Float64()*20
		}
		if server.CPU > 70 {
			server.CPU = 50 + rand.Float64()*15
		}

	case "warning":
		// Higher CPU with more variation
		server.CPU += (rand.Float64() - 0.5) * 15
		if server.CPU < 60 {
			server.CPU = 60 + rand.Float64()*20
		}
		if server.CPU > 90 {
			server.CPU = 85 + rand.Float64()*5
		}

	case "critical":
		// Very high CPU with spikes
		server.CPU += (rand.Float64() - 0.5) * 8
		if server.CPU < 85 {
			server.CPU = 85 + rand.Float64()*10
		}
		if server.CPU > 99 {
			server.CPU = 95 + rand.Float64()*4
		}
	}

	// Memory follows CPU somewhat
	targetMemory := server.CPU + (rand.Float64()-0.5)*20
	server.Memory += (targetMemory - server.Memory) * 0.1
	if server.Memory < 20 {
		server.Memory = 20
	}
	if server.Memory > 98 {
		server.Memory = 98
	}

	// Network traffic varies
	server.Network += (rand.Float64() - 0.5) * 2
	if server.Network < 0.1 {
		server.Network = 0.1
	}
	if server.Network > 15 {
		server.Network = 15
	}

	// Status updates based on metrics
	p.updateServerStatus(server)
}

func (p *ServerDataProvider) updateServerStatus(server *ServerMetrics) {
	if server.CPU > 90 || server.Memory > 90 {
		server.Status = "critical"
	} else if server.CPU > 75 || server.Memory > 80 {
		server.Status = "warning"
	} else {
		server.Status = "healthy"
	}
}

// RestartServers simulates restarting selected servers
func (p *ServerDataProvider) RestartServers(indices []int) int {
	restarted := 0
	for _, index := range indices {
		if index >= 0 && index < len(p.servers) {
			server := &p.servers[index]
			server.CPU = 15 + rand.Float64()*25 // Fresh restart metrics
			server.Memory = 25 + rand.Float64()*30
			server.Network = 0.5 + rand.Float64()*2
			server.Status = "healthy"
			server.Uptime = 0
			server.LastSeen = time.Now()
			restarted++
		}
	}
	return restarted
}

// GetServer returns the actual server by index for additional operations
func (p *ServerDataProvider) GetServer(index int) (*ServerMetrics, bool) {
	if index >= 0 && index < len(p.servers) {
		return &p.servers[index], true
	}
	return nil, false
}

type animatedTableModel struct {
	table             *vtable.TeaTable
	provider          *ServerDataProvider
	quitting          bool
	lastUpdate        time.Time
	tickCount         int
	activeFilter      string
	activeSort        string
	statusMessage     string
	animationsEnabled bool
	showAnimated      bool
}

func newAnimatedTableModel() *animatedTableModel {
	// Create data provider
	provider := NewServerDataProvider()

	// Configure table columns using convenience functions
	columns := []vtable.TableColumn{
		vtable.NewColumn("Server", 12),
		vtable.NewCenterColumn("Status", 10),
		vtable.NewRightColumn("CPU", 8),
		vtable.NewRightColumn("Memory", 8),
		vtable.NewRightColumn("Network", 12),
		vtable.NewRightColumn("Uptime", 8),
	}

	// Create table with dark theme and convenience function
	table, err := vtable.NewTeaTableWithTheme(columns, provider, vtable.DarkTheme())
	if err != nil {
		log.Fatal(err)
	}

	// Configure animation tick rate
	table.SetTickInterval(200 * time.Millisecond)

	// Create animated formatter for table rows
	animatedFormatter := func(data vtable.Data[vtable.TableRow], index int, ctx vtable.RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) vtable.RenderResult {

		row := data.Item
		if len(row.Cells) < 6 {
			// Fallback for malformed data
			return vtable.RenderResult{
				Content: "Invalid row data",
			}
		}

		// Get animation counter
		counter := 0
		if c, ok := animationState["counter"]; ok {
			if ci, ok := c.(int); ok {
				counter = ci
			}
		}
		counter++

		// Parse metrics for dynamic styling
		cpu, _ := strconv.ParseFloat(strings.TrimSuffix(row.Cells[2], "%"), 64)
		memory, _ := strconv.ParseFloat(strings.TrimSuffix(row.Cells[3], "%"), 64)
		status := row.Cells[1]

		// Build animated row with status-based effects
		var content strings.Builder

		// Selection and cursor indicators
		if data.Selected && isCursor {
			content.WriteString("✓>")
		} else if data.Selected {
			content.WriteString("✓ ")
		} else if isCursor {
			content.WriteString("> ")
		} else {
			content.WriteString("  ")
		}

		// Server name with status color
		serverName := row.Cells[0]
		nameStyle := getServerNameStyle(status)
		content.WriteString(nameStyle.Render(fmt.Sprintf("%-10s", serverName)))
		content.WriteString(" │ ")

		// Animated status with indicators
		statusContent := getAnimatedStatus(status, counter)
		content.WriteString(fmt.Sprintf("%-10s", statusContent))
		content.WriteString(" │ ")

		// CPU with color coding and animated bars
		cpuContent := getAnimatedMetric(cpu, row.Cells[2], "cpu", counter)
		content.WriteString(fmt.Sprintf("%8s", cpuContent))
		content.WriteString(" │ ")

		// Memory with color coding
		memContent := getAnimatedMetric(memory, row.Cells[3], "memory", counter)
		content.WriteString(fmt.Sprintf("%8s", memContent))
		content.WriteString(" │ ")

		// Network with live graph
		networkContent := getAnimatedNetwork(row.Cells[4], counter)
		content.WriteString(fmt.Sprintf("%12s", networkContent))
		content.WriteString(" │ ")

		// Uptime with live timer
		uptimeContent := getAnimatedUptime(row.Cells[5], ctx.CurrentTime)
		content.WriteString(fmt.Sprintf("%8s", uptimeContent))

		return vtable.RenderResult{
			Content: content.String(),
			RefreshTriggers: []vtable.RefreshTrigger{{
				Type:     vtable.TriggerTimer,
				Interval: 200 * time.Millisecond,
			}},
			AnimationState: map[string]any{
				"counter":   counter,
				"timestamp": ctx.CurrentTime.Unix(),
			},
		}
	}

	// Start with animated formatter
	table.SetAnimatedFormatter(animatedFormatter)

	// Enable real-time updates
	table.EnableRealTimeUpdates(1 * time.Second)

	model := &animatedTableModel{
		table:             table,
		provider:          provider,
		lastUpdate:        time.Now(),
		activeFilter:      "all",
		activeSort:        "none",
		animationsEnabled: true,
		showAnimated:      true,
	}

	return model
}

func getServerNameStyle(status string) lipgloss.Style {
	switch status {
	case "critical":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	case "warning":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	case "healthy":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Gray
	}
}

func getAnimatedStatus(status string, counter int) string {
	switch status {
	case "critical":
		// Blinking critical status
		if counter%4 < 2 {
			return "🔴 CRIT"
		}
		return "🚨 CRIT"
	case "warning":
		// Pulsing warning
		symbols := []string{"⚠️ ", "🟡 ", "⚠️ ", "🟡 "}
		return symbols[counter%4] + "WARN"
	case "healthy":
		// Steady healthy
		return "✅ OK"
	default:
		return "❓ UNK"
	}
}

func getAnimatedMetric(value float64, original string, metricType string, counter int) string {
	var color lipgloss.Color

	// Color based on value
	if value >= 90 {
		color = lipgloss.Color("9") // Red
	} else if value >= 75 {
		color = lipgloss.Color("11") // Yellow
	} else if value >= 50 {
		color = lipgloss.Color("3") // Orange
	} else {
		color = lipgloss.Color("10") // Green
	}

	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(original)
}

func getAnimatedNetwork(original string, counter int) string {
	// Parse network value
	val := strings.TrimSuffix(original, " MB/s")
	network, _ := strconv.ParseFloat(val, 64)

	// Create mini network activity graph
	var graph string
	baseLevel := int(network * 2) // Scale for visualization

	// Simulate network activity spikes
	spike := 0
	if counter%8 == 0 || counter%13 == 0 {
		spike = 2
	}

	level := baseLevel + spike
	if level > 8 {
		level = 8
	}

	// Build vertical bar graph
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	if level > 0 && level <= len(bars) {
		graph = bars[level-1]
	} else {
		graph = "▁"
	}

	return fmt.Sprintf("%s %s", graph, original)
}

func getAnimatedUptime(original string, currentTime time.Time) string {
	// Add live seconds ticker for visual feedback
	seconds := currentTime.Second()
	ticker := ""
	if seconds%4 == 0 {
		ticker = "⏰"
	} else {
		ticker = "⏱️ "
	}

	return ticker + original
}

func (m *animatedTableModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *animatedTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "a":
			// Toggle animations
			if m.animationsEnabled {
				m.table.DisableAnimations()
				m.animationsEnabled = false
				m.statusMessage = "Animations disabled"
			} else {
				if cmd := m.table.EnableAnimations(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				m.animationsEnabled = true
				m.statusMessage = "Animations enabled"
			}
			return m, tea.Batch(cmds...)
		case "tab":
			// Toggle between regular and animated view
			m.showAnimated = !m.showAnimated
			if !m.showAnimated {
				m.table.ClearAnimatedFormatter()
				m.statusMessage = "Switched to static view"
			} else {
				// Re-apply animated formatter (recreate it)
				animatedFormatter := func(data vtable.Data[vtable.TableRow], index int, ctx vtable.RenderContext,
					animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) vtable.RenderResult {

					row := data.Item
					if len(row.Cells) < 6 {
						return vtable.RenderResult{Content: "Invalid row data"}
					}

					counter := 0
					if c, ok := animationState["counter"]; ok {
						if ci, ok := c.(int); ok {
							counter = ci
						}
					}
					counter++

					cpu, _ := strconv.ParseFloat(strings.TrimSuffix(row.Cells[2], "%"), 64)
					memory, _ := strconv.ParseFloat(strings.TrimSuffix(row.Cells[3], "%"), 64)
					status := row.Cells[1]

					var content strings.Builder

					if data.Selected && isCursor {
						content.WriteString("✓>")
					} else if data.Selected {
						content.WriteString("✓ ")
					} else if isCursor {
						content.WriteString("> ")
					} else {
						content.WriteString("  ")
					}

					serverName := row.Cells[0]
					nameStyle := getServerNameStyle(status)
					content.WriteString(nameStyle.Render(fmt.Sprintf("%-10s", serverName)))
					content.WriteString(" │ ")

					statusContent := getAnimatedStatus(status, counter)
					content.WriteString(fmt.Sprintf("%-10s", statusContent))
					content.WriteString(" │ ")

					cpuContent := getAnimatedMetric(cpu, row.Cells[2], "cpu", counter)
					content.WriteString(fmt.Sprintf("%8s", cpuContent))
					content.WriteString(" │ ")

					memContent := getAnimatedMetric(memory, row.Cells[3], "memory", counter)
					content.WriteString(fmt.Sprintf("%8s", memContent))
					content.WriteString(" │ ")

					networkContent := getAnimatedNetwork(row.Cells[4], counter)
					content.WriteString(fmt.Sprintf("%12s", networkContent))
					content.WriteString(" │ ")

					uptimeContent := getAnimatedUptime(row.Cells[5], ctx.CurrentTime)
					content.WriteString(fmt.Sprintf("%8s", uptimeContent))

					return vtable.RenderResult{
						Content: content.String(),
						RefreshTriggers: []vtable.RefreshTrigger{{
							Type:     vtable.TriggerTimer,
							Interval: 200 * time.Millisecond,
						}},
						AnimationState: map[string]any{
							"counter":   counter,
							"timestamp": ctx.CurrentTime.Unix(),
						},
					}
				}
				m.table.SetAnimatedFormatter(animatedFormatter)
				m.statusMessage = "Switched to animated view"
			}
			return m, tea.Batch(cmds...)
		case " ", "space":
			m.table.ToggleCurrentSelection()
			return m, tea.Batch(cmds...)
		case "enter":
			// Restart selected servers
			selectedIndices := m.table.GetSelectedIndices()
			if len(selectedIndices) > 0 {
				restarted := m.provider.RestartServers(selectedIndices)
				m.table.ClearSelection()
				m.table.RefreshData()
				m.statusMessage = fmt.Sprintf("Restarted %d servers", restarted)
			}
			return m, tea.Batch(cmds...)
		case "1":
			// Show all servers
			m.table.ClearFilters()
			m.activeFilter = "all"
			return m, tea.Batch(cmds...)
		case "2":
			// Show only healthy servers
			m.table.ClearFilters()
			m.table.SetFilter("status", "healthy")
			m.activeFilter = "healthy"
			return m, tea.Batch(cmds...)
		case "3":
			// Show only warning servers
			m.table.ClearFilters()
			m.table.SetFilter("status", "warning")
			m.activeFilter = "warning"
			return m, tea.Batch(cmds...)
		case "4":
			// Show only critical servers
			m.table.ClearFilters()
			m.table.SetFilter("status", "critical")
			m.activeFilter = "critical"
			return m, tea.Batch(cmds...)
		case "5":
			// Show high CPU servers (>80%)
			m.table.ClearFilters()
			m.table.SetFilter("high_cpu", true)
			m.activeFilter = "high CPU"
			return m, tea.Batch(cmds...)
		case "s":
			// Cycle through sorts
			m.table.ClearSort()
			switch m.activeSort {
			case "none":
				m.table.SetSort("cpu", "desc")
				m.activeSort = "cpu desc"
			case "cpu desc":
				m.table.SetSort("memory", "desc")
				m.activeSort = "memory desc"
			case "memory desc":
				m.table.SetSort("name", "asc")
				m.activeSort = "name asc"
			default:
				m.table.ClearSort()
				m.activeSort = "none"
			}
			return m, tea.Batch(cmds...)
		}
	case vtable.GlobalAnimationTickMsg:
		// Track animation ticks and update server metrics
		m.tickCount++
		m.lastUpdate = time.Now()
		m.provider.UpdateMetrics()
	}

	// Update the table
	updatedTable, cmd := m.table.Update(msg)
	m.table = updatedTable.(*vtable.TeaTable)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *animatedTableModel) View() string {
	if m.quitting {
		return "Server monitoring stopped.\n"
	}

	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("🖥️  Server Monitoring Dashboard")

	sb.WriteString(title + "\n\n")

	// Status bar
	viewMode := "Static"
	if m.showAnimated {
		viewMode = "Animated"
	}

	animStatus := "OFF"
	animColor := lipgloss.Color("9")
	if m.animationsEnabled {
		animStatus = "ON"
		animColor = lipgloss.Color("10")
	}

	statusBar := fmt.Sprintf("View: %s | Animations: %s | Filter: %s | Sort: %s | Ticks: %d",
		lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("15")).Padding(0, 1).Render(viewMode),
		lipgloss.NewStyle().Background(animColor).Foreground(lipgloss.Color("0")).Padding(0, 1).Render(animStatus),
		lipgloss.NewStyle().Background(lipgloss.Color("208")).Foreground(lipgloss.Color("0")).Padding(0, 1).Render(m.activeFilter),
		lipgloss.NewStyle().Background(lipgloss.Color("93")).Foreground(lipgloss.Color("0")).Padding(0, 1).Render(m.activeSort),
		m.tickCount,
	)

	sb.WriteString(statusBar + "\n\n")

	// Table
	sb.WriteString(m.table.View())

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	help := helpStyle.Render(
		"Controls: ↑/↓ navigate • SPACE select • ENTER restart selected • TAB toggle view • A toggle animations • S cycle sort • Q quit\n" +
			"Filters: 1 all • 2 healthy • 3 warning • 4 critical • 5 high CPU\n" +
			"Features: Row-level animations • Live metrics • Real-time status • Sorting • Multi-selection")

	sb.WriteString("\n" + help)

	// Status message
	if m.statusMessage != "" {
		sb.WriteString("\n\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)
		sb.WriteString(statusStyle.Render(m.statusMessage))
		m.statusMessage = ""
	}

	return sb.String()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	model := newAnimatedTableModel()

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
