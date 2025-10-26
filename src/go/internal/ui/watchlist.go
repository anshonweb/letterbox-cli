package ui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	watchlistPageTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true).
				Margin(1, 0, 1, 0)

	watchInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7"))

	watchInputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	watchInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	watchHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	exportStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				MarginTop(1)
)

type watchlistResultMsg struct {
	movies []Movie
	err    error
}

type exportResultMsg struct {
	filePath string
	err      error
}

type WatchlistModel struct {
	input               textinput.Model
	spinner             spinner.Model
	table               table.Model
	showSpinner         bool
	showTable           bool
	submitted           bool
	quitting            bool
	err                 error
	promptingExportPath bool
	exportInput         textinput.Model
	exportPath          string
	exportErr           error
	lastExportMsg       time.Time
	watchlist           []Movie
	targetUser          string
	baseStyle           lipgloss.Style
}

func NewWatchlistModel() WatchlistModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Letterboxd username..."
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 40
	ti.Prompt = "Username: "
	ti.PromptStyle = watchInputPromptStyle
	ti.Cursor.Style = watchInputCursorStyle
	ti.TextStyle = watchInputTextStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B"))

	exportTi := textinput.New()
	exportTi.Placeholder = "e.g., my_watchlist.csv or exports/watchlist.csv"
	exportTi.CharLimit = 256
	exportTi.Width = 60
	exportTi.Prompt = "Export Path (relative): "
	exportTi.PromptStyle = watchInputPromptStyle.Copy()
	exportTi.Cursor.Style = watchInputCursorStyle.Copy()
	exportTi.TextStyle = watchInputTextStyle.Copy()

	return WatchlistModel{
		input:       ti,
		spinner:     sp,
		exportInput: exportTi,
		baseStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonGetWatchlist(username string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("python3", "../../../python/scripts/get_watchlist.py", username)
		out, err := cmd.Output()
		if err != nil {
			var errData map[string]string
			if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
				return watchlistResultMsg{err: fmt.Errorf(errData["error"])}
			}
			return watchlistResultMsg{err: fmt.Errorf("failed to run script: %w, output: %s", err, string(out))}
		}

		var maybeErr map[string]string
		if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
			return watchlistResultMsg{err: fmt.Errorf(maybeErr["error"])}
		}

		var movies []Movie
		if err := json.Unmarshal(out, &movies); err != nil {
			return watchlistResultMsg{err: fmt.Errorf("failed to parse watchlist JSON: %w", err)}
		}

		return watchlistResultMsg{movies: movies}
	}
}

func exportWatchlistToCSV(watchlist []Movie, username, relativeFilePath string) tea.Cmd {
	return func() tea.Msg {
		filePath, err := filepath.Abs(relativeFilePath)
		if err != nil {
			return exportResultMsg{err: fmt.Errorf("invalid path format: %w", err)}
		}

		dir := filepath.Dir(filePath)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				if os.IsPermission(err) {
					return exportResultMsg{err: fmt.Errorf("permission denied creating directory %s", dir)}
				}
				return exportResultMsg{err: fmt.Errorf("failed to create directory %s: %w", dir, err)}
			}
		}

		file, err := os.Create(filePath)
		if err != nil {
			if os.IsPermission(err) {
				return exportResultMsg{err: fmt.Errorf("permission denied creating file %s", filePath)}
			}
			return exportResultMsg{err: fmt.Errorf("failed to create file %s: %w", filePath, err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		header := []string{"Title", "Year", "Director"}
		if err := writer.Write(header); err != nil {
			return exportResultMsg{err: fmt.Errorf("failed to write CSV header: %w", err)}
		}

		for _, movie := range watchlist {
			row := []string{movie.Title, fmt.Sprintf("%d", movie.Year), movie.Director}
			if err := writer.Write(row); err != nil {
				fmt.Printf("Error writing row for %s: %v\n", movie.Title, err)
			}
		}

		return exportResultMsg{filePath: filePath}
	}
}

func (m WatchlistModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m WatchlistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				cmds = append(cmds, exportWatchlistToCSV(m.watchlist, m.targetUser, path))
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
			if m.showTable {
				m.showTable = false
				m.submitted = false
				m.input.Focus()
				m.exportPath = ""
				m.exportErr = nil
				return m, nil
			} else {
				return NewMenuModel(), nil
			}

		case "enter":
			if !m.submitted {
				m.submitted = true
				m.showSpinner = true
				m.targetUser = m.input.Value()
				cmds = append(cmds, m.spinner.Tick, callPythonGetWatchlist(m.targetUser))
			}

		case "e":
			if m.showTable && len(m.watchlist) > 0 {
				m.promptingExportPath = true
				m.exportInput.Focus()
				safeUsername := strings.ReplaceAll(strings.ReplaceAll(m.targetUser, "/", "_"), " ", "_")
				defaultFileName := fmt.Sprintf("exports/watchlist_%s.csv", safeUsername)
				m.exportInput.SetValue(defaultFileName)
				return m, textinput.Blink
			}
		}

	case watchlistResultMsg:
		m.showSpinner = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.showTable = true
			m.watchlist = msg.movies

			rows := []table.Row{}
			for _, movie := range m.watchlist {
				rows = append(rows, table.Row{
					movie.Title,
					fmt.Sprintf("%d", movie.Year),
					movie.Director,
				})
			}

			columns := []table.Column{
				{Title: "Title", Width: 40},
				{Title: "Year", Width: 6},
				{Title: "Director", Width: 25},
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(15),
			)

			s := table.DefaultStyles()
			s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
			s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("#00A86B"))
			t.SetStyles(s)

			m.table = t
		}

	case exportResultMsg:
		m.exportErr = msg.err
		m.exportPath = msg.filePath
		m.lastExportMsg = time.Now()

	case tea.WindowSizeMsg:
		if m.showTable {
			m.table.SetWidth(msg.Width - 4)
		}
		m.exportInput.Width = msg.Width - 20
	}

	if m.showSpinner {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !m.showTable {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m WatchlistModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n(Press 'esc' to go back)", m.err)
	}

	if m.promptingExportPath {
		return lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("Exporting watchlist for: %s", m.targetUser),
			exportInputStyle.Render(m.exportInput.View()),
			"\n(Enter path relative to current dir. Press Enter to confirm, Esc to cancel)",
		)
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Fetching watchlist for '%s'...\n\n", m.spinner.View(), m.targetUser)
	}

	if m.showTable {
		exportMsg := ""
		if time.Since(m.lastExportMsg) < 5*time.Second {
			if m.exportErr != nil {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("Export failed: %v", m.exportErr))
			} else if m.exportPath != "" {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("Watchlist exported to %s", m.exportPath))
			}
		}

		view := m.baseStyle.Render(m.table.View()) + "\n(Use ↑/↓ to scroll, 'e' to export, Esc to go back)"
		if exportMsg != "" {
			view += "\n" + exportMsg
		}
		return view
	}

	title := watchlistPageTitleStyle.Render("View Watchlist")
	inputBlock := lipgloss.JoinVertical(lipgloss.Left,
		m.input.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strings.Repeat("─", m.input.Width+len(m.input.Prompt))),
	)
	help := watchHelpStyle.Render("type a username and press enter")

	final := lipgloss.JoinVertical(lipgloss.Left,
		title,
		inputBlock,
		"\n\n\n",
		help,
	)
	return lipgloss.NewStyle().Margin(1, 2).Render(final)
}
