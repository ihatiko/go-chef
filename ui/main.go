package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"ugm/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
