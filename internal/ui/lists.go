package ui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	exportInputStyle = lipgloss.NewStyle().MarginTop(1)
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

type exportListResultMsg struct {
	filePath string
	err      error
}

type ListsModel struct {
	input               textinput.Model
	spinner             spinner.Model
	table               table.Model
	showSpinner         bool
	showTable           bool
	submitted           bool
	quitting            bool
	viewingDetails      bool
	loadingDetails      bool
	promptingExportPath bool
	exportInput         textinput.Model
	exportPath          string
	exportErr           error
	lastExportMsg       time.Time
	selectedList        ListSearchResult
	listDetails         []Movie
	detailsTable        table.Model
	lists               []ListSearchResult
	err                 error
	baseStyle           lipgloss.Style
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

	exportTi := textinput.New()
	exportTi.Placeholder = "e.g., my_list_export.csv or exports/my_list.csv"
	exportTi.CharLimit = 256
	exportTi.Width = 60
	exportTi.Prompt = "Export Path (relative): "
	exportTi.PromptStyle = listInputPromptStyle.Copy()
	exportTi.Cursor.Style = listInputCursorStyle.Copy()
	exportTi.TextStyle = listInputTextStyle.Copy()

	return ListsModel{
		input:       ti,
		spinner:     sp,
		exportInput: exportTi,
		baseStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonSearchLists(query string) tea.Cmd {
	return func() tea.Msg {
		pyExecName := "search_lists"
		if runtime.GOOS == "windows" {
			pyExecName += ".exe"
		}

		baseDir := ""
		snapDir := os.Getenv("SNAP")
		if snapDir != "" {
			baseDir = snapDir
		} else {
			goExecPath, err := os.Executable()
			if err != nil {
				return searchListsResultMsg{err: fmt.Errorf("fatal: could not get executable path: %w", err)}
			}
			baseDir = filepath.Dir(goExecPath)
		}

		pyExecPath := filepath.Join(baseDir, "py_execs", pyExecName)

		if _, err := os.Stat(pyExecPath); os.IsNotExist(err) {
			wd, _ := os.Getwd()
			arch := runtime.GOARCH
			osDir := runtime.GOOS + "_" + arch
			altPyExecPath := filepath.Join(wd, "..", "..", "dist_py", osDir, pyExecName)

			if _, altErr := os.Stat(altPyExecPath); !os.IsNotExist(altErr) {
				pyExecPath = altPyExecPath
			} else {
				return searchListsResultMsg{err: fmt.Errorf("python executable not found at %s or %s",
					filepath.Join("$SNAP or ExecDir", "py_execs", pyExecName),
					filepath.Join("project_root", "dist_py", osDir, pyExecName))}
			}
		}

		cmd := exec.Command(pyExecPath, query)
		out, err := cmd.Output()

		if err != nil {
			var errData map[string]string
			if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
				return searchListsResultMsg{err: fmt.Errorf(errData["error"])}
			}
			return searchListsResultMsg{err: fmt.Errorf("failed to run script '%s': %w, output: %s", pyExecPath, err, string(out))}
		}
		var maybeErr map[string]string
		if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
			return searchListsResultMsg{err: fmt.Errorf(maybeErr["error"])}
		}
		var lists []ListSearchResult
		if err := json.Unmarshal(out, &lists); err != nil {
			return searchListsResultMsg{err: fmt.Errorf("failed to parse list search JSON: %w", err)}
		}
		return searchListsResultMsg{lists: lists}
	}
}

func callPythonGetListDetails(owner, slug string) tea.Cmd {
	return func() tea.Msg {
		pyExecName := "get_list_details"
		if runtime.GOOS == "windows" {
			pyExecName += ".exe"
		}

		baseDir := ""
		snapDir := os.Getenv("SNAP")
		if snapDir != "" {
			baseDir = snapDir
		} else {
			goExecPath, err := os.Executable()
			if err != nil {
				return listDetailsResultMsg{err: fmt.Errorf("fatal: could not get executable path: %w", err)}
			}
			baseDir = filepath.Dir(goExecPath)
		}

		pyExecPath := filepath.Join(baseDir, "py_execs", pyExecName)

		if _, err := os.Stat(pyExecPath); os.IsNotExist(err) {
			wd, _ := os.Getwd()
			arch := runtime.GOARCH
			osDir := runtime.GOOS + "_" + arch
			altPyExecPath := filepath.Join(wd, "..", "..", "dist_py", osDir, pyExecName)

			if _, altErr := os.Stat(altPyExecPath); !os.IsNotExist(altErr) {
				pyExecPath = altPyExecPath
			} else {
				return listDetailsResultMsg{err: fmt.Errorf("python executable not found at %s or %s",
					filepath.Join("$SNAP or ExecDir", "py_execs", pyExecName),
					filepath.Join("project_root", "dist_py", osDir, pyExecName))}
			}
		}

		cmd := exec.Command(pyExecPath, owner, slug)
		out, err := cmd.Output()

		if err != nil {
			var errData map[string]string
			if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
				return listDetailsResultMsg{err: fmt.Errorf(errData["error"])}
			}
			return listDetailsResultMsg{err: fmt.Errorf("failed to run script '%s': %w, output: %s", pyExecPath, err, string(out))}
		}
		var maybeErr map[string]string
		if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
			return listDetailsResultMsg{err: fmt.Errorf(maybeErr["error"])}
		}
		var movies []Movie
		if err := json.Unmarshal(out, &movies); err != nil {
			return listDetailsResultMsg{err: fmt.Errorf("failed to parse list details JSON: %w", err)}
		}
		return listDetailsResultMsg{movies: movies}
	}
}
func exportListToCSV(movies []Movie, listName, owner, relativeFilePath string) tea.Cmd {
	return func() tea.Msg {
		filePath, err := filepath.Abs(relativeFilePath)
		if err != nil {
			return exportListResultMsg{err: fmt.Errorf("invalid path format: %w", err)}
		}

		dir := filepath.Dir(filePath)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				if os.IsPermission(err) {
					return exportListResultMsg{err: fmt.Errorf("permission denied creating directory %s", dir)}
				}
				return exportListResultMsg{err: fmt.Errorf("failed to create directory %s: %w", dir, err)}
			}
		}

		file, err := os.Create(filePath)
		if err != nil {
			if os.IsPermission(err) {
				return exportListResultMsg{err: fmt.Errorf("permission denied creating file %s", filePath)}
			}
			return exportListResultMsg{err: fmt.Errorf("failed to create file %s: %w", filePath, err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		header := []string{"Title", "Year"}
		if err := writer.Write(header); err != nil {
			return exportListResultMsg{err: fmt.Errorf("failed to write CSV header: %w", err)}
		}

		for _, movie := range movies {
			row := []string{movie.Title, fmt.Sprintf("%d", movie.Year)}
			if err := writer.Write(row); err != nil {
				fmt.Printf("Error writing row for %s: %v\n", movie.Title, err)
			}
		}

		return exportListResultMsg{filePath: filePath}
	}
}

func (m ListsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ListsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.promptingExportPath {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				m.promptingExportPath = false
				m.exportInput.Blur()
				m.exportInput.Reset()
				m.exportErr = nil
				return m, nil
			case "enter":
				m.promptingExportPath = false
				m.exportInput.Blur()
				path := m.exportInput.Value()
				m.exportInput.Reset()
				m.exportPath = ""
				m.exportErr = nil
				cmds = append(cmds, exportListToCSV(m.listDetails, m.selectedList.Name, m.selectedList.Owner, path))
				return m, tea.Batch(cmds...)
			}
		}
		m.exportInput, cmd = m.exportInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

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
				m.exportPath = ""
				m.exportErr = nil
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

		case "e":
			if m.viewingDetails && len(m.listDetails) > 0 {
				m.promptingExportPath = true
				m.exportInput.Focus()
				safeListName := strings.ReplaceAll(strings.ReplaceAll(m.selectedList.Name, "/", "_"), " ", "_")
				safeOwner := strings.ReplaceAll(strings.ReplaceAll(m.selectedList.Owner, "/", "_"), " ", "_")
				defaultFileName := fmt.Sprintf("exports/list_%s_%s.csv", safeOwner, safeListName)
				m.exportInput.SetValue(defaultFileName)
				return m, textinput.Blink
			}

		}

	case searchListsResultMsg:
		m.showSpinner = false
		if msg.err != nil {
			m.err = msg.err
		} else {
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
		}
		return m, nil

	case listDetailsResultMsg:
		m.loadingDetails = false
		m.showSpinner = false
		if msg.err != nil {
			m.err = msg.err
		} else {
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
		}
		return m, nil

	case exportListResultMsg:
		m.exportErr = msg.err
		m.exportPath = msg.filePath
		m.lastExportMsg = time.Now()
		return m, nil

	case tea.WindowSizeMsg:
		if m.showTable {
			m.table.SetWidth(msg.Width - 4)
		}
		if m.viewingDetails {
			m.detailsTable.SetWidth(msg.Width - 4)
		}
		m.exportInput.Width = msg.Width - 20
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

	if m.promptingExportPath {
		return lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("Exporting list: %s by %s", m.selectedList.Name, m.selectedList.Owner),
			exportInputStyle.Render(m.exportInput.View()),
			"\n(Enter path and press Enter to confirm, Esc to cancel)",
		)
	}

	if m.loadingDetails {
		return fmt.Sprintf("\n\n   %s Fetching details for '%s'...\n\n", m.spinner.View(), m.selectedList.Name)
	}
	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Searching for lists matching '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if m.viewingDetails {
		title := listPageTitleStyle.Render(fmt.Sprintf("Movies in: %s", m.selectedList.Name))
		exportMsg := ""
		if time.Since(m.lastExportMsg) < 5*time.Second {
			if m.exportErr != nil {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("Export failed: %v", m.exportErr))
			} else if m.exportPath != "" {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("List exported to %s", m.exportPath))
			}
		}

		viewContent := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Margin(1, 0).Render(title),
			m.baseStyle.Render(m.detailsTable.View()),
			"\n(Use ↑/↓ to navigate, 'e' to export, Esc to go back)",
		)
		if exportMsg != "" {
			viewContent += "\n" + exportMsg
		}
		return viewContent
	}

	if m.showTable {
		return m.baseStyle.Render(m.table.View()) + "\n(Use ↑/↓ to scroll, Enter to select, Esc to go back)"
	}

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
