package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Movie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

type searchResultMsg struct {
	movies []Movie
	err    error
}

type SearchModel struct {
	input       textinput.Model
	spinner     spinner.Model
	showSpinner bool
	table       table.Model
	showTable   bool
	submitted   bool
	quitting    bool
	movies      []Movie
	baseStyle   lipgloss.Style
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter movie name..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return SearchModel{
		input:       ti,
		spinner:     sp,
		showSpinner: false,
		showTable:   false,
		submitted:   false,
		baseStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonSearch(query string) ([]Movie, error) {
	cmd := exec.Command("python3", "../../../python/scripts/search_movie.py", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var movies []Movie
	err = json.Unmarshal(out, &movies)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if !m.submitted {
				m.submitted = true
				m.showSpinner = true

				// async backend call
				return m, func() tea.Msg {
					movies, err := callPythonSearch(m.input.Value())
					return searchResultMsg{movies, err}
				}
			}
		}

	case searchResultMsg:
		m.showSpinner = false
		m.showTable = true
		m.movies = msg.movies

		rows := []table.Row{}
		for i, movie := range m.movies {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				movie.Title,
				fmt.Sprintf("%d", movie.Year),
			})
		}

		columns := []table.Column{
			{Title: "No", Width: 4},
			{Title: "Title", Width: 30},
			{Title: "Year", Width: 6},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(len(rows)+3),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
		t.SetStyles(s)

		m.table = t
		return m, nil

	case tea.WindowSizeMsg:
		if m.showTable {
			m.table.SetWidth(msg.Width - 4)
		}
	}

	if m.showSpinner {
		m.spinner, cmd = m.spinner.Update(msg)
	} else if !m.showTable {
		m.input, cmd = m.input.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m SearchModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Searching for '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if !m.showTable {
		return fmt.Sprintf("🎬 Search a Movie\n\n%s\n\nPress Enter to search, Esc to quit.", m.input.View())
	}

	return m.baseStyle.Render(m.table.View()) + "\n(Press Esc to quit)"
}
