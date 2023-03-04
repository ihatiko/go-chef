package ui

import (
	"fmt"
	build_project_ui "github.com/ihatiko/go-chef/ui/build-project"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 10

type Category string

const (
	Project Category = "Project"
	Feature Category = "Feature/Domain -- TODO"
	Main    Category = "Main"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(3)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(1)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                         { return 1 }
func (d itemDelegate) Spacing() int                        { return 0 }
func (d itemDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Bold(true).Foreground(lipgloss.Color("170")).Render("~ " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

type BaseProgramModel struct {
	list    list.Model
	choice  Category
	Windows map[Category]tea.Model
}

func Init(createProjectFunc func(args []string)) *BaseProgramModel {
	items := []list.Item{
		item(Project),
		item(Feature),
	}

	const defaultWidth = 200

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Go Chef golang template generator"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.Styles.Title.Background(lipgloss.Color("170")).Bold(true)
	m := &BaseProgramModel{list: l, choice: Main, Windows: map[Category]tea.Model{
		Project: build_project_ui.InitialModel(createProjectFunc),
	}}
	return m
}
func (m BaseProgramModel) Init() tea.Cmd {
	return nil
}

func (m BaseProgramModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.choice != Main {
		return m.getCurrentView().Update(msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = Category(i)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BaseProgramModel) getCurrentView() tea.Model {
	return m.Windows[m.choice]
}

func (m BaseProgramModel) View() string {
	switch m.choice {
	case Project:
		return m.getCurrentView().View()
	case Feature:
		m.choice = Main
		return "TODO"
	}
	return "\n" + m.list.View()
}

func StartUI(createProjectFunc func(args []string)) {
	if _, err := tea.NewProgram(Init(createProjectFunc)).Run(); err != nil {
		fmt.Println("Error running project:", err)
		os.Exit(1)
	}
}
