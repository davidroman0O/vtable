package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// FileEntry represents a file or directory
type FileEntry struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	IsDir        bool        `json:"is_dir"`
	Size         int64       `json:"size"`
	ModTime      time.Time   `json:"mod_time"`
	Permissions  fs.FileMode `json:"permissions"`
	IsHidden     bool        `json:"is_hidden"`
	Extension    string      `json:"extension"`
	IsExecutable bool        `json:"is_executable"`
}

// FileSystemProvider manages file system data
type FileSystemProvider struct {
	currentPath    string
	entries        []FileEntry
	selection      map[int]bool
	showHidden     bool
	sortBy         string // "name", "size", "date", "type"
	sortDescending bool
	lastError      error
	pathHistory    []string
	historyIndex   int
}

func NewFileSystemProvider(startPath string) *FileSystemProvider {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			startPath = "/"
		}
	}

	provider := &FileSystemProvider{
		currentPath:    startPath,
		selection:      make(map[int]bool),
		showHidden:     false,
		sortBy:         "name",
		sortDescending: false,
		pathHistory:    []string{startPath},
		historyIndex:   0,
	}

	provider.loadDirectory()
	return provider
}

func (p *FileSystemProvider) loadDirectory() {
	entries, err := os.ReadDir(p.currentPath)
	if err != nil {
		p.lastError = err
		p.entries = []FileEntry{}
		return
	}

	p.lastError = nil
	p.entries = make([]FileEntry, 0, len(entries))

	// Add parent directory entry if not at root
	if p.currentPath != "/" && p.currentPath != "" {
		parentEntry := FileEntry{
			Name:    "..",
			Path:    filepath.Dir(p.currentPath),
			IsDir:   true,
			Size:    0,
			ModTime: time.Now(),
		}
		p.entries = append(p.entries, parentEntry)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		isHidden := strings.HasPrefix(name, ".")

		// Skip hidden files if not showing them
		if isHidden && !p.showHidden && name != ".." {
			continue
		}

		fullPath := filepath.Join(p.currentPath, name)

		fileEntry := FileEntry{
			Name:         name,
			Path:         fullPath,
			IsDir:        info.IsDir(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			Permissions:  info.Mode(),
			IsHidden:     isHidden,
			Extension:    getFileExtension(name),
			IsExecutable: isExecutable(info.Mode()),
		}

		p.entries = append(p.entries, fileEntry)
	}

	p.sortEntries()
	p.selection = make(map[int]bool) // Clear selection when changing directories
}

func (p *FileSystemProvider) sortEntries() {
	sort.Slice(p.entries, func(i, j int) bool {
		// Always keep ".." at the top
		if p.entries[i].Name == ".." {
			return true
		}
		if p.entries[j].Name == ".." {
			return false
		}

		// Directories first (unless sorting by name only)
		if p.sortBy != "name" {
			if p.entries[i].IsDir != p.entries[j].IsDir {
				return p.entries[i].IsDir
			}
		}

		var less bool
		switch p.sortBy {
		case "size":
			less = p.entries[i].Size < p.entries[j].Size
		case "date":
			less = p.entries[i].ModTime.Before(p.entries[j].ModTime)
		case "type":
			ext1 := p.entries[i].Extension
			ext2 := p.entries[j].Extension
			if ext1 == ext2 {
				less = strings.ToLower(p.entries[i].Name) < strings.ToLower(p.entries[j].Name)
			} else {
				less = ext1 < ext2
			}
		default: // "name"
			less = strings.ToLower(p.entries[i].Name) < strings.ToLower(p.entries[j].Name)
		}

		if p.sortDescending {
			return !less
		}
		return less
	})
}

func (p *FileSystemProvider) GetTotal() int {
	return len(p.entries)
}

func (p *FileSystemProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[FileEntry], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.entries) {
		return []vtable.Data[FileEntry]{}, nil
	}

	end := start + count
	if end > len(p.entries) {
		end = len(p.entries)
	}

	result := make([]vtable.Data[FileEntry], end-start)
	for i := start; i < end; i++ {
		result[i-start] = vtable.Data[FileEntry]{
			ID:       fmt.Sprintf("file-%d", i),
			Item:     p.entries[i],
			Selected: p.selection[i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Navigation methods
func (p *FileSystemProvider) NavigateToDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}

	// Add to history if it's a new path
	if path != p.currentPath {
		// Remove any history after current position
		p.pathHistory = p.pathHistory[:p.historyIndex+1]
		p.pathHistory = append(p.pathHistory, path)
		p.historyIndex = len(p.pathHistory) - 1
	}

	p.currentPath = path
	p.loadDirectory()
	return nil
}

func (p *FileSystemProvider) NavigateBack() bool {
	if p.historyIndex > 0 {
		p.historyIndex--
		p.currentPath = p.pathHistory[p.historyIndex]
		p.loadDirectory()
		return true
	}
	return false
}

func (p *FileSystemProvider) NavigateForward() bool {
	if p.historyIndex < len(p.pathHistory)-1 {
		p.historyIndex++
		p.currentPath = p.pathHistory[p.historyIndex]
		p.loadDirectory()
		return true
	}
	return false
}

func (p *FileSystemProvider) NavigateUp() error {
	parent := filepath.Dir(p.currentPath)
	if parent != p.currentPath {
		return p.NavigateToDirectory(parent)
	}
	return nil
}

func (p *FileSystemProvider) ToggleHidden() {
	p.showHidden = !p.showHidden
	p.loadDirectory()
}

func (p *FileSystemProvider) SetSort(sortBy string) {
	if p.sortBy == sortBy {
		p.sortDescending = !p.sortDescending
	} else {
		p.sortBy = sortBy
		p.sortDescending = false
	}
	p.sortEntries()
}

func (p *FileSystemProvider) GetCurrentPath() string {
	return p.currentPath
}

func (p *FileSystemProvider) GetLastError() error {
	return p.lastError
}

func (p *FileSystemProvider) GetSortInfo() (string, bool) {
	return p.sortBy, p.sortDescending
}

// Implement required DataProvider methods
func (p *FileSystemProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *FileSystemProvider) SetSelected(index int, selected bool) bool {
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

func (p *FileSystemProvider) SelectAll() bool {
	for i := 0; i < len(p.entries); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *FileSystemProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *FileSystemProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *FileSystemProvider) GetItemID(item *FileEntry) string {
	return item.Path
}

func (p *FileSystemProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.entries) {
			ids = append(ids, p.entries[idx].Path)
		}
	}
	return ids
}

func (p *FileSystemProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *FileSystemProvider) SelectRange(startID, endID string) bool {
	return true
}

// Helper functions
func getFileExtension(name string) string {
	ext := filepath.Ext(name)
	if ext != "" {
		return ext[1:] // Remove the dot
	}
	return ""
}

func isExecutable(mode fs.FileMode) bool {
	return mode&0111 != 0
}

func getFileIcon(entry FileEntry) string {
	if entry.Name == ".." {
		return "📁" // Parent directory
	}

	if entry.IsDir {
		if entry.IsHidden {
			return "📂" // Hidden directory
		}
		return "📁" // Directory
	}

	// File icons based on extension
	switch strings.ToLower(entry.Extension) {
	case "go":
		return "🐹"
	case "js", "ts":
		return "🟨"
	case "py":
		return "🐍"
	case "html", "htm":
		return "🌐"
	case "css":
		return "🎨"
	case "json":
		return "📋"
	case "md", "markdown":
		return "📝"
	case "txt":
		return "📄"
	case "pdf":
		return "📕"
	case "zip", "tar", "gz", "rar":
		return "📦"
	case "jpg", "jpeg", "png", "gif", "svg":
		return "🖼️"
	case "mp3", "wav", "flac":
		return "🎵"
	case "mp4", "avi", "mkv", "mov":
		return "🎬"
	case "exe", "bin":
		return "⚙️"
	default:
		if entry.IsExecutable {
			return "⚙️"
		}
		if entry.IsHidden {
			return "👻"
		}
		return "📄"
	}
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fK", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fM", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1fG", float64(size)/(1024*1024*1024))
	}
}

func getPermissionString(mode fs.FileMode) string {
	return mode.String()
}

// Main application model
type FileNavigatorModel struct {
	fileList      *vtable.TeaList[FileEntry]
	provider      *FileSystemProvider
	statusMessage string
	viewMode      string // "list" or "details"
}

func newFileNavigatorDemo() *FileNavigatorModel {
	// Start in current directory
	startPath, _ := os.Getwd()
	provider := NewFileSystemProvider(startPath)

	// Configure viewport
	viewportConfig := vtable.ViewportConfig{
		Height:               15,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 12,
		ChunkSize:            50,
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

	// Create formatter
	formatter := func(data vtable.Data[FileEntry], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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

		// Get file icon
		icon := getFileIcon(entry)

		// Format based on view mode
		name := entry.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		// Color coding
		nameStyle := lipgloss.NewStyle()
		if entry.IsDir {
			nameStyle = nameStyle.Foreground(lipgloss.Color("39")).Bold(true) // Blue for directories
		} else if entry.IsExecutable {
			nameStyle = nameStyle.Foreground(lipgloss.Color("46")) // Green for executables
		} else if entry.IsHidden {
			nameStyle = nameStyle.Foreground(lipgloss.Color("8")) // Gray for hidden
		} else {
			nameStyle = nameStyle.Foreground(lipgloss.Color("252")) // Default
		}

		styledName := nameStyle.Render(name)

		// Size formatting
		sizeStr := ""
		if !entry.IsDir && entry.Name != ".." {
			sizeStr = formatFileSize(entry.Size)
		}

		// Date formatting
		dateStr := entry.ModTime.Format("Jan 02 15:04")

		return fmt.Sprintf("%s%s %-45s %8s %s",
			prefix,
			icon,
			styledName,
			sizeStr,
			dateStr,
		)
	}

	// Create the list
	list, err := vtable.NewTeaList(viewportConfig, provider, styleConfig, formatter)
	if err != nil {
		log.Fatal(err)
	}

	return &FileNavigatorModel{
		fileList:      list,
		provider:      provider,
		viewMode:      "list",
		statusMessage: " ", // Always reserve space
	}
}

func (m *FileNavigatorModel) Init() tea.Cmd {
	return m.fileList.Init()
}

func (m *FileNavigatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case " ", "space":
			// Toggle selection
			m.fileList.ToggleCurrentSelection()
			return m, nil

		case "enter":
			// Navigate into directory or show file info
			if item, ok := m.fileList.GetCurrentItem(); ok {
				if item.IsDir {
					err := m.provider.NavigateToDirectory(item.Path)
					if err != nil {
						m.statusMessage = fmt.Sprintf("Error: %v", err)
					} else {
						m.fileList.RefreshData()
						m.statusMessage = fmt.Sprintf("Navigated to: %s", item.Path)
					}
				} else {
					m.statusMessage = fmt.Sprintf("File: %s (%s)", item.Name, formatFileSize(item.Size))
				}
			}
			return m, nil

		case "backspace", "left", "h":
			// Navigate back in history
			if m.provider.NavigateBack() {
				m.fileList.RefreshData()
				m.statusMessage = fmt.Sprintf("Back to: %s", m.provider.GetCurrentPath())
			} else {
				m.statusMessage = "No previous directory in history"
			}
			return m, nil

		case "right", "l":
			// Navigate forward in history
			if m.provider.NavigateForward() {
				m.fileList.RefreshData()
				m.statusMessage = fmt.Sprintf("Forward to: %s", m.provider.GetCurrentPath())
			} else {
				m.statusMessage = "No next directory in history"
			}
			return m, nil

		case "u":
			// Navigate up to parent directory
			err := m.provider.NavigateUp()
			if err != nil {
				m.statusMessage = fmt.Sprintf("Error: %v", err)
			} else {
				m.fileList.RefreshData()
				m.statusMessage = fmt.Sprintf("Up to: %s", m.provider.GetCurrentPath())
			}
			return m, nil

		case ".":
			// Toggle hidden files
			m.provider.ToggleHidden()
			m.fileList.RefreshData()
			if m.provider.showHidden {
				m.statusMessage = "Showing hidden files"
			} else {
				m.statusMessage = "Hiding hidden files"
			}
			return m, nil

		case "r":
			// Refresh current directory
			m.provider.loadDirectory()
			m.fileList.RefreshData()
			m.statusMessage = "Directory refreshed"
			return m, nil

		// Sorting
		case "s":
			// Sort by size
			m.provider.SetSort("size")
			m.fileList.RefreshData()
			sortBy, desc := m.provider.GetSortInfo()
			direction := "ascending"
			if desc {
				direction = "descending"
			}
			m.statusMessage = fmt.Sprintf("Sorted by %s (%s)", sortBy, direction)
			return m, nil

		case "n":
			// Sort by name
			m.provider.SetSort("name")
			m.fileList.RefreshData()
			sortBy, desc := m.provider.GetSortInfo()
			direction := "ascending"
			if desc {
				direction = "descending"
			}
			m.statusMessage = fmt.Sprintf("Sorted by %s (%s)", sortBy, direction)
			return m, nil

		case "t":
			// Sort by time
			m.provider.SetSort("date")
			m.fileList.RefreshData()
			sortBy, desc := m.provider.GetSortInfo()
			direction := "ascending"
			if desc {
				direction = "descending"
			}
			m.statusMessage = fmt.Sprintf("Sorted by %s (%s)", sortBy, direction)
			return m, nil

		case "e":
			// Sort by extension/type
			m.provider.SetSort("type")
			m.fileList.RefreshData()
			sortBy, desc := m.provider.GetSortInfo()
			direction := "ascending"
			if desc {
				direction = "descending"
			}
			m.statusMessage = fmt.Sprintf("Sorted by %s (%s)", sortBy, direction)
			return m, nil

		case "a":
			// Select all
			m.fileList.SelectAll()
			m.statusMessage = fmt.Sprintf("Selected all %d items", m.provider.GetTotal())
			return m, nil

		case "c":
			// Clear selection
			m.fileList.ClearSelection()
			m.statusMessage = "Selection cleared"
			return m, nil
		}
	}

	// Update the list
	updatedList, cmd := m.fileList.Update(msg)
	m.fileList = updatedList.(*vtable.TeaList[FileEntry])
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *FileNavigatorModel) View() string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("VTable Example 11: File System Navigator")

	sb.WriteString(title + "\n\n")

	// Path bar
	pathStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)

	currentPath := m.provider.GetCurrentPath()
	if len(currentPath) > 80 {
		// Truncate long paths
		currentPath = "..." + currentPath[len(currentPath)-77:]
	}

	pathBar := fmt.Sprintf("📁 %s", currentPath)
	sb.WriteString(pathStyle.Render(pathBar) + "\n\n")

	// Sort and selection info
	sortBy, desc := m.provider.GetSortInfo()
	direction := "↑"
	if desc {
		direction = "↓"
	}

	selectedCount := len(m.provider.GetSelectedIndices())
	totalCount := m.provider.GetTotal()

	infoBar := fmt.Sprintf("Sort: %s %s | Items: %d | Selected: %d | Hidden: %v",
		sortBy, direction, totalCount, selectedCount, m.provider.showHidden)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	sb.WriteString(infoStyle.Render(infoBar) + "\n\n")

	// The main file list
	sb.WriteString(m.fileList.View())

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	help := helpStyle.Render(
		"Navigation: ↑/↓ browse • ENTER enter dir • ←/→/h/l history • u parent • SPACE select • a select all • c clear\n" +
			"Sorting: n name • s size • t time • e type | View: . toggle hidden • r refresh\n" +
			"Icons: 📁 dir • 🐹 go • 🟨 js/ts • 🐍 py • 🌐 html • 📄 file • ⚙️ executable • q quit")

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
	model := newFileNavigatorDemo()

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
