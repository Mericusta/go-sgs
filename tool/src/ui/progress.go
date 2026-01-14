package ui

import (
	"strings"

	"github.com/Mericusta/go-sgs/tool/src/common"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type Progress struct {
	percent  float64        // 百分比 [0,1]
	progress progress.Model // 底层 model
	padding  int            // 左边距
}

func NewProgress(initPercent float64, padding, maxWidth int) *Progress {
	return &Progress{
		percent: initPercent,
		progress: progress.New(
			progress.WithWidth(maxWidth),
		),
		padding: padding,
	}
}

func (m *Progress) SetPercent(deltaPercent float64) { m.percent += deltaPercent }
func (m *Progress) ResetPercent()                   { m.percent = 0 }

func (m *Progress) Init() tea.Cmd { return common.CMD_tick() }

func (m *Progress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case common.MSG_tick:
		m.percent += 0.25
		if m.percent > 1.0 {
			m.percent = 1.0
			return m, common.CMD_back()
		}
		return m, nil
	}
	return m, nil
}

func (m Progress) View() string {
	pad := strings.Repeat(" ", m.padding)
	return "\n" + pad + m.progress.ViewAs(m.percent) + "\n"
}
