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

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	diaryPageTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true).
				Margin(1, 0, 1, 0)

	diaryInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7"))

	diaryInputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	diaryInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	diaryHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
)

type DiaryEntry struct {
	Title     string  `json:"title"`
	Year      int     `json:"year"`
	Rating    float64 `json:"rating"`
	WatchDate string  `json:"watch_date"`
	Rewatch   bool    `json:"rewatch"`
	Slug      string  `json:"slug"`
}
type diaryResultMsg struct {
	entries []DiaryEntry
	err     error
}

type exportDiaryResultMsg struct {
	filePath string
	err      error
}

type DiaryModel struct {
	input               textinput.Model
	spinner             spinner.Model
	table               table.Model
	paginator           paginator.Model
	showSpinner         bool
	showDiary           bool
	submitted           bool
	quitting            bool
	err                 error
	promptingExportPath bool
	exportInput         textinput.Model
	exportPath          string
	exportErr           error
	lastExportMsg       time.Time
	diaryEntries        []DiaryEntry
	targetUser          string
	baseStyle           lipgloss.Style
	width               int
}

func NewDiaryModel() DiaryModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Letterboxd username..."
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 40
	ti.Prompt = "Username: "
	ti.PromptStyle = diaryInputPromptStyle
	ti.Cursor.Style = diaryInputCursorStyle
	ti.TextStyle = diaryInputTextStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B"))
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 15
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("•")

	exportTi := textinput.New()
	exportTi.Placeholder = "e.g., my_diary.csv or exports/diary.csv"
	exportTi.CharLimit = 256
	exportTi.Width = 60
	exportTi.Prompt = "Export Path (relative): "
	exportTi.PromptStyle = diaryInputPromptStyle.Copy()
	exportTi.Cursor.Style = diaryInputCursorStyle.Copy()
	exportTi.TextStyle = diaryInputTextStyle.Copy()

	columns := []table.Column{
		{Title: "Watched", Width: 10},
		{Title: "Title", Width: 35},
		{Title: "Year", Width: 6},
		{Title: "Rating", Width: 8},
		{Title: "Rewatch", Width: 7},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(p.PerPage),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)

	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	return DiaryModel{
		input:       ti,
		spinner:     sp,
		paginator:   p,
		table:       t,
		exportInput: exportTi,
		baseStyle:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonGetDiary(username string) tea.Cmd {
	return func() tea.Msg {
		pyExecName := "get_diary"
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
				return diaryResultMsg{err: fmt.Errorf("fatal: could not get executable path: %w", err)}
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
				return diaryResultMsg{err: fmt.Errorf("python executable not found at %s or %s",
					filepath.Join("$SNAP or ExecDir", "py_execs", pyExecName),
					filepath.Join("project_root", "dist_py", osDir, pyExecName))}
			}
		}

		cmd := exec.Command(pyExecPath, username)
		out, err := cmd.Output()

		if err != nil {
			var errData map[string]string
			if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
				return diaryResultMsg{err: fmt.Errorf(errData["error"])}
			}
			return diaryResultMsg{err: fmt.Errorf("failed to run script '%s': %w, output: %s", pyExecPath, err, string(out))}
		}
		var maybeErr map[string]string
		if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
			return diaryResultMsg{err: fmt.Errorf(maybeErr["error"])}
		}
		var entries []DiaryEntry
		if err := json.Unmarshal(out, &entries); err != nil {
			return diaryResultMsg{err: fmt.Errorf("failed to parse diary JSON: %w", err)}
		}
		return diaryResultMsg{entries: entries}
	}
}

func exportDiaryToCSV(entries []DiaryEntry, username, relativeFilePath string) tea.Cmd {
	return func() tea.Msg {
		filePath, err := filepath.Abs(relativeFilePath)
		if err != nil {
			return exportDiaryResultMsg{err: fmt.Errorf("invalid path format: %w", err)}
		}
		dir := filepath.Dir(filePath)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				if os.IsPermission(err) {
					return exportDiaryResultMsg{err: fmt.Errorf("permission denied creating directory %s", dir)}
				}
				return exportDiaryResultMsg{err: fmt.Errorf("failed to create directory %s: %w", dir, err)}
			}
		}
		file, err := os.Create(filePath)
		if err != nil {
			if os.IsPermission(err) {
				return exportDiaryResultMsg{err: fmt.Errorf("permission denied creating file %s", filePath)}
			}
			return exportDiaryResultMsg{err: fmt.Errorf("failed to create file %s: %w", filePath, err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		header := []string{"WatchDate", "Title", "Year", "Rating", "Rewatch"}
		if err := writer.Write(header); err != nil {
			return exportDiaryResultMsg{err: fmt.Errorf("failed to write CSV header: %w", err)}
		}
		for _, entry := range entries {
			rewatchStr := ""
			if entry.Rewatch {
				rewatchStr = "Yes"
			}
			row := []string{
				entry.WatchDate,
				entry.Title,
				fmt.Sprintf("%d", entry.Year),
				fmt.Sprintf("%.1f", entry.Rating),
				rewatchStr,
			}
			if err := writer.Write(row); err != nil {
				fmt.Printf("Error writing diary row for %s: %v\n", entry.Title, err)
			}
		}
		return exportDiaryResultMsg{filePath: filePath}
	}
}

func ratingToStars(rating float64) string {
	stars := ""
	fullStars := int(rating)
	hasHalf := rating-float64(fullStars) >= 0.5
	for i := 0; i < fullStars; i++ {
		stars += "★"
	}
	if hasHalf {
		stars += "½"
	}
	emptyStars := 5 - fullStars
	if hasHalf {
		emptyStars--
	}
	for i := 0; i < emptyStars; i++ {
		stars += "☆"
	}
	if rating <= 0 {
		return strings.Repeat("☆", 5)
	}
	return stars
}

func (m *DiaryModel) updateTableRows() {
	start, end := m.paginator.GetSliceBounds(len(m.diaryEntries))
	pageEntries := m.diaryEntries[start:end]

	rows := make([]table.Row, len(pageEntries))
	for i, entry := range pageEntries {
		rewatchStr := ""
		if entry.Rewatch {
			rewatchStr = "✔"
		}
		rows[i] = table.Row{
			entry.WatchDate,
			entry.Title,
			fmt.Sprintf("%d", entry.Year),
			ratingToStars(entry.Rating),
			rewatchStr,
		}
	}
	m.table.SetRows(rows)
}

func (m DiaryModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m DiaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				cmds = append(cmds, exportDiaryToCSV(m.diaryEntries, m.targetUser, path))
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
				m.showDiary = false
				m.submitted = false
				m.input.Focus()
				return m, nil
			}
			if m.showDiary {
				m.showDiary = false
				m.submitted = false
				m.input.Focus()
				m.exportPath = ""
				m.exportErr = nil
				m.diaryEntries = nil
				return m, nil
			} else {
				return NewMenuModel(), nil
			}
		case "enter":
			if !m.submitted {
				m.submitted = true
				m.showSpinner = true
				m.targetUser = m.input.Value()
				cmds = append(cmds, m.spinner.Tick, callPythonGetDiary(m.targetUser))
			}
		case "e":
			if m.showDiary && len(m.diaryEntries) > 0 {
				m.promptingExportPath = true
				m.exportInput.Focus()
				safeUsername := strings.ReplaceAll(strings.ReplaceAll(m.targetUser, "/", "_"), " ", "_")
				defaultFileName := fmt.Sprintf("exports/diary_%s.csv", safeUsername)
				m.exportInput.SetValue(defaultFileName)
				return m, textinput.Blink
			}
		case "left", "h", "right", "l":
			if !m.showDiary {

				if m.input.Focused() {
					m.input, cmd = m.input.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}

	case diaryResultMsg:
		m.showSpinner = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.showDiary = true
			m.diaryEntries = msg.entries
			m.paginator.SetTotalPages(len(m.diaryEntries))
			m.paginator.Page = 0
			m.updateTableRows()
		}

	case exportDiaryResultMsg:
		m.exportErr = msg.err
		m.exportPath = msg.filePath
		m.lastExportMsg = time.Now()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.table.SetWidth(msg.Width - 4)
		m.exportInput.Width = msg.Width - 20
	}

	if m.showSpinner {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !m.showDiary {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	} else {

		prevPage := m.paginator.Page
		m.paginator, cmd = m.paginator.Update(msg)
		cmds = append(cmds, cmd)

		if m.paginator.Page != prevPage {
			m.updateTableRows()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m DiaryModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n(Press 'esc' to go back)", m.err)
	}

	if m.promptingExportPath {
		return lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("Exporting diary for: %s", m.targetUser),
			exportInputStyle.Render(m.exportInput.View()),
			"\n(Enter path relative to current dir. Press Enter to confirm, Esc to cancel)",
		)
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Fetching diary for '%s'...\n\n", m.spinner.View(), m.targetUser)
	}

	if m.showDiary {
		exportMsg := ""
		if time.Since(m.lastExportMsg) < 5*time.Second {
			if m.exportErr != nil {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("Export failed: %v", m.exportErr))
			} else if m.exportPath != "" {
				exportMsg = exportStatusStyle.Render(fmt.Sprintf("Diary exported to %s", m.exportPath))
			}
		}

		tableRender := m.baseStyle.Render(m.table.View())

		viewContent := lipgloss.JoinVertical(lipgloss.Left,
			tableRender,
			m.paginator.View(),
			"\n(Use ←/→ to change page, 'e' to export, Esc to go back)",
		)

		if exportMsg != "" {
			viewContent += "\n" + exportMsg
		}

		title := diaryPageTitleStyle.Render(fmt.Sprintf("%s's Diary", m.targetUser))
		return lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().Margin(0, 2).Render(title), viewContent)
	}

	title := diaryPageTitleStyle.Render("View Diary")
	inputBlock := lipgloss.JoinVertical(lipgloss.Left,
		m.input.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strings.Repeat("─", m.input.Width+len(m.input.Prompt))),
	)
	help := diaryHelpStyle.Render("type a username and press enter")
	final := lipgloss.JoinVertical(lipgloss.Left, title, inputBlock, "\n\n\n", help)
	return lipgloss.NewStyle().Margin(1, 2).Render(final)
}
