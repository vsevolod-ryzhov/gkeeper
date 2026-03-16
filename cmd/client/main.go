package main

import (
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config.ParseFlags()
	p := tea.NewProgram(
		tui.NewMainModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
