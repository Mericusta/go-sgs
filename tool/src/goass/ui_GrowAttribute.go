package goass

import (
	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type ui_GrowAttribute struct {
	debugData *data_Debug
	logicData *data_Logic

	attributionKey Attribution
	title          string
	content        string
}

func NewUIGrowAttribute(debugData *data_Debug, logicData *data_Logic, attributionKey Attribution) *ui_GrowAttribute {
	m := &ui_GrowAttribute{
		debugData:      debugData,
		logicData:      logicData,
		attributionKey: attributionKey,
	}
	m.title = attributionKey.String()
	m.content = "Grow " + m.title
	return m
}

func (m *ui_GrowAttribute) Init() tea.Cmd {
	m.logicData.IncreaseAttributeLevel(m.attributionKey)
	return common.CMD_back()
}

func (m *ui_GrowAttribute) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, nil
}

func (m *ui_GrowAttribute) View() string {
	return m.content
}

func (m *ui_GrowAttribute) ChoiceTitle() string {
	return m.title
}
