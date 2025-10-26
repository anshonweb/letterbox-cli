package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	menuAppHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 1)

	menuLogoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A86B")).
			Bold(true).
			Margin(1, 0)

	menuSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("245"))

	menuSelectedItemOrangeStyle = lipgloss.NewStyle().
					PaddingLeft(0).
					Foreground(lipgloss.Color("#FF9800"))

	menuSelectedItemBlueStyle = lipgloss.NewStyle().
					PaddingLeft(0).
					Foreground(lipgloss.Color("#4FC3F7"))

	menuSelectedItemGreenStyle = lipgloss.NewStyle().
					PaddingLeft(0).
					Foreground(lipgloss.Color("#00A86B"))

	menuShortcutStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	menuBottomHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

var logo = `
 ██╗     ███████╗████████╗████████╗███████╗██████╗  ██████╗██╗     ██╗
 ██║     ██╔════╝╚══██╔══╝╚══██╔══╝██╔════╝██╔══██╗██╔════╝██║     ██║
 ██║     █████╗     ██║      ██║   █████╗  ██████╔╝██║     ██║     ██║
 ██║     ██╔══╝     ██║      ██║   ██╔══╝  ██╔══██╗██║     ██║     ██║
 ███████╗███████╗   ██║      ██║   ███████╗██║  ██║╚██████╗███████╗██║
 ╚══════╝╚══════╝   ╚═╝      ╚═╝   ╚══════╝╚═╝  ╚═╝ ╚═════╝╚══════╝╚═╝
`

type item string

func (i item) FilterValue() string { return string(i) }

type menuItemDelegate struct{}

func (d menuItemDelegate) Height() int                             { return 1 }
func (d menuItemDelegate) Spacing() int                            { return 0 }
func (d menuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d menuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	shortcut := menuShortcutStyle.Render(fmt.Sprintf("[%d]", index+1))
	paddingNeeded := m.Width() - lipgloss.Width(string(i)) - lipgloss.Width(shortcut) - 4
	padding := strings.Repeat(" ", max(1, paddingNeeded))

	itemText := string(i)
	fullItemStr := itemText + padding + shortcut

	if index == m.Index() {
		var selectedStyle lipgloss.Style
		colorIndex := index % 4
		switch colorIndex {
		case 0, 3:
			selectedStyle = menuSelectedItemOrangeStyle
		case 1:
			selectedStyle = menuSelectedItemBlueStyle
		case 2:
			selectedStyle = menuSelectedItemGreenStyle
		default:
			selectedStyle = menuSelectedItemOrangeStyle
		}
		fmt.Fprint(w, selectedStyle.Render("▸ "+fullItemStr))
	} else {
		fmt.Fprint(w, menuItemStyle.Render("  "+fullItemStr))
	}
}

type MenuModel struct {
	list      list.Model
	Choice    string
	Quitting  bool
	termWidth int
}

func NewMenuModel() MenuModel {
	items := []list.Item{
		item("user profile"),
		item("search movie"),
		item("diary"),
		item("watchlist"),
		item("view lists"),
	}

	const defaultWidth = 50
	listHeight := len(items)

	l := list.New(items, menuItemDelegate{}, defaultWidth, listHeight)

	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return MenuModel{list: l, termWidth: 80}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.list.SetSize(msg.Width-4, len(m.list.Items()))
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "1", "2", "3", "4", "5":
			index := int(keypress[0] - '1')
			if index < len(m.list.Items()) {
				m.list.Select(index)
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch string(i) {
					case "user profile":
						m.Choice = "View a person's profile"
					case "search movie":
						m.Choice = "Search a movie"
					case "diary":
						m.Choice = "Get diary of a person"
					case "watchlist":
						m.Choice = "Get Watchlist"
					case "view lists":
						m.Choice = "View Lists of Letterboxd"
					default:
						m.Choice = ""
					}
					if m.Choice != "" {
						return m, tea.Quit
					}
				}
				return m, nil
			}

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch string(i) {
				case "user profile":
					m.Choice = "View a person's profile"
				case "search movie":
					m.Choice = "Search a movie"
				case "diary":
					m.Choice = "Get diary of a person"
				case "watchlist":
					m.Choice = "Get Watchlist"
				case "view lists":
					m.Choice = "View Lists of Letterboxd"
				default:
					m.Choice = ""
				}
			}
			if m.Choice != "" {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	if !m.Quitting && m.Choice == "" {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m MenuModel) View() string {
	if m.Choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected: %s", m.Choice))
	}
	if m.Quitting {
		return quitTextStyle.Render("Exiting LetterCLI...")
	}

	logoRender := menuLogoStyle.Render(logo)
	subtitleRender := menuSubtitleStyle.Render("a terminal client for letterboxd")
	listRender := m.list.View()
	bottomHelp := menuBottomHelpStyle.Render("press [1-5] or click to select")

	finalView := lipgloss.JoinVertical(lipgloss.Left,
		logoRender,
		subtitleRender,
		"\n\n",
		listRender,
		"\n\n\n",
		bottomHelp,
	)

	return lipgloss.NewStyle().Margin(1, 2).Render(finalView)
}
