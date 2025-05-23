package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// LogEntry represents a system log entry that changes over time
type LogEntry struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	Count     int       `json:"count"`
	Status    string    `json:"status"`
}

// DynamicDataProvider manages log entries with real-time updates
type DynamicDataProvider struct {
	entries       []LogEntry
	selection     map[int]bool
	nextID        int
	updateTicker  *time.Ticker
	lastUpdate    time.Time
	totalAdded    int
	totalRemoved  int
	totalModified int
}

func NewDynamicDataProvider() *DynamicDataProvider {
	provider := &DynamicDataProvider{
		entries:      generateInitialLogs(),
		selection:    make(map[int]bool),
		nextID:       100,
		lastUpdate:   time.Now(),
		updateTicker: time.NewTicker(2 * time.Second),
	}
	return provider
}

func generateInitialLogs() []LogEntry {
	services := []string{"api-gateway", "auth-service", "database", "cache", "web-frontend", "queue-worker"}
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	messages := []string{
		"Request processed successfully",
		"Database connection established",
		"Cache miss for key user_123",
		"Authentication token expired",
		"Rate limit exceeded",
		"Deployment completed",
		"Health check passed",
		"Memory usage high",
		"Disk space low",
		"Connection timeout",
	}

	logs := make([]LogEntry, 0, 20)
	for i := 1; i <= 20; i++ {
		logs = append(logs, LogEntry{
			ID:        i,
			Timestamp: time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second),
			Level:     levels[rand.Intn(len(levels))],
			Service:   services[rand.Intn(len(services))],
			Message:   messages[rand.Intn(len(messages))],
			Count:     rand.Intn(50) + 1,
			Status:    "active",
		})
	}
	return logs
}

func (p *DynamicDataProvider) GetTotal() int {
	return len(p.entries)
}

func (p *DynamicDataProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[LogEntry], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.entries) {
		return []vtable.Data[LogEntry]{}, nil
	}

	end := start + count
	if end > len(p.entries) {
		end = len(p.entries)
	}

	result := make([]vtable.Data[LogEntry], end-start)
	for i := start; i < end; i++ {
		result[i-start] = vtable.Data[LogEntry]{
			ID:       fmt.Sprintf("log-%d", p.entries[i].ID),
			Item:     p.entries[i],
			Selected: p.selection[i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement required DataProvider methods
func (p *DynamicDataProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *DynamicDataProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.entries) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *DynamicDataProvider) SelectAll() bool {
	for i := 0; i < len(p.entries); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *DynamicDataProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *DynamicDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *DynamicDataProvider) GetItemID(item *LogEntry) string {
	return fmt.Sprintf("%d", item.ID)
}

func (p *DynamicDataProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.entries) {
			ids = append(ids, fmt.Sprintf("%d", p.entries[idx].ID))
		}
	}
	return ids
}

func (p *DynamicDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *DynamicDataProvider) SelectRange(startID, endID string) bool {
	return true
}

// Dynamic data operations
func (p *DynamicDataProvider) AddNewEntry() {
	services := []string{"api-gateway", "auth-service", "database", "cache", "web-frontend", "queue-worker"}
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	messages := []string{
		"New request received",
		"Service started successfully",
		"Configuration updated",
		"Backup completed",
		"Connection established",
		"Task queued",
		"Alert triggered",
		"Performance metrics updated",
	}

	newEntry := LogEntry{
		ID:        p.nextID,
		Timestamp: time.Now(),
		Level:     levels[rand.Intn(len(levels))],
		Service:   services[rand.Intn(len(services))],
		Message:   messages[rand.Intn(len(messages))],
		Count:     1,
		Status:    "active",
	}

	// Add to the beginning for newest-first ordering
	p.entries = append([]LogEntry{newEntry}, p.entries...)
	p.nextID++
	p.totalAdded++

	// Adjust selection indices due to insertion at beginning
	newSelection := make(map[int]bool)
	for idx := range p.selection {
		newSelection[idx+1] = true
	}
	p.selection = newSelection
}

func (p *DynamicDataProvider) RemoveOldEntries() {
	// Remove entries older than 10 minutes
	cutoff := time.Now().Add(-10 * time.Minute)
	originalLen := len(p.entries)

	filtered := make([]LogEntry, 0, len(p.entries))
	indexMap := make(map[int]int) // old index -> new index

	newIdx := 0
	for oldIdx, entry := range p.entries {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
			indexMap[oldIdx] = newIdx
			newIdx++
		}
	}

	p.entries = filtered
	p.totalRemoved += originalLen - len(p.entries)

	// Update selection indices
	newSelection := make(map[int]bool)
	for oldIdx := range p.selection {
		if newIdx, exists := indexMap[oldIdx]; exists {
			newSelection[newIdx] = true
		}
	}
	p.selection = newSelection
}

func (p *DynamicDataProvider) UpdateEntryCounts() {
	// Simulate updating counts for existing entries
	for i := range p.entries {
		if rand.Float32() < 0.3 { // 30% chance to update
			p.entries[i].Count += rand.Intn(5) + 1
			p.totalModified++

			// Sometimes change status
			if rand.Float32() < 0.1 { // 10% chance
				if p.entries[i].Status == "active" {
					p.entries[i].Status = "resolved"
				} else {
					p.entries[i].Status = "active"
				}
			}
		}
	}
}

func (p *DynamicDataProvider) SimulateRealtimeUpdates() {
	// Add new entries
	if rand.Float32() < 0.8 { // 80% chance
		numToAdd := rand.Intn(3) + 1
		for i := 0; i < numToAdd; i++ {
			p.AddNewEntry()
		}
	}

	// Update existing entries
	p.UpdateEntryCounts()

	// Remove old entries occasionally
	if rand.Float32() < 0.3 { // 30% chance
		p.RemoveOldEntries()
	}

	p.lastUpdate = time.Now()
}

func (p *DynamicDataProvider) DeleteSelected() int {
	if len(p.selection) == 0 {
		return 0
	}

	// Sort indices in descending order to delete from end to beginning
	indices := p.GetSelectedIndices()
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}

	deleted := 0
	for _, idx := range indices {
		if idx >= 0 && idx < len(p.entries) {
			// Remove entry
			p.entries = append(p.entries[:idx], p.entries[idx+1:]...)
			deleted++
		}
	}

	// Clear selection and update totals
	p.ClearSelection()
	p.totalRemoved += deleted

	return deleted
}

func (p *DynamicDataProvider) GetStats() (int, int, int) {
	return p.totalAdded, p.totalRemoved, p.totalModified
}

// Custom tick message for real-time updates
type TickMsg struct{}

// Main application model
type DynamicDataModel struct {
	logList        *vtable.TeaList[LogEntry]
	provider       *DynamicDataProvider
	autoUpdate     bool
	statusMessage  string
	updateInterval time.Duration
}

func newDynamicDataDemo() *DynamicDataModel {
	provider := NewDynamicDataProvider()

	// Configure viewport - consistent with other examples
	viewportConfig := vtable.ViewportConfig{
		Height:               12,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 9,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create style config
	styleConfig := vtable.StyleConfig{
		BorderStyle:      "245",
		HeaderStyle:      "bold 252 on 238",
		RowStyle:         "252",
		SelectedRowStyle: "bold 252 on 63",
	}

	// Create formatter for log entries
	formatter := func(data vtable.Data[LogEntry], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		entry := data.Item

		// ASCII-only selection indicator (exactly 2 chars)
		prefix := "  "
		if data.Selected && isCursor {
			prefix = "*>"
		} else if data.Selected {
			prefix = "* "
		} else if isCursor {
			prefix = "> "
		}

		// Format timestamp
		timeStr := entry.Timestamp.Format("15:04:05")

		// Level styling
		levelStyle := getLevelStyle(entry.Level)
		coloredLevel := levelStyle.Render(fmt.Sprintf("%-5s", entry.Level))

		// Status indicator
		statusChar := "●"
		if entry.Status == "resolved" {
			statusChar = "○"
		}

		// Truncate message if too long
		message := entry.Message
		if len(message) > 35 {
			message = message[:32] + "..."
		}

		return fmt.Sprintf("%s%s %s %s %-12s %s [%d] %s",
			prefix,
			timeStr,
			coloredLevel,
			statusChar,
			entry.Service,
			message,
			entry.Count,
			entry.Status,
		)
	}

	// Create the list
	list, err := vtable.NewTeaList(viewportConfig, provider, styleConfig, formatter)
	if err != nil {
		log.Fatal(err)
	}

	return &DynamicDataModel{
		logList:        list,
		provider:       provider,
		autoUpdate:     true,
		updateInterval: 2 * time.Second,
	}
}

// Helper function for level styling
func getLevelStyle(level string) lipgloss.Style {
	switch level {
	case "ERROR":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	case "WARN":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "INFO":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	case "DEBUG":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Default
	}
}

func (m *DynamicDataModel) Init() tea.Cmd {
	if m.autoUpdate {
		return tea.Batch(
			m.logList.Init(),
			m.tick(),
		)
	}
	return m.logList.Init()
}

func (m *DynamicDataModel) tick() tea.Cmd {
	return tea.Tick(m.updateInterval, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m *DynamicDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case " ", "space":
			// Toggle selection
			m.logList.ToggleCurrentSelection()
			return m, nil

		case "a":
			// Manually add new entry
			m.provider.AddNewEntry()
			m.logList.RefreshData()
			m.statusMessage = "Added new log entry"
			return m, nil

		case "d", "delete":
			// Delete selected entries
			deleted := m.provider.DeleteSelected()
			if deleted > 0 {
				m.logList.RefreshData()
				m.statusMessage = fmt.Sprintf("Deleted %d entries", deleted)
			} else {
				m.statusMessage = "No entries selected for deletion"
			}
			return m, nil

		case "r":
			// Force refresh/update
			m.provider.SimulateRealtimeUpdates()
			m.logList.RefreshData()
			m.statusMessage = "Forced data refresh"
			return m, nil

		case "t":
			// Toggle auto-update
			m.autoUpdate = !m.autoUpdate
			if m.autoUpdate {
				m.statusMessage = "Auto-update enabled"
				return m, m.tick()
			} else {
				m.statusMessage = "Auto-update disabled"
			}
			return m, nil

		case "c":
			// Clear all entries
			m.provider.entries = []LogEntry{}
			m.provider.ClearSelection()
			m.logList.RefreshData()
			m.statusMessage = "Cleared all entries"
			return m, nil
		}

	case TickMsg:
		if m.autoUpdate {
			// Simulate real-time updates
			m.provider.SimulateRealtimeUpdates()
			m.logList.RefreshData()
			return m, m.tick()
		}
	}

	// Update the list
	updatedList, cmd := m.logList.Update(msg)
	m.logList = updatedList.(*vtable.TeaList[LogEntry])
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *DynamicDataModel) View() string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("VTable Example 10: Dynamic Data")

	sb.WriteString(title + "\n\n")

	// Status bar with fixed-width formatting to prevent jitter
	added, removed, modified := m.provider.GetStats()
	autoStatus := "OFF"
	if m.autoUpdate {
		autoStatus = "ON "
	}

	// Use fixed-width formatting to prevent layout shifts
	statusBar := fmt.Sprintf("Total:%4d | Added:%4d | Removed:%4d | Modified:%5d | Auto:%3s | Last:%s",
		m.provider.GetTotal(),
		added,
		removed,
		modified,
		autoStatus,
		m.provider.lastUpdate.Format("15:04:05"),
	)

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)

	sb.WriteString(statusStyle.Render(statusBar) + "\n\n")

	// The main list
	sb.WriteString(m.logList.View())

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	help := helpStyle.Render(
		"Controls: ↑/↓ navigate • SPACE select • a add entry • d delete selected • r force refresh • t toggle auto-update • c clear all • q quit\n" +
			"Features: Real-time updates • Auto-add entries • Auto-remove old entries • Count updates • Status changes\n" +
			"Data Flow: New entries appear at top • Old entries auto-expire • Counts increment over time")

	sb.WriteString("\n" + help)

	// Status message - always reserve space to prevent layout jitter
	statusMsgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	statusMsgText := m.statusMessage
	if statusMsgText == "" {
		statusMsgText = " " // Always show something to maintain consistent height
	}

	sb.WriteString("\n\n" + statusMsgStyle.Render(statusMsgText))
	m.statusMessage = "" // Clear after showing

	return sb.String()
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	model := newDynamicDataDemo()

	// Configure the Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Clean exit
	fmt.Print("\033[H\033[2J")
	fmt.Print("\033[?25h")
	fmt.Print("\n\n")
}
