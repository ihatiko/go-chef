package user

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ugm/internal/tui/common"
)

func (bu BubbleUser) Update(msg tea.Msg) (BubbleUser, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		horizontal, vertical := common.ListStyle.GetFrameSize()
		paginatorHeight := lipgloss.Height(bu.list.Paginator.View())

		bu.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
		bu.viewport = viewport.New(msg.Width, msg.Height)
		bu.viewport.SetContent(bu.detailView())
	}

	var cmd tea.Cmd
	bu.list, cmd = bu.list.Update(msg)

	return bu, cmd
}
