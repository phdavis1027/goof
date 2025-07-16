package main

import (
	"flag"
	"fmt"
	"os"
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
	
	// UI state
	message       string
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
	case "enter":
		// currentLine := m.textLines[m.textCursor]
		m.textLines = append(m.textLines[:m.textCursor+1], m.textLines[m.textCursor:]...)
		m.textLines[m.textCursor+1] = ""
		m.textCursor++
	case "up":
		if m.textCursor > 0 {
			m.textCursor--
		}
	case "down":
		if m.textCursor < len(m.textLines)-1 {
			m.textCursor++
		}
	case "backspace":
		if len(m.textLines[m.textCursor]) > 0 {
			m.textLines[m.textCursor] = m.textLines[m.textCursor][:len(m.textLines[m.textCursor])-1]
		} else if m.textCursor > 0 {
			m.textLines = append(m.textLines[:m.textCursor], m.textLines[m.textCursor+1:]...)
			m.textCursor--
		}
	default:
		if len(msg.String()) == 1 {
			m.textLines[m.textCursor] += msg.String()
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
	
	for i, line := range m.textLines {
		cursor := " "
		if i == m.textCursor {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, line)
	}
	
	s += "\nPress Ctrl+S to save, Esc to cancel, Enter for new line, Up/Down to navigate"
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

func main() {
	initIndex := flag.Bool("init-index", false, "Initialize Meilisearch index with proper attributes")
	flag.Parse()
	
	if *initIndex {
		fmt.Println("Initializing Meilisearch index...")
		if err := InitializeIndexIfNeeded(); err != nil {
			fmt.Printf("Error initializing index: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Index initialized successfully!")
		return
	}
	
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
