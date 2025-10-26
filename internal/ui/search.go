package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	navStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	navActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4FC3F7")).
			Underline(true)

	SearchBorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 2).
			Margin(1, 2)
)

var (
	movieTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A86B")).
			Bold(true).
			MarginBottom(1)

	movieSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	movieRatingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	movieDetailKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7")).
				Bold(true).
				Width(10)

	movieDetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229"))

	movieTaglineStyle = lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("245")).
				Margin(0, 0, 1, 0)

	movieSmallMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				MarginBottom(1)

	statContainerStyle = lipgloss.NewStyle().
				Margin(1, 0)

	movieStatNumberBlueStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#4FC3F7"))

	movieStatNumberOrangeStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FF9800"))

	movieStatNumberGreenStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#00A86B"))

	movieStatLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	synopsisHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true)

	synopsisBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				PaddingLeft(1)

	synopsisLineStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#00A86B")).
				PaddingLeft(1)

	movieReleaseDateStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				MarginTop(1)

	movieAuthorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true)

	similarMovieStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229"))

	searchInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4FC3F7"))

	searchInputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	searchInputTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B"))

	searchPageTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00A86B")).
				Bold(true).
				Margin(1, 0, 1, 0)

	searchHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	docStyle = lipgloss.NewStyle().Padding(0, 1)
)

type Movie struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Slug     string `json:"slug"`
	Director string `json:"director"`
}
type Review struct {
	Author string  `json:"author"`
	Text   string  `json:"text"`
	Rating float64 `json:"rating"`
}

type Provider struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Link string `json:"link"`
}

type SimilarMovie struct {
	Name   string  `json:"name"`
	Rating float64 `json:"rating"`
}

type MovieDetails struct {
	Title       string         `json:"title"`
	Year        int            `json:"year"`
	Director    string         `json:"director"`
	Genres      []string       `json:"genres"`
	Rating      float64        `json:"rating"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	Reviews     []Review       `json:"reviews"`
	Runtime     string         `json:"runtime"`
	Providers   []Provider     `json:"providers"`
	Cast        []string       `json:"cast"`
	Tagline     string         `json:"tagline"`
	Members     int            `json:"members"`
	Fans        int            `json:"fans"`
	Likes       int            `json:"likes"`
	ReviewCount int            `json:"review_count"`
	Lists       int            `json:"lists"`
	Similar     []SimilarMovie `json:"similar"`
}

type searchResultMsg struct {
	movies []Movie
	err    error
}

type detailsResultMsg struct {
	details MovieDetails
	err     error
}

type SearchModel struct {
	input            textinput.Model
	spinner          spinner.Model
	similarPaginator paginator.Model
	showSpinner      bool
	table            table.Model
	showTable        bool
	submitted        bool
	quitting         bool
	viewingDetails   bool
	loadingDetails   bool
	selectedMovie    Movie
	movieDetails     MovieDetails
	movies           []Movie
	baseStyle        lipgloss.Style
	tabs             []string
	activeTab        int
	width            int
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter movie name..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	ti.Prompt = "Query: "
	ti.PromptStyle = searchInputPromptStyle
	ti.Cursor.Style = searchInputCursorStyle
	ti.TextStyle = searchInputTextStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B"))

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 5
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("•")

	return SearchModel{
		input:            ti,
		spinner:          sp,
		similarPaginator: p,
		baseStyle:        lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

func callPythonSearch(query string) ([]Movie, error) {
	pyExecName := "search_movie"
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
			return nil, fmt.Errorf("fatal: could not get executable path: %w", err)
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
			return nil, fmt.Errorf("python executable not found at %s or %s",
				filepath.Join("$SNAP or ExecDir", "py_execs", pyExecName),
				filepath.Join("project_root", "dist_py", osDir, pyExecName))
		}
	}

	cmd := exec.Command(pyExecPath, query)
	out, err := cmd.Output()

	if err != nil {
		var errData map[string]string
		if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
			return nil, fmt.Errorf(errData["error"])
		}
		return nil, fmt.Errorf("failed to run script '%s': %w, output: %s", pyExecPath, err, string(out))
	}
	var maybeErr map[string]string
	if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
		return nil, fmt.Errorf(maybeErr["error"])
	}
	var movies []Movie
	if err := json.Unmarshal(out, &movies); err != nil {
		return nil, fmt.Errorf("failed to parse movie search JSON: %w", err)
	}
	return movies, nil
}

func callPythonGetDetails(slug string) (MovieDetails, error) {
	pyExecName := "get_movie_details"
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
			return MovieDetails{}, fmt.Errorf("fatal: could not get executable path: %w", err)
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
			return MovieDetails{}, fmt.Errorf("python executable not found at %s or %s",
				filepath.Join("$SNAP or ExecDir", "py_execs", pyExecName),
				filepath.Join("project_root", "dist_py", osDir, pyExecName))
		}
	}

	cmd := exec.Command(pyExecPath, slug)
	out, err := cmd.Output()

	if err != nil {
		var errData map[string]string
		if json.Unmarshal(out, &errData) == nil && errData["error"] != "" {
			return MovieDetails{}, fmt.Errorf(errData["error"])
		}
		return MovieDetails{}, fmt.Errorf("failed to run script '%s': %w, output: %s", pyExecPath, err, string(out))
	}
	var maybeErr map[string]string
	if json.Unmarshal(out, &maybeErr) == nil && maybeErr["error"] != "" {
		return MovieDetails{}, fmt.Errorf(maybeErr["error"])
	}
	var details MovieDetails
	if err := json.Unmarshal(out, &details); err != nil {
		return MovieDetails{}, fmt.Errorf("failed to parse movie details JSON: %w", err)
	}
	return details, nil
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			if m.loadingDetails {
				return m, nil
			}
			if m.viewingDetails {
				m.viewingDetails = false
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
				cmds = append(cmds, m.spinner.Tick, func() tea.Msg {
					movies, err := callPythonSearch(m.input.Value())
					return searchResultMsg{movies, err}
				})
			} else if m.showTable && !m.viewingDetails {
				cursor := m.table.Cursor()
				if len(m.movies) > cursor {
					m.selectedMovie = m.movies[cursor]
					m.loadingDetails = true
					m.showSpinner = true
					cmds = append(cmds, m.spinner.Tick, func() tea.Msg {
						details, err := callPythonGetDetails(m.selectedMovie.Slug)
						return detailsResultMsg{details, err}
					})
				}
			}

		case "left", "h":
			if m.viewingDetails && m.activeTab != 2 {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = len(m.tabs) - 1
				}
				return m, nil
			}

		case "right", "l":
			if m.viewingDetails && m.activeTab != 2 {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, nil
			}

		case "shift+tab":
			if m.viewingDetails {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = len(m.tabs) - 1
				}
				return m, nil
			}

		case "tab":
			if m.viewingDetails {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, nil
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
				movie.Director,
			})
		}

		columns := []table.Column{
			{Title: "No", Width: 4},
			{Title: "Title", Width: 40},
			{Title: "Year", Width: 6},
			{Title: "Director", Width: 25},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(min(len(rows)+1, 15)),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("#00A86B"))
		t.SetStyles(s)

		m.table = t
		return m, nil

	case detailsResultMsg:
		m.loadingDetails = false
		m.showSpinner = false
		m.viewingDetails = true
		m.movieDetails = msg.details

		m.tabs = []string{"Information", "Reviews", "Similar", "Where to Watch"}
		m.activeTab = 0

		m.similarPaginator.SetTotalPages(len(m.movieDetails.Similar))
		m.similarPaginator.Page = 0
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.showTable {
			m.table.SetWidth(msg.Width - 4)
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
	}

	if m.viewingDetails && m.activeTab == 2 {
		m.similarPaginator, cmd = m.similarPaginator.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func formatLargeNumber(n int) string {
	if n > 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000.0)
	}
	if n > 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000.0)
	}
	return fmt.Sprintf("%d", n)
}

func (m SearchModel) renderMovieInfo() string {
	d := m.movieDetails

	hyperlink := fmt.Sprintf("\x1b]8;;%s\x07%s\x1b]8;;\x07", d.URL, d.Title)
	title := movieTitleStyle.Render(hyperlink)
	subtitle := movieSubtitleStyle.Render(fmt.Sprintf("%d • %s", d.Year, movieRatingStyle.Render(fmt.Sprintf("★ %.1f/5", d.Rating))))
	tagline := movieTaglineStyle.Render(fmt.Sprintf(`"%s"`, d.Tagline))
	header := lipgloss.JoinVertical(lipgloss.Left, title, subtitle, tagline)

	smallMeta := movieSmallMetaStyle.Render(
		fmt.Sprintf("• %d   • %s   • %s", d.Year, d.Runtime, strings.Join(d.Genres, ", ")),
	)

	statMembers := lipgloss.JoinVertical(lipgloss.Center,
		movieStatNumberBlueStyle.Render(formatLargeNumber(d.Members)),
		movieStatLabelStyle.Render("members"),
	)
	statFans := lipgloss.JoinVertical(lipgloss.Center,
		movieStatNumberOrangeStyle.Render(formatLargeNumber(d.Fans)),
		movieStatLabelStyle.Render("fans"),
	)
	statLikes := lipgloss.JoinVertical(lipgloss.Center,
		movieStatNumberGreenStyle.Render(formatLargeNumber(d.Likes)),
		movieStatLabelStyle.Render("likes"),
	)
	statReviews := lipgloss.JoinVertical(lipgloss.Center,
		movieStatNumberBlueStyle.Render(formatLargeNumber(d.ReviewCount)),
		movieStatLabelStyle.Render("reviews"),
	)
	statLists := lipgloss.JoinVertical(lipgloss.Center,
		movieStatNumberOrangeStyle.Render(formatLargeNumber(d.Lists)),
		movieStatLabelStyle.Render("lists"),
	)

	statsBlock := statContainerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			statMembers,
			"   ",
			statFans,
			"   ",
			statLikes,
			"   ",
			statReviews,
			"   ",
			statLists,
		),
	)

	directorLine := lipgloss.JoinHorizontal(lipgloss.Left,
		movieDetailKeyStyle.Render("director"),
		movieDetailValueStyle.Render(d.Director),
	)
	castLine := lipgloss.JoinHorizontal(lipgloss.Left,
		movieDetailKeyStyle.Render("cast"),
		movieDetailValueStyle.Render(strings.Join(d.Cast, ", ")),
	)
	detailsBlock := lipgloss.JoinVertical(lipgloss.Left, directorLine, castLine)

	synopsisHeader := synopsisHeaderStyle.Render("synopsis")
	synopsisBody := synopsisBodyStyle.Render(d.Description)
	synopsisBlock := lipgloss.JoinVertical(lipgloss.Left,
		synopsisHeader,
		lipgloss.JoinHorizontal(lipgloss.Left,
			synopsisLineStyle.Render(""),
			synopsisBody,
		),
	)

	finalRender := lipgloss.JoinVertical(lipgloss.Left,
		header,
		smallMeta,
		statsBlock,
		"",
		detailsBlock,
		"",
		synopsisBlock,
		"",
	)

	return docStyle.Render(finalRender)
}

func (m SearchModel) renderMovieReviews() string {
	d := m.movieDetails
	if len(d.Reviews) == 0 {
		return "No reviews available."
	}
	lines := []string{}
	for _, r := range d.Reviews {
		coloredAuthor := movieAuthorStyle.Render(r.Author)
		coloredRating := movieRatingStyle.Render(fmt.Sprintf("★ %.1f/5", r.Rating))
		lines = append(lines, fmt.Sprintf("• %s %s\n%s", coloredAuthor, coloredRating, r.Text))
	}
	return strings.Join(lines, "\n\n")
}

func (m SearchModel) renderSimilarTab() string {
	if len(m.movieDetails.Similar) == 0 {
		return "No similar movies found."
	}

	var similarBlocks []string
	start, end := m.similarPaginator.GetSliceBounds(len(m.movieDetails.Similar))
	paginatedSimilar := m.movieDetails.Similar[start:end]

	for _, s := range paginatedSimilar {
		movieHeader := similarMovieStyle.Render(s.Name)
		rating := movieRatingStyle.Render(fmt.Sprintf("★ %.1f/5", s.Rating))
		line := lipgloss.JoinHorizontal(lipgloss.Top, movieHeader, " ", rating)
		similarBlocks = append(similarBlocks, line)
	}

	paginatorView := m.similarPaginator.View()
	if len(similarBlocks) > 0 {
		paginatorView = "\n\n" + paginatorView
	}

	return strings.Join(similarBlocks, "\n\n") + paginatorView
}

func (m SearchModel) renderProviders() string {
	p := m.movieDetails.Providers
	if len(p) == 0 {
		return "No streaming or purchase options available."
	}

	var lines []string
	for _, pr := range p {
		line := fmt.Sprintf("• %s (%s)\n  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Render(pr.Name),
			pr.Type,
			pr.Link)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n\n")
}

func (m SearchModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}

	if m.loadingDetails {
		return fmt.Sprintf("\n\n   %s Fetching details for '%s'...\n\n", m.spinner.View(), m.selectedMovie.Title)
	}

	if m.viewingDetails {
		var renderedTabs []string
		for i, t := range m.tabs {
			var style lipgloss.Style
			if i == m.activeTab {
				style = navActiveStyle
			} else {
				style = navStyle
			}
			renderedTabs = append(renderedTabs, style.Render("→ "+t))
		}
		tabsRow := strings.Join(renderedTabs, "  ")

		var content string
		switch m.activeTab {
		case 0:
			content = m.renderMovieInfo()
		case 1:
			content = m.renderMovieReviews()
		case 2:
			content = m.renderSimilarTab()
		case 3:
			content = m.renderProviders()
		}

		full := fmt.Sprintf("%s\n\n%s", tabsRow, content)

		helpText := "\n(Use ←/→ to switch tabs, ESC to go back)"
		if m.activeTab == 2 {
			helpText = "\n(Use ←/→ to change page, Tab to switch tabs, ESC to go back)"
		}

		return SearchBorderBox.Render(full) + helpText
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Searching for '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if !m.showTable {
		title := searchPageTitleStyle.Render("Search Movies")
		inputBlock := lipgloss.JoinVertical(lipgloss.Left,
			m.input.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strings.Repeat("─", m.input.Width+len(m.input.Prompt))),
		)
		help := searchHelpStyle.Render("type a movie name and press enter")

		final := lipgloss.JoinVertical(lipgloss.Left,
			title,
			inputBlock,
			"\n\n\n",
			help,
		)
		return lipgloss.NewStyle().Margin(1, 2).Render(final)
	}

	return m.baseStyle.Render(m.table.View()) + "\n(Enter to view details, Esc to go back)"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return a
	}
	return b
}
