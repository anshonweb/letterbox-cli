package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	userHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00A86B")).
			MarginBottom(1)

	userLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Width(18)

	userStatValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229"))

	userBioStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("245")).
			MarginTop(1).
			Padding(0, 1)

	userFavoriteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				PaddingLeft(2)

	userRecentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			PaddingLeft(2)

	userSocialHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.Color("#00A86B")).
				MarginBottom(1)

	userSocialListStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	userReviewMovieStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229"))

	userReviewTextStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Italic(true).
				Foreground(lipgloss.Color("245"))

	userReviewRatingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	userPageTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				Margin(1, 0, 1, 0)

	userQueryLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	userHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
)

type UserReview struct {
	MovieName  string  `json:"movie_name"`
	MovieYear  int     `json:"movie_year"`
	Rating     float64 `json:"rating"`
	ReviewText string  `json:"review_text"`
	ReviewDate string  `json:"review_date"`
}

type UserDetails struct {
	Username     string       `json:"username"`
	FilmsWatched int          `json:"films_watched"`
	Bio          string       `json:"bio"`
	Following    []string     `json:"following"`
	Followers    []string     `json:"followers"`
	Favorites    []string     `json:"favorites"`
	LastWatched  string       `json:"last_watched"`
	Reviews      []UserReview `json:"reviews"`
	This_year    int          `json:"this_year"`
	Recent       []string     `json:"recent"`
	Website      string       `json:"website"`
	Location     string       `json:"location"`
}

type userDetailsResultMsg struct {
	details UserDetails
	err     error
}

type UserModel struct {
	input           textinput.Model
	spinner         spinner.Model
	paginator       paginator.Model
	socialPaginator paginator.Model
	loading         bool
	submitted       bool
	viewing         bool
	quitting        bool
	width           int
	err             error
	tabs            []string
	activeTab       int
	userDetails     UserDetails
}

func NewUserModel() UserModel {
	ti := textinput.New()
	ti.Placeholder = "Enter a Letterboxd username..."
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30
	ti.Prompt = ""

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A86B"))

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 5
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("•")

	socialP := paginator.New()
	socialP.Type = paginator.Dots
	socialP.PerPage = 5
	socialP.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Render("•")
	socialP.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("•")

	return UserModel{
		input:           ti,
		spinner:         sp,
		paginator:       p,
		socialPaginator: socialP,
		tabs:            []string{"Profile", "Favorites", "Recent", "Reviews", "Social"},
	}
}

func callPythonGetUserDetails(username string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("python3", "../../../python/scripts/user_details.py", username)
		out, err := cmd.Output()
		if err != nil {
			return userDetailsResultMsg{err: err}
		}

		var details UserDetails
		if err := json.Unmarshal(out, &details); err != nil {
			return userDetailsResultMsg{err: err}
		}
		return userDetailsResultMsg{details: details}
	}
}

func (m UserModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m UserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.viewing {
				m.viewing = false
				m.submitted = false
				m.userDetails = UserDetails{}
				m.input.Focus()
				return m, textinput.Blink
			}
			return NewMenuModel(), nil
		case "enter":
			if !m.submitted {
				m.submitted = true
				m.loading = true
				username := m.input.Value()
				cmds = append(cmds, m.spinner.Tick, callPythonGetUserDetails(username))
			}
		case "tab":
			if m.viewing {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
			}
		case "shift+tab":
			if m.viewing {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = len(m.tabs) - 1
				}
			}
		case "right", "l":
			if m.viewing {
				if m.activeTab <= 2 { // Profile, Favorites, Recent
					m.activeTab = (m.activeTab + 1) % len(m.tabs)
				}
			}
		case "left", "h":
			if m.viewing {
				if m.activeTab <= 2 { // Profile, Favorites, Recent
					m.activeTab--
					if m.activeTab < 0 {
						m.activeTab = len(m.tabs) - 1
					}
				}
			}
		}

	case userDetailsResultMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.userDetails = msg.details

		sort.Slice(m.userDetails.Reviews, func(i, j int) bool {
			t1, _ := time.Parse("2006-01-02", m.userDetails.Reviews[i].ReviewDate)
			t2, _ := time.Parse("2006-01-02", m.userDetails.Reviews[j].ReviewDate)
			return t1.After(t2)
		})

		m.paginator.SetTotalPages(len(m.userDetails.Reviews))
		m.paginator.Page = 0
		maxSocialItems := max(len(m.userDetails.Following), len(m.userDetails.Followers))
		m.socialPaginator.SetTotalPages(maxSocialItems)
		m.socialPaginator.Page = 0

		m.loading = false
		m.viewing = true

	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !m.viewing {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.viewing {
		switch m.activeTab {
		case 3: // Reviews
			m.paginator, cmd = m.paginator.Update(msg)
			cmds = append(cmds, cmd)
		case 4: // Social
			m.socialPaginator, cmd = m.socialPaginator.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m UserModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if m.loading {
		return fmt.Sprintf("\n\n   %s Fetching profile for '%s'...\n\n", m.spinner.View(), m.input.Value())
	}

	if m.viewing {
		var renderedTabs []string
		for i, t := range m.tabs {
			style := navStyle
			if i == m.activeTab {
				style = navActiveStyle
			}
			renderedTabs = append(renderedTabs, style.Render("→ "+t))
		}
		tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

		var content string
		switch m.activeTab {
		case 0:
			content = m.renderProfileTab()
		case 1:
			content = m.renderFavoritesTab()
		case 2:
			content = m.renderRecentTab()
		case 3:
			content = m.renderReviewsTab()
		case 4:
			content = m.renderSocialTab()
		}

		helpText := "\n(Use Tab to switch tabs, ESC to go back)"
		if m.activeTab <= 2 {
			helpText = "\n(Use ←/→ or Tab to switch tabs, ESC to go back)"
		} else {
			helpText = "\n(Use ←/→ to change page, Tab to switch tabs, ESC to go back)"
		}

		return SearchBorderBox.Render(lipgloss.JoinVertical(lipgloss.Left, tabsRow, "", content)) + helpText
	}

	// Updated input view
	title := userPageTitleStyle.Render("user profile")
	queryLabel := userQueryLabelStyle.Render("username:")
	inputBlock := lipgloss.JoinVertical(lipgloss.Left,
		queryLabel,
		m.input.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(strings.Repeat("─", m.input.Width+1)),
	)
	help := userHelpStyle.Render("type a username and press enter")

	final := lipgloss.JoinVertical(lipgloss.Left,
		title,
		inputBlock,
		"\n\n\n",
		help,
	)
	return lipgloss.NewStyle().Margin(1, 2).Render(final)
}

func (m UserModel) renderProfileTab() string {
	header := userHeaderStyle.Render(fmt.Sprintf("@%s", m.userDetails.Username))
	bio := userBioStyle.Render(m.userDetails.Bio)

	stats := []string{
		lipgloss.JoinHorizontal(lipgloss.Left, userLabelStyle.Render("Films Watched:"), userStatValueStyle.Render(fmt.Sprintf("%d", m.userDetails.FilmsWatched))),
		lipgloss.JoinHorizontal(lipgloss.Left, userLabelStyle.Render("Last Watched:"), userStatValueStyle.Render(m.userDetails.LastWatched)),
		lipgloss.JoinHorizontal(lipgloss.Left, userLabelStyle.Render("This Year: "), userStatValueStyle.Render(fmt.Sprintf("%d", m.userDetails.This_year))),
	}

	if m.userDetails.Website != "" {
		stats = append(stats, lipgloss.JoinHorizontal(lipgloss.Left, userLabelStyle.Render("Website: "), userStatValueStyle.Render(m.userDetails.Website)))
	}
	if m.userDetails.Location != "" {
		stats = append(stats, lipgloss.JoinHorizontal(lipgloss.Left, userLabelStyle.Render("Location: "), userStatValueStyle.Render(m.userDetails.Location)))
	}

	statsBlock := lipgloss.JoinVertical(lipgloss.Left, stats...)
	return lipgloss.JoinVertical(lipgloss.Left, header, statsBlock, bio)
}

func (m UserModel) renderFavoritesTab() string {
	if len(m.userDetails.Favorites) == 0 {
		return "No favorite films listed."
	}
	var favs []string
	for _, f := range m.userDetails.Favorites {
		favs = append(favs, userFavoriteStyle.Render("♥︎ "+f))
	}
	return lipgloss.JoinVertical(lipgloss.Left, favs...)
}

func (m UserModel) renderRecentTab() string {
	if len(m.userDetails.Recent) == 0 {
		return "No recent activity found."
	}

	displayCount := min(5, len(m.userDetails.Recent))
	recentToDisplay := m.userDetails.Recent[:displayCount]

	var recentItems []string
	for _, movie := range recentToDisplay {
		recentItems = append(recentItems, userRecentStyle.Render("• "+movie))
	}

	return strings.Join(recentItems, "\n")
}

func (m UserModel) renderReviewsTab() string {
	if len(m.userDetails.Reviews) == 0 {
		return "No reviews found."
	}

	var reviewBlocks []string
	start, end := m.paginator.GetSliceBounds(len(m.userDetails.Reviews))
	paginatedReviews := m.userDetails.Reviews[start:end]

	for _, r := range paginatedReviews {
		movieHeader := userReviewMovieStyle.Render(fmt.Sprintf("%s (%d)", r.MovieName, r.MovieYear))
		rating := userReviewRatingStyle.Render(fmt.Sprintf("★ %.1f/5", r.Rating/2.0))

		headerLine := lipgloss.JoinHorizontal(lipgloss.Top, movieHeader, " ", rating)
		reviewText := userReviewTextStyle.Render(r.ReviewText)

		reviewBlocks = append(reviewBlocks, lipgloss.JoinVertical(lipgloss.Left, headerLine, reviewText))
	}

	paginatorView := m.paginator.View()
	if len(reviewBlocks) > 0 {
		paginatorView = "\n\n" + paginatorView
	}

	return strings.Join(reviewBlocks, "\n\n") + paginatorView
}

func (m UserModel) renderSocialTab() string {
	maxItems := max(len(m.userDetails.Following), len(m.userDetails.Followers))
	if maxItems == 0 {
		return "No social information available."
	}

	start, end := m.socialPaginator.GetSliceBounds(maxItems)

	var paginatedFollowing []string
	if start < len(m.userDetails.Following) {
		paginatedFollowing = m.userDetails.Following[start:min(end, len(m.userDetails.Following))]
	}
	var followingList []string
	for _, f := range paginatedFollowing {
		followingList = append(followingList, "• "+f)
	}
	followingHeader := userSocialHeaderStyle.Render("Following")
	followingBlock := lipgloss.JoinVertical(lipgloss.Left, followingHeader, userSocialListStyle.Render(strings.Join(followingList, "\n")))

	var paginatedFollowers []string
	if start < len(m.userDetails.Followers) {
		paginatedFollowers = m.userDetails.Followers[start:min(end, len(m.userDetails.Followers))]
	}
	var followersList []string
	for _, f := range paginatedFollowers {
		followersList = append(followersList, "• "+f)
	}
	followersHeader := userSocialHeaderStyle.Render("Followers")
	followersBlock := lipgloss.JoinVertical(lipgloss.Left, followersHeader, userSocialListStyle.Render(strings.Join(followersList, "\n")))

	socialContent := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.width/2-5).Render(followingBlock),
		lipgloss.NewStyle().Width(m.width/2-5).Render(followersBlock),
	)

	paginatorView := "\n\n" + m.socialPaginator.View()

	return lipgloss.JoinVertical(lipgloss.Left, socialContent, paginatorView)
}
