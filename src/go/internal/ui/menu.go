package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("#00A86B"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	logo = `
▜   ▗ ▗     ▌      ▌    ▜ ▘
▐ █▌▜▘▜▘█▌▛▘▛▌▛▌▚▘▛▌▄▖▛▘▐ ▌
▐▖▙▖▐▖▐▖▙▖▌ ▙▌▙▌▞▖▙▌  ▙▖▐▖▌
                           `

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A86B")).
			Bold(true)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(str))
}

type MenuModel struct {
	list      list.Model
	Choice    string
	Quitting  bool
	termWidth int
}

func NewMenuModel() MenuModel {
	items := []list.Item{
		item("View a person's profile"),
		item("Search a movie"),
		item("View Lists of Letterboxd"),
		item("Get Watchlist"),
		item("Get diary of a person"),
	}

	const defaultWidth = 40
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "🎬 Letterboxd CLI — Choose an option"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return MenuModel{list: l, termWidth: defaultWidth}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.termWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.Choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MenuModel) View() string {
	logoStr := logoStyle.Width(m.termWidth).Align(lipgloss.Center).Render(logo)

	if m.Choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("You selected: %s", m.Choice))
	}
	if m.Quitting {
		return quitTextStyle.Render("Goodbye! See you on Letterboxd 👋")
	}

	return logoStr + "\n\n" + m.list.View()
}
