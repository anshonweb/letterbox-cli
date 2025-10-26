package main

import (
	"fmt"
	"os"

	"github.com/anshonweb/letterbox-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.NewRootModel())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
