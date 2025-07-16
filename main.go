package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateMenu state = iota
	stateSearch
	stateSearchResults
	stateEntry
	stateEntryField
)

type searchStep int

const (
	searchStepQuery searchStep = iota
	searchStepSymptom
	searchStepProgram
	searchStepProgramVersion
	searchStepDistro
	searchStepDistroVersion
	searchStepSolution
	searchStepExecute
)

type entryStep int

const (
	entryStepSymptom entryStep = iota
	entryStepProgram
	entryStepProgramVersion
	entryStepDistro
	entryStepDistroVersion
	entryStepResources
	entryStepSolution
	entryStepConfirm
)

type model struct {
	state         state
	cursor        int
	
	// Search state
	searchStep    searchStep
	filter        Filter
	searchResults []ErrorReport
	
	// Entry state
	entryStep     entryStep
	currentReport ErrorReport
	
	// Multi-line text editing
	currentText   string
	textLines     []string
	textCursor    int
	charCursor    int // Position within current line
	
	// UI state
	message       string
	clipboard     string // Internal clipboard for copy/paste
}

func initialModel() model {
	return model{
		state:         stateMenu,
		cursor:        0,
		filter:        Filter{},
		searchResults: []ErrorReport{},
		currentReport: ErrorReport{
			Resources: []string{},
			Date:      time.Now(),
		},
		textLines: []string{""},
		textCursor: 0,
		charCursor: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateSearch:
			return m.updateSearch(msg)
		case stateSearchResults:
			return m.updateSearchResults(msg)
		case stateEntry:
			return m.updateEntry(msg)
		case stateEntryField:
			return m.updateEntryField(msg)
		}
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0:
			m.state = stateSearch
			m.searchStep = searchStepQuery
			m.filter = Filter{}
		case 1:
			m.state = stateEntry
			m.entryStep = entryStepSymptom
			m.currentReport = ErrorReport{
				Resources: []string{},
				Date:      time.Now(),
			}
		}
	}
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		m.cursor = 0
	case "enter":
		if m.searchStep == searchStepExecute {
			results, _ := SearchErrorReports(m.filter)
			m.searchResults = results
			m.state = stateSearchResults
			m.cursor = 0
		} else {
			m.searchStep++
		}
	case "tab":
		if m.searchStep < searchStepExecute {
			m.searchStep++
		}
	case "shift+tab":
		if m.searchStep > searchStepQuery {
			m.searchStep--
		}
	case "backspace":
		m.updateSearchField(msg.String())
	default:
		if len(msg.String()) == 1 {
			m.updateSearchField(msg.String())
		}
	}
	return m, nil
}

func (m *model) updateSearchField(input string) {
	switch m.searchStep {
	case searchStepQuery:
		if input == "backspace" {
			if len(m.filter.Q) > 0 {
				m.filter.Q = m.filter.Q[:len(m.filter.Q)-1]
			}
		} else {
			m.filter.Q += input
		}
	case searchStepSymptom:
		if input == "backspace" {
			if len(m.filter.Symptom) > 0 {
				m.filter.Symptom = m.filter.Symptom[:len(m.filter.Symptom)-1]
			}
		} else {
			m.filter.Symptom += input
		}
	case searchStepProgram:
		if input == "backspace" {
			if len(m.filter.Program) > 0 {
				m.filter.Program = m.filter.Program[:len(m.filter.Program)-1]
			}
		} else {
			m.filter.Program += input
		}
	case searchStepProgramVersion:
		if input == "backspace" {
			if len(m.filter.ProgramVersion) > 0 {
				m.filter.ProgramVersion = m.filter.ProgramVersion[:len(m.filter.ProgramVersion)-1]
			}
		} else {
			m.filter.ProgramVersion += input
		}
	case searchStepDistro:
		if input == "backspace" {
			if len(m.filter.Distro) > 0 {
				m.filter.Distro = m.filter.Distro[:len(m.filter.Distro)-1]
			}
		} else {
			m.filter.Distro += input
		}
	case searchStepDistroVersion:
		if input == "backspace" {
			if len(m.filter.DistroVersion) > 0 {
				m.filter.DistroVersion = m.filter.DistroVersion[:len(m.filter.DistroVersion)-1]
			}
		} else {
			m.filter.DistroVersion += input
		}
	case searchStepSolution:
		if input == "backspace" {
			if len(m.filter.Solution) > 0 {
				m.filter.Solution = m.filter.Solution[:len(m.filter.Solution)-1]
			}
		} else {
			m.filter.Solution += input
		}
	}
}

func (m model) updateSearchResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		m.cursor = 0
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.searchResults)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m model) updateEntry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		m.cursor = 0
	case "enter":
		if m.entryStep == entryStepConfirm {
			SaveErrorReport(m.currentReport)
			m.message = "Error report saved successfully!"
			m.state = stateMenu
			m.cursor = 0
		} else {
			m.state = stateEntryField
			m.currentText = m.getCurrentFieldText()
			m.textLines = strings.Split(m.currentText, "\n")
			if len(m.textLines) == 0 {
				m.textLines = []string{""}
			}
			m.textCursor = 0
			m.charCursor = 0
		}
	case "tab":
		if m.entryStep < entryStepConfirm {
			m.entryStep++
		}
	case "shift+tab":
		if m.entryStep > entryStepSymptom {
			m.entryStep--
		}
	}
	return m, nil
}

func (m model) getCurrentFieldText() string {
	switch m.entryStep {
	case entryStepSymptom:
		return m.currentReport.Symptom
	case entryStepProgram:
		return m.currentReport.Program
	case entryStepProgramVersion:
		return m.currentReport.ProgramVersion
	case entryStepDistro:
		return m.currentReport.Distro
	case entryStepDistroVersion:
		return m.currentReport.DistroVersion
	case entryStepResources:
		return strings.Join(m.currentReport.Resources, "\n")
	case entryStepSolution:
		return m.currentReport.Solution
	}
	return ""
}

func (m *model) setCurrentFieldText(text string) {
	switch m.entryStep {
	case entryStepSymptom:
		m.currentReport.Symptom = text
	case entryStepProgram:
		m.currentReport.Program = text
	case entryStepProgramVersion:
		m.currentReport.ProgramVersion = text
	case entryStepDistro:
		m.currentReport.Distro = text
	case entryStepDistroVersion:
		m.currentReport.DistroVersion = text
	case entryStepResources:
		lines := strings.Split(text, "\n")
		resources := []string{}
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				resources = append(resources, strings.TrimSpace(line))
			}
		}
		m.currentReport.Resources = resources
	case entryStepSolution:
		m.currentReport.Solution = text
	}
}

func (m model) updateEntryField(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateEntry
	case "ctrl+s":
		text := strings.Join(m.textLines, "\n")
		m.setCurrentFieldText(text)
		m.state = stateEntry
	case "ctrl+c":
		// Copy current line to clipboard
		if m.textCursor < len(m.textLines) {
			m.clipboard = m.textLines[m.textCursor]
			m.copyToSystemClipboard(m.clipboard)
		}
	case "ctrl+v":
		// Paste from clipboard
		clipText := m.getFromSystemClipboard()
		if clipText != "" {
			m.clipboard = clipText
		}
		if m.clipboard != "" {
			// Insert clipboard content at current cursor position
			currentLine := m.textLines[m.textCursor]
			newLine := currentLine[:m.charCursor] + m.clipboard + currentLine[m.charCursor:]
			m.textLines[m.textCursor] = newLine
			m.charCursor += len(m.clipboard)
		}
	case "enter":
		// Split current line at cursor position
		currentLine := m.textLines[m.textCursor]
		leftPart := currentLine[:m.charCursor]
		rightPart := currentLine[m.charCursor:]
		m.textLines[m.textCursor] = leftPart
		m.textLines = append(m.textLines[:m.textCursor+1], m.textLines[m.textCursor:]...)
		m.textLines[m.textCursor+1] = rightPart
		m.textCursor++
		m.charCursor = 0
	case "up":
		if m.textCursor > 0 {
			m.textCursor--
			// Keep char cursor within bounds of new line
			if m.charCursor > len(m.textLines[m.textCursor]) {
				m.charCursor = len(m.textLines[m.textCursor])
			}
		}
	case "down":
		if m.textCursor < len(m.textLines)-1 {
			m.textCursor++
			// Keep char cursor within bounds of new line
			if m.charCursor > len(m.textLines[m.textCursor]) {
				m.charCursor = len(m.textLines[m.textCursor])
			}
		}
	case "left":
		if m.charCursor > 0 {
			m.charCursor--
		} else if m.textCursor > 0 {
			// Move to end of previous line
			m.textCursor--
			m.charCursor = len(m.textLines[m.textCursor])
		}
	case "right":
		if m.charCursor < len(m.textLines[m.textCursor]) {
			m.charCursor++
		} else if m.textCursor < len(m.textLines)-1 {
			// Move to start of next line
			m.textCursor++
			m.charCursor = 0
		}
	case "home":
		m.charCursor = 0
	case "end":
		m.charCursor = len(m.textLines[m.textCursor])
	case "backspace":
		if m.charCursor > 0 {
			// Remove character before cursor
			currentLine := m.textLines[m.textCursor]
			m.textLines[m.textCursor] = currentLine[:m.charCursor-1] + currentLine[m.charCursor:]
			m.charCursor--
		} else if m.textCursor > 0 {
			// Join with previous line
			prevLine := m.textLines[m.textCursor-1]
			currentLine := m.textLines[m.textCursor]
			m.textLines[m.textCursor-1] = prevLine + currentLine
			m.textLines = append(m.textLines[:m.textCursor], m.textLines[m.textCursor+1:]...)
			m.textCursor--
			m.charCursor = len(prevLine)
		}
	case "delete":
		if m.charCursor < len(m.textLines[m.textCursor]) {
			// Remove character at cursor
			currentLine := m.textLines[m.textCursor]
			m.textLines[m.textCursor] = currentLine[:m.charCursor] + currentLine[m.charCursor+1:]
		} else if m.textCursor < len(m.textLines)-1 {
			// Join with next line
			currentLine := m.textLines[m.textCursor]
			nextLine := m.textLines[m.textCursor+1]
			m.textLines[m.textCursor] = currentLine + nextLine
			m.textLines = append(m.textLines[:m.textCursor+1], m.textLines[m.textCursor+2:]...)
		}
	default:
		if len(msg.String()) == 1 {
			// Insert character at cursor position
			currentLine := m.textLines[m.textCursor]
			m.textLines[m.textCursor] = currentLine[:m.charCursor] + msg.String() + currentLine[m.charCursor:]
			m.charCursor++
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		return m.viewMenu()
	case stateSearch:
		return m.viewSearch()
	case stateSearchResults:
		return m.viewSearchResults()
	case stateEntry:
		return m.viewEntry()
	case stateEntryField:
		return m.viewEntryField()
	}
	return ""
}

func (m model) viewMenu() string {
	s := "Error Report Manager\n\n"
	
	if m.message != "" {
		s += fmt.Sprintf("✓ %s\n\n", m.message)
	}
	
	options := []string{
		"Search Error Reports",
		"Enter New Error Report",
	}
	
	for i, option := range options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}
	
	s += "\nPress q to quit"
	return s
}

func (m model) viewSearch() string {
	s := "Search Error Reports\n\n"
	
	fields := []struct {
		label string
		value string
		step  searchStep
	}{
		{"General Query", m.filter.Q, searchStepQuery},
		{"Symptom", m.filter.Symptom, searchStepSymptom},
		{"Program", m.filter.Program, searchStepProgram},
		{"Program Version", m.filter.ProgramVersion, searchStepProgramVersion},
		{"Distro", m.filter.Distro, searchStepDistro},
		{"Distro Version", m.filter.DistroVersion, searchStepDistroVersion},
		{"Solution", m.filter.Solution, searchStepSolution},
	}
	
	for _, field := range fields {
		cursor := " "
		if m.searchStep == field.step {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, field.label, field.value)
	}
	
	cursor := " "
	if m.searchStep == searchStepExecute {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Execute Search\n", cursor)
	
	s += "\nPress Enter to select, Tab/Shift+Tab to navigate, Esc to go back"
	return s
}

func (m model) viewSearchResults() string {
	s := "Search Results\n\n"
	
	if len(m.searchResults) == 0 {
		s += "No results found"
	} else {
		for i, result := range m.searchResults {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s - %s\n", cursor, result.Program, result.Symptom)
		}
		
		if len(m.searchResults) > 0 && m.cursor < len(m.searchResults) {
			selected := m.searchResults[m.cursor]
			s += fmt.Sprintf("\n--- Details ---\n")
			s += fmt.Sprintf("Date: %s\n", selected.Date.Format("2006-01-02"))
			s += fmt.Sprintf("Program: %s %s\n", selected.Program, selected.ProgramVersion)
			s += fmt.Sprintf("Distro: %s %s\n", selected.Distro, selected.DistroVersion)
			s += fmt.Sprintf("Symptom: %s\n", selected.Symptom)
			if len(selected.Resources) > 0 {
				s += fmt.Sprintf("Resources: %s\n", strings.Join(selected.Resources, ", "))
			}
			s += fmt.Sprintf("Solution: %s\n", selected.Solution)
		}
	}
	
	s += "\nPress Esc to go back"
	return s
}

func (m model) viewEntry() string {
	s := "Enter New Error Report\n\n"
	
	fields := []struct {
		label string
		value string
		step  entryStep
	}{
		{"Symptom", m.currentReport.Symptom, entryStepSymptom},
		{"Program", m.currentReport.Program, entryStepProgram},
		{"Program Version", m.currentReport.ProgramVersion, entryStepProgramVersion},
		{"Distro", m.currentReport.Distro, entryStepDistro},
		{"Distro Version", m.currentReport.DistroVersion, entryStepDistroVersion},
		{"Resources", strings.Join(m.currentReport.Resources, ", "), entryStepResources},
		{"Solution", m.currentReport.Solution, entryStepSolution},
	}
	
	for _, field := range fields {
		cursor := " "
		if m.entryStep == field.step {
			cursor = ">"
		}
		displayValue := field.value
		if len(displayValue) > 50 {
			displayValue = displayValue[:50] + "..."
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, field.label, displayValue)
	}
	
	cursor := " "
	if m.entryStep == entryStepConfirm {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Save Report\n", cursor)
	
	s += "\nPress Enter to edit field, Tab/Shift+Tab to navigate, Esc to go back"
	return s
}

func (m model) viewEntryField() string {
	fieldName := m.getCurrentFieldName()
	s := fmt.Sprintf("Edit %s\n\n", fieldName)
	
	const lineWidth = 70 // Maximum line width before wrapping
	
	for i, line := range m.textLines {
		// Wrap long lines for display
		wrappedLines := wrapLine(line, lineWidth)
		
		for j, wrappedLine := range wrappedLines {
			lineCursor := " "
			if i == m.textCursor {
				if j == 0 {
					lineCursor = ">"
				} else {
					lineCursor = "|"
				}
				// Show cursor position within the line
				if j == 0 && m.charCursor <= len(wrappedLine) {
					// Insert cursor marker
					if m.charCursor == len(wrappedLine) {
						wrappedLine += "█"
					} else {
						wrappedLine = wrappedLine[:m.charCursor] + "█" + wrappedLine[m.charCursor:]
					}
				}
			}
			s += fmt.Sprintf("%s %s\n", lineCursor, wrappedLine)
		}
	}
	
	s += "\nPress Ctrl+S to save, Esc to cancel, Enter for new line"
	s += "\nArrow keys to navigate, Ctrl+C to copy line, Ctrl+V to paste"
	return s
}

func (m model) getCurrentFieldName() string {
	switch m.entryStep {
	case entryStepSymptom:
		return "Symptom"
	case entryStepProgram:
		return "Program"
	case entryStepProgramVersion:
		return "Program Version"
	case entryStepDistro:
		return "Distro"
	case entryStepDistroVersion:
		return "Distro Version"
	case entryStepResources:
		return "Resources (one per line)"
	case entryStepSolution:
		return "Solution"
	}
	return ""
}

// copyToSystemClipboard copies text to system clipboard
func (m *model) copyToSystemClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return // Unsupported OS
	}
	
	cmd.Stdin = strings.NewReader(text)
	cmd.Run() // Ignore errors for now
}

// getFromSystemClipboard gets text from system clipboard
func (m *model) getFromSystemClipboard() string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows":
		cmd = exec.Command("powershell", "-command", "Get-Clipboard")
	default:
		return "" // Unsupported OS
	}
	
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// wrapLine wraps a line to fit within the specified width
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	
	if len(line) <= width {
		return []string{line}
	}
	
	var wrapped []string
	for len(line) > width {
		// Try to break at word boundary
		breakPoint := width
		for i := width - 1; i >= 0; i-- {
			if line[i] == ' ' {
				breakPoint = i
				break
			}
		}
		
		wrapped = append(wrapped, strings.TrimSpace(line[:breakPoint]))
		line = strings.TrimSpace(line[breakPoint:])
	}
	
	if len(line) > 0 {
		wrapped = append(wrapped, line)
	}
	
	return wrapped
}

func main() {
	initIndex := flag.Bool("init-index", false, "Initialize Meilisearch index with proper attributes")
	flag.Parse()
	
	if *initIndex {
		logToFile("Initializing Meilisearch index...\n")
		if err := InitializeIndexIfNeeded(); err != nil {
			logToFile("Error initializing index: %v\n", err)
			os.Exit(1)
		}
		logToFile("Index initialized successfully!\n")
		return
	}
	
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		logToFile("Error: %v", err)
		os.Exit(1)
	}
}
