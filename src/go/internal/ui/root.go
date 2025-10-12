package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	current tea.Model
}

func NewRootModel() RootModel {
	return RootModel{current: NewMenuModel()}
}

func (m RootModel) Init() tea.Cmd {
	return m.current.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.current.Update(msg)

	switch typed := newModel.(type) {
	case MenuModel:
		if typed.Choice == "Search a movie" {
			return RootModel{current: NewSearchModel()}, nil
		}
		return RootModel{current: typed}, cmd

	case SearchModel:
		return RootModel{current: typed}, cmd
	}

	return m, cmd
}

func (m RootModel) View() string {
	return m.current.View()
}
