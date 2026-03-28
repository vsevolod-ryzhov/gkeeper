package main

import (
	"fmt"
	"gkeeper/internal/config"
	"gkeeper/internal/grpcclient"
	"gkeeper/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

func main() {
	config.ParseFlags()

	logger := zap.Must(zap.NewProduction())
	client := grpcclient.NewClient(logger)
	defer client.Close()

	p := tea.NewProgram(
		tui.NewMainModel(client),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
