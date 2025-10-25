package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listPageTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true).
				Margin(1, 0, 1, 0)

	listInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7"))

	listInputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	listInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	listHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
)

type ListSearchResult struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Slug  string `json:"slug"`
}

type searchListsResultMsg struct {
	lists []ListSearchResult
	err   error
}

type listDetailsResultMsg struct {
	movies []Movie
	err    error
}

type ListsModel struct {
	input          textinput.Model
	spinner        spinner.Model
	table          table.Model
	showSpinner    bool
	showTable      bool
	submitted      bool
	quitting       bool
	viewingDetails bool
	loadingDetails bool
	selectedList   ListSearchResult
	listDetails    []Movie
	detailsTable   table.Model
	lists          []ListSearchResult
	err            error
	baseStyle      lipgloss.Style
}

func NewListsModel() ListsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter list search query..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	ti.Prompt = "Query: "
	ti.PromptStyle = listInputPromptStyle
	ti.Cursor.Style = listInputCursorStyle
	ti.TextStyle = listInputTextStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B"))

	return ListsModel{
		input:     ti,
		spinner:   sp,
		baseStyle: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonSearchLists(query string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("python3", "../../../python/scripts/search_lists.py", query)
		out, err := cmd.Output()
		if err != nil {
			return searchListsResultMsg{err: err}
		}

		var lists []ListSearchResult
		if err := json.Unmarshal(out, &lists); err != nil {
			return searchListsResultMsg{err: err}
		}
		return searchListsResultMsg{lists: lists}
	}
}

func callPythonGetListDetails(owner, slug string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("python3", "../../../python/scripts/get_list_details.py", owner, slug)
		out, err := cmd.Output()
		if err != nil {
			return listDetailsResultMsg{err: err}
		}

		var movies []Movie
		if err := json.Unmarshal(out, &movies); err != nil {
			return listDetailsResultMsg{err: err}
		}
		return listDetailsResultMsg{movies: movies}
	}
}

func (m ListsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ListsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.err != nil {
				m.err = nil
				m.showTable = false
				m.submitted = false
				m.input.Focus()
				return m, nil
			}
			if m.loadingDetails {
				return m, nil
			}
			if m.viewingDetails {
				m.viewingDetails = false
				m.listDetails = nil
				return m, nil
			} else if m.showTable {
				m.showTable = false
				m.submitted = false
				m.input.Focus()
				return m, nil
			} else {
				return NewMenuModel(), nil
			}

		case "enter":
			if !m.submitted {
				m.submitted = true
				m.showSpinner = true
				cmds = append(cmds, m.spinner.Tick, callPythonSearchLists(m.input.Value()))
			} else if m.showTable && !m.viewingDetails {
				cursor := m.table.Cursor()
				if len(m.lists) > cursor {
					m.selectedList = m.lists[cursor]
					m.loadingDetails = true
					m.showSpinner = true
					cmds = append(cmds, m.spinner.Tick, callPythonGetListDetails(m.selectedList.Owner, m.selectedList.Slug))
				}
			}
		}

	case searchListsResultMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.showSpinner = false
		m.showTable = true
		m.lists = msg.lists

		rows := []table.Row{}
		for _, l := range m.lists {
			rows = append(rows, table.Row{l.Name, l.Owner})
		}

		columns := []table.Column{
			{Title: "List Name", Width: 40},
			{Title: "Owner", Width: 25},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("#00A86B"))
		t.SetStyles(s)

		m.table = t
		return m, nil

	case listDetailsResultMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.loadingDetails = false
		m.showSpinner = false
		m.viewingDetails = true
		m.listDetails = msg.movies

		rows := []table.Row{}
		for _, movie := range m.listDetails {
			rows = append(rows, table.Row{
				movie.Title,
				fmt.Sprintf("%d", movie.Year),
			})
		}

		columns := []table.Column{
			{Title: "Title", Width: 40},
			{Title: "Year", Width: 6},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(min(len(rows)+1, 20)),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("#00A86B"))
		t.SetStyles(s)

		m.detailsTable = t
		return m, nil

	case tea.WindowSizeMsg:
		if m.showTable {
			m.table.SetWidth(msg.Width - 4)
		}
		if m.viewingDetails {
			m.detailsTable.SetWidth(msg.Width - 4)
		}
	}

	if m.showSpinner {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !m.showTable {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else if !m.viewingDetails {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

	} else {
		m.detailsTable, cmd = m.detailsTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ListsModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n(Press 'esc' to go back)", m.err)
	}

	if m.loadingDetails {
		return fmt.Sprintf("\n\n   %s Fetching details for '%s'...\n\n", m.spinner.View(), m.selectedList.Name)
	}

	if m.viewingDetails {
		title := listPageTitleStyle.Render(fmt.Sprintf("Movies in: %s", m.selectedList.Name))
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Margin(1, 0).Render(title),
			m.baseStyle.Render(m.detailsTable.View()),
			"\n(Use ↑/↓ to navigate, Esc to go back to lists)",
		)
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Searching for lists matching '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if !m.showTable {
		title := listPageTitleStyle.Render("Search Letterboxd Lists")
		inputBlock := lipgloss.JoinVertical(lipgloss.Left,
			m.input.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strings.Repeat("─", m.input.Width+len(m.input.Prompt))),
		)
		help := listHelpStyle.Render("type a query and press enter")

		final := lipgloss.JoinVertical(lipgloss.Left,
			title,
			inputBlock,
			"\n\n\n",
			help,
		)
		return lipgloss.NewStyle().Margin(1, 2).Render(final)
	}

	return m.baseStyle.Render(m.table.View()) + "\n(Use ↑/↓ to scroll, Enter to select, Esc to go back)"
}
