package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// LogEntry represents a log entry with potentially long text
type LogEntry struct {
	ID        int
	Timestamp string
	Level     string
	Service   string
	Message   string
	Details   string
}

// LogDataProvider provides log data with long text for scrolling demonstration
type LogDataProvider struct {
	logs      []LogEntry
	selection map[int]bool
}

func NewLogDataProvider() *LogDataProvider {
	return &LogDataProvider{
		logs:      generateLogs(),
		selection: make(map[int]bool),
	}
}

func generateLogs() []LogEntry {
	return []LogEntry{
		{ID: 1, Timestamp: "10:30:15", Level: "ERROR", Service: "auth-service", Message: "🔐 Authentication failed for user john.doe@company.com due to invalid credentials - this is a very long message that needs scrolling", Details: "IP: 192.168.1.100, User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0 - additional details that are quite long"},
		{ID: 2, Timestamp: "10:30:16", Level: "WARN", Service: "api-gateway", Message: "⚡ Rate limit exceeded for API key abc123def456ghi789 from client application", Details: "Current rate: 1000 req/min, Limit: 500 req/min, Client: mobile-app-v2.1.0"},
		{ID: 3, Timestamp: "10:30:17", Level: "INFO", Service: "payment", Message: "💳 Payment processed successfully for transaction TX-98765432 with amount $299.99", Details: "Amount: $299.99, Card: ****1234, Merchant: TechStore Inc., Gateway: Stripe"},
		{ID: 4, Timestamp: "10:30:18", Level: "DEBUG", Service: "database", Message: "🔍 Connection pool statistics: active=45, idle=5, max=50 connections", Details: "Pool ID: db-pool-primary, Host: postgres-cluster-01.internal, Response time: 12ms"},
		{ID: 5, Timestamp: "10:30:19", Level: "ERROR", Service: "email-service", Message: "📧 Failed to send welcome email to new user registration due to SMTP timeout", Details: "SMTP Error: Connection timeout to mail.company.com:587, Recipient: newuser@email.com"},
		{ID: 6, Timestamp: "10:30:20", Level: "INFO", Service: "cache", Message: "🚀 Cache hit rate improved to 95.2% after optimization", Details: "Previous: 87.3%, Current: 95.2%, Memory usage: 2.1GB/4GB, Keys: 125,432"},
		{ID: 7, Timestamp: "10:30:21", Level: "WARN", Service: "storage", Message: "💾 Disk usage warning: partition /data is 85% full", Details: "Used: 850GB, Available: 150GB, Threshold: 80%, Auto-cleanup scheduled"},
	}
}

func (p *LogDataProvider) GetTotal() int {
	return len(p.logs)
}

func (p *LogDataProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.logs) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	end := start + count
	if end > len(p.logs) {
		end = len(p.logs)
	}

	result := make([]vtable.Data[vtable.TableRow], end-start)
	for i := start; i < end; i++ {
		log := p.logs[i]

		// Convert to table row
		tableRow := vtable.TableRow{
			Cells: []string{
				log.Timestamp,
				log.Level,
				log.Service,
				log.Message,
				log.Details,
			},
		}

		result[i-start] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("log-%d", log.ID),
			Item:     tableRow,
			Selected: p.selection[i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement remaining DataProvider methods
func (p *LogDataProvider) GetSelectionMode() vtable.SelectionMode { return vtable.SelectionMultiple }
func (p *LogDataProvider) SetSelected(index int, selected bool) bool {
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}
func (p *LogDataProvider) SelectAll() bool {
	for i := 0; i < len(p.logs); i++ {
		p.selection[i] = true
	}
	return true
}
func (p *LogDataProvider) ClearSelection() { p.selection = make(map[int]bool) }
func (p *LogDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}
func (p *LogDataProvider) GetSelectedIDs() []string                          { return []string{} }
func (p *LogDataProvider) SetSelectedByIDs(ids []string, selected bool) bool { return true }
func (p *LogDataProvider) SelectRange(startID, endID string) bool            { return true }
func (p *LogDataProvider) GetItemID(item *vtable.TableRow) string            { return "log" }

type cellAnimationModel struct {
	table             *vtable.TeaTable
	provider          *LogDataProvider
	quitting          bool
	statusMessage     string
	animationsEnabled bool
	animatedFormatter vtable.ItemFormatterAnimated[vtable.TableRow] // Store the animated formatter
}

func newCellAnimationModel() *cellAnimationModel {
	provider := NewLogDataProvider()

	// Configure table with columns that will demonstrate different scrolling modes
	config := vtable.TableConfig{
		Columns: []vtable.TableColumn{
			{Title: "Time", Width: 9, Alignment: vtable.AlignLeft, Field: "timestamp"},
			{Title: "Level", Width: 12, Alignment: vtable.AlignCenter, Field: "level"},
			{Title: "Service", Width: 16, Alignment: vtable.AlignLeft, Field: "service"},
			{Title: "Message", Width: 45, Alignment: vtable.AlignLeft, Field: "message"}, // Bounce scrolling
			{Title: "Details", Width: 40, Alignment: vtable.AlignLeft, Field: "details"}, // Smooth scrolling
		},
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: vtable.ViewportConfig{
			Height:               8,
			TopThresholdIndex:    1,
			BottomThresholdIndex: 6,
			ChunkSize:            20,
			InitialIndex:         0,
		},
	}

	table, err := vtable.NewTeaTable(config, provider, *vtable.DarkTheme())
	if err != nil {
		log.Fatal(err)
	}

	model := &cellAnimationModel{
		table:             table,
		provider:          provider,
		animationsEnabled: true,
	}

	// Set up a default formatter that just renders normally
	defaultFormatter := func(
		data vtable.Data[vtable.TableRow],
		index int,
		ctx vtable.RenderContext,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) string {
		return vtable.FormatTableRow(
			data,
			index,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
			config,
			*vtable.DarkTheme(),
		)
	}

	// Set the default formatter for when animations are disabled
	table.SetDefaultFormatter(defaultFormatter)

	// Set up the animated formatter with new comprehensive scrolling
	animatedFormatter := func(
		data vtable.Data[vtable.TableRow],
		index int,
		ctx vtable.RenderContext,
		animationState map[string]any,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) vtable.RenderResult {

		row := data.Item
		if len(row.Cells) < 5 {
			return vtable.RenderResult{Content: "Invalid row data"}
		}

		// Check if animations are enabled at the table level
		if !table.IsAnimationEnabled() {
			content := vtable.FormatTableRow(
				data,
				index,
				isCursor,
				isTopThreshold,
				isBottomThreshold,
				config,
				*vtable.DarkTheme(),
			)
			return vtable.RenderResult{Content: content}
		}

		// Get or initialize animation state for different scrolling elements
		messageState := make(map[string]any)
		detailsState := make(map[string]any)
		phase := 0.0

		// Extract existing states
		if state, ok := animationState["message_scroll"]; ok {
			if s, ok := state.(map[string]any); ok {
				messageState = s
			}
		}
		if state, ok := animationState["details_scroll"]; ok {
			if s, ok := state.(map[string]any); ok {
				detailsState = s
			}
		}
		if p, ok := animationState["phase"]; ok {
			if ph, ok := p.(float64); ok {
				phase = ph
			}
		}

		// Update phase for level animation
		deltaTime := ctx.DeltaTime.Seconds()
		phase += deltaTime * 3.0 // 3 cycles per second for level pulsing

		// Keep phase in reasonable range
		for phase >= 2*math.Pi {
			phase -= 2 * math.Pi
		}

		// Create modified row data for animation
		animatedRow := vtable.TableRow{
			Cells: make([]string, len(row.Cells)),
		}

		// Copy all cells first
		copy(animatedRow.Cells, row.Cells)

		// Apply animations to specific cells
		if isCursor {
			// Animate level with emoji switching
			level := row.Cells[1]
			intensity := (math.Sin(phase) + 1.0) / 2.0 // 0.0 to 1.0

			var animatedLevel string
			switch level {
			case "ERROR":
				if int(intensity*4)%2 == 0 {
					animatedLevel = "⚠️ " + level // "⚠️ ERROR"
				} else {
					animatedLevel = "🔴 " + level // "🔴 ERROR"
				}
			case "WARN":
				if int(intensity*4)%2 == 0 {
					animatedLevel = "⚠️ " + level // "⚠️ WARN"
				} else {
					animatedLevel = "🟡 " + level // "🟡 WARN"
				}
			case "INFO":
				animatedLevel = "✅ " + level // "✅ INFO"
			case "DEBUG":
				animatedLevel = "🔍 " + level // "🔍 DEBUG"
			default:
				animatedLevel = level
			}

			animatedRow.Cells[1] = animatedLevel

			// For cursor row: Use ultra-slow speed for testing
			message := row.Cells[3]
			details := row.Cells[4]

			// Use reasonable config for testing deltaTime control
			testConfig := vtable.SimpleScrollConfig{
				Speed:          3, // Smooth and pleasant to watch
				WordAware:      true,
				PauseDuration:  1000 * time.Millisecond, // 1 second pause
				MinScrollWidth: 2,
			}

			// Message column: Ultra-slow scrolling
			scrolledMessage, newMessageState := vtable.CreateSimpleHorizontalScrolling(
				message,
				45, // Message column width
				testConfig,
				messageState,
				ctx.DeltaTime,
			)
			animatedRow.Cells[3] = scrolledMessage

			// Details column: Ultra-slow scrolling
			scrolledDetails, newDetailsState := vtable.CreateSimpleHorizontalScrolling(
				details,
				40, // Details column width
				testConfig,
				detailsState,
				ctx.DeltaTime,
			)
			animatedRow.Cells[4] = scrolledDetails

			// Update states
			messageState = newMessageState
			detailsState = newDetailsState

		} else {
			// For non-cursor rows, also use ultra-slow speed for consistent testing
			message := row.Cells[3]
			details := row.Cells[4]

			// Use reasonable config for testing deltaTime control
			testConfig := vtable.SimpleScrollConfig{
				Speed:          1.5, // Smooth and pleasant to watch
				WordAware:      true,
				PauseDuration:  1000 * time.Millisecond, // 1 second pause
				MinScrollWidth: 2,
			}

			scrolledMessage, newMessageState := vtable.CreateSimpleHorizontalScrolling(
				message,
				45,
				testConfig,
				messageState,
				ctx.DeltaTime,
			)
			animatedRow.Cells[3] = scrolledMessage
			messageState = newMessageState

			scrolledDetails, newDetailsState := vtable.CreateSimpleHorizontalScrolling(
				details,
				40,
				testConfig,
				detailsState,
				ctx.DeltaTime,
			)
			animatedRow.Cells[4] = scrolledDetails
			detailsState = newDetailsState
		}

		// Create animated data struct
		animatedData := vtable.Data[vtable.TableRow]{
			ID:       data.ID,
			Item:     animatedRow,
			Selected: data.Selected,
			Metadata: data.Metadata,
		}

		// Use the proper table formatting function
		content := vtable.FormatTableRow(
			animatedData,
			index,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
			config,
			*vtable.DarkTheme(),
		)

		return vtable.RenderResult{
			Content: content,
			RefreshTriggers: []vtable.RefreshTrigger{{
				Type:     vtable.TriggerTimer,
				Interval: 50 * time.Millisecond, // 20fps for smoother animations
			}},
			AnimationState: map[string]any{
				"message_scroll": messageState,
				"details_scroll": detailsState,
				"phase":          phase,
			},
		}
	}

	// Store the animated formatter and set it initially
	model.animatedFormatter = animatedFormatter
	table.SetAnimatedFormatter(animatedFormatter)

	return model
}

func (m *cellAnimationModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *cellAnimationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// Disable animations - library will handle stopping the loop and clearing cache
				m.table.DisableAnimations()
				m.animationsEnabled = false
				m.statusMessage = "Animations disabled"
			} else {
				// Enable animations - library will handle restarting the loop
				if cmd := m.table.EnableAnimations(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				m.animationsEnabled = true
				m.statusMessage = "Animations enabled"
			}
			return m, tea.Batch(cmds...)
		case " ", "space":
			m.table.ToggleCurrentSelection()
			return m, tea.Batch(cmds...)
		}
	}

	// Update table
	updatedTable, cmd := m.table.Update(msg)
	m.table = updatedTable.(*vtable.TeaTable)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *cellAnimationModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("🎯 Simple Bouncing Horizontal Scroll Demo")

	sb.WriteString(title + "\n\n")

	// Status
	animStatus := "OFF"
	animColor := lipgloss.Color("9")
	if m.animationsEnabled {
		animStatus = "ON"
		animColor = lipgloss.Color("10")
	}

	status := fmt.Sprintf("Animations: %s | Features: Always Visible Text, Speed Control, Word-Aware Bouncing",
		lipgloss.NewStyle().Background(animColor).Foreground(lipgloss.Color("0")).Padding(0, 1).Render(animStatus),
	)
	sb.WriteString(status + "\n\n")

	// Table
	sb.WriteString(m.table.View())

	// Help
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).MarginTop(1).Render(
		"Controls: ↑/↓ navigate • SPACE select • A toggle animations • Q quit\n\n" +
			"Simple Scrolling Demo by Row:\n" +
			"• Row 0: Word-aware + Slow smooth     • Row 1: Fast + Character-by-char\n" +
			"• Row 2: Default bounce (both)        • Row 3+: Custom pause timing\n" +
			"• Cursor: Fast word-aware + Fast char • All modes bounce properly!\n\n" +
			"Features: Text ALWAYS visible, proper bouncing, speed control,\n" +
			"word-aware stopping, and no more disappearing text!")

	sb.WriteString("\n" + help)

	if m.statusMessage != "" {
		sb.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(m.statusMessage))
		m.statusMessage = ""
	}

	return sb.String()
}

func main() {
	model := newCellAnimationModel()

	p := tea.NewProgram(
		model,
		// tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
