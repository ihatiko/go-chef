package build_project_ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ihatiko/go-chef/constants"
)

type (
	errMsg error
)

const (
	projectName = iota
	projectPath
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type CreateProjectModel struct {
	inputs                []textinput.Model
	focused               int
	err                   error
	FilledInputProcessing func(args []string)
	titleStyle            lipgloss.Style
}

func InitialModel(filledInputProcessing func(args []string), titleStyle lipgloss.Style) *CreateProjectModel {
	var inputs = make([]textinput.Model, 2)
	inputs[projectName] = textinput.New()
	inputs[projectName].Placeholder = "awesomeProject1"
	inputs[projectName].Focus()
	inputs[projectName].CharLimit = 500
	inputs[projectName].Width = 30
	inputs[projectName].Prompt = ""

	inputs[projectPath] = textinput.New()
	inputs[projectPath].Placeholder = "/path/desired/here"
	inputs[projectPath].Width = 30
	inputs[projectPath].Prompt = ""

	return &CreateProjectModel{
		inputs:                inputs,
		focused:               0,
		err:                   nil,
		FilledInputProcessing: filledInputProcessing,
		titleStyle:            titleStyle,
	}
}

func (m *CreateProjectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CreateProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				var values []string
				values = append(values, fmt.Sprintf("--project_path=%s", m.inputs[projectPath].Value()))
				values = append(values, fmt.Sprintf("--project_name=%s", m.inputs[projectName].Value()))
				m.FilledInputProcessing(values)
				return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m *CreateProjectModel) View() string {
	return fmt.Sprintf(
		`
 %s

 %s
 %s

 %s  
 %s

 %s  
`,
		m.titleStyle.Render(constants.ProjectTitle),
		inputStyle.Width(30).Render("Project name"),
		m.inputs[projectName].View(),
		inputStyle.Width(30).Render("Project path"),
		m.inputs[projectPath].View(),
		continueStyle.Render("Press enter to continue"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *CreateProjectModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *CreateProjectModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
