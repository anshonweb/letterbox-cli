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
	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00A86B")).
				Padding(0, 1)

	searchSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CDCFE")).
				Padding(0, 1)

	searchBodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1)

	SearchUrlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4FC3F7")).
			Underline(true).
			Padding(0, 1)

	SearchBorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 2).
			Margin(1, 2)
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

type MovieDetails struct {
	Title       string     `json:"title"`
	Year        int        `json:"year"`
	Director    string     `json:"director"`
	Genres      []string   `json:"genres"`
	Rating      float64    `json:"rating"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	Reviews     []Review   `json:"reviews"`
	Providers   []Provider `json:"providers`
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
	input          textinput.Model
	spinner        spinner.Model
	showSpinner    bool
	table          table.Model
	showTable      bool
	submitted      bool
	quitting       bool
	viewingDetails bool
	loadingDetails bool
	selectedMovie  Movie
	movieDetails   MovieDetails
	movies         []Movie
	baseStyle      lipgloss.Style
	tabs           []string
	tabContent     []string
	activeTab      int
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
		input:     ti,
		spinner:   sp,
		baseStyle: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")),
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

func callPythonGetDetails(slug string) (MovieDetails, error) {
	cmd := exec.Command("python3", "../../../python/scripts/get_movie_details.py", slug)
	out, err := cmd.Output()
	if err != nil {
		return MovieDetails{}, err
	}

	var details MovieDetails
	err = json.Unmarshal(out, &details)
	if err != nil {
		return MovieDetails{}, err
	}
	return details, nil
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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
				m.quitting = true
				return m, tea.Quit
			}

		case "enter":
			if !m.submitted {
				m.submitted = true
				m.showSpinner = true
				return m, func() tea.Msg {
					movies, err := callPythonSearch(m.input.Value())
					return searchResultMsg{movies, err}
				}
			} else if m.showTable && !m.viewingDetails {
				cursor := m.table.Cursor()
				if len(m.movies) > cursor {
					m.selectedMovie = m.movies[cursor]
					m.loadingDetails = true
					m.showSpinner = true
					return m, func() tea.Msg {
						details, err := callPythonGetDetails(m.selectedMovie.Slug)
						return detailsResultMsg{details, err}
					}
				}
			}

		case "left", "h", "p":
			if m.viewingDetails {
				m.activeTab = max(m.activeTab-1, 0)
				return m, nil
			}

		case "right", "l", "n", "tab":
			if m.viewingDetails {
				m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
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
			table.WithHeight(len(rows)+3),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
		s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
		t.SetStyles(s)

		m.table = t
		return m, nil

	case detailsResultMsg:
		m.loadingDetails = false
		m.showSpinner = false
		m.viewingDetails = true
		m.movieDetails = msg.details

		m.tabs = []string{"Information", "Reviews", "Where to Watch"}
		m.activeTab = 0
		m.tabContent = []string{
			renderMovieInfo(msg.details),
			renderMovieReviews(msg.details),
			renderProviders(msg.details.Providers),
		}
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
	} else if !m.viewingDetails {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

var ()

func renderMovieInfo(d MovieDetails) string {
	header := searchTitleStyle.Render(fmt.Sprintf("🎬 %s (%d)", d.Title, d.Year))
	info := searchSectionStyle.Render(
		fmt.Sprintf("Director: %s\nGenres: %s\nRating: ★ %.1f/5",
			d.Director,
			strings.Join(d.Genres, ", "),
			d.Rating,
		),
	)
	description := searchBodyStyle.Render(d.Description)
	link := SearchUrlStyle.Render("🔗 " + d.URL)

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", header, info, description, link)
}

var (
	authorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	ratingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
)

func renderMovieReviews(d MovieDetails) string {
	if len(d.Reviews) == 0 {
		return "No reviews available."

	}
	lines := []string{}
	for _, r := range d.Reviews {
		coloredAuthor := authorStyle.Render(r.Author)
		coloredRating := ratingStyle.Render(fmt.Sprintf("★ %.1f/5", r.Rating))
		lines = append(lines, fmt.Sprintf("• %s %s\n%s", coloredAuthor, coloredRating, r.Text))
	}

	return strings.Join(lines, "\n\n")
}

func renderProviders(p []Provider) string {
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
				style = searchTitleStyle.Copy().Underline(true)
			} else {
				style = searchTitleStyle.Copy().Faint(true)
			}
			renderedTabs = append(renderedTabs, style.Render(t))
		}
		tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

		content := m.tabContent[m.activeTab]

		full := fmt.Sprintf("%s\n\n%s", tabsRow, content)
		return SearchBorderBox.Render(full) + "\n(Use Left/Right to switch tabs, ESC to go back)"
	}

	if m.showSpinner {
		return fmt.Sprintf("\n\n   %s Searching for '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if !m.showTable {
		return fmt.Sprintf("🎬 Search a Movie\n\n%s\n\nPress Enter to search, Esc to quit.", m.input.View())
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
	if a > b {
		return a
	}
	return b
}
