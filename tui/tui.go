package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type Choice string

const (
	Yes   Choice = "Yes"
	Later Choice = "Later"
)

type Model struct {
	Cursor  int
	Choice  Choice
	Title   string
	Choices []Choice
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.Choice = m.Choices[m.Cursor]
			return m, tea.Quit

		case "down", "j":
			m.Cursor++
			if m.Cursor >= len(m.Choices) {
				m.Cursor = 0
			}

		case "up", "k":
			m.Cursor--
			if m.Cursor < 0 {
				m.Cursor = len(m.Choices) - 1
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("%s%s", m.Title, "\n\n"))

	for i := 0; i < len(m.Choices); i++ {
		if m.Cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(string(m.Choices[i]))
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n\n")

	return s.String()
}
