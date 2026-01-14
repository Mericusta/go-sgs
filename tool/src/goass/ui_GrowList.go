package goass

import (
	"fmt"

	"github.com/Mericusta/go-sgs/tool/src/common"
	"github.com/Mericusta/go-sgs/tool/src/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type ui_GrowList struct {
	*ui.List

	debugData *data_Debug
	logicData *data_Logic

	title   string
	content string
}

func NewUIGrowList(debugData *data_Debug, logicData *data_Logic) *ui_GrowList {
	m := &ui_GrowList{
		List:      ui.NewList(0),
		debugData: debugData,
		logicData: logicData,
		title:     "Grow List",
	}
	uiGrowAttributes := NewUIGrowAttributes(debugData, logicData)
	uiGrowSkills := NewUIGrowSkillList(debugData, logicData)
	m.AddChoices(uiGrowAttributes, uiGrowSkills)
	return m
}

func (m *ui_GrowList) Init() tea.Cmd {
	return nil
}

func (m *ui_GrowList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds tea.Cmd
	_list, _cmd := m.List.Update(msg)
	m.List, cmds = _list.(*ui.List), tea.Batch(_cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, cmds
}

func (m ui_GrowList) View() string {
	s := fmt.Sprintf("Grow List: %v\n\n", m.GetChosenIndex())
	s += m.List.View()
	return s
}

func (m ui_GrowList) ChoiceTitle() string {
	return m.title
}
