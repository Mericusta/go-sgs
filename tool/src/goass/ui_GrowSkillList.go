package goass

import (
	"fmt"

	"github.com/Mericusta/go-sgs/tool/src/ui"

	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type Skill int

const (
	SKILL_1 Skill = iota + 1
	SKILL_2
)

func (a Skill) String() string {
	switch a {
	case SKILL_1:
		return "SKILL_1"
	case SKILL_2:
		return "SKILL_2"
	}
	return "[Skill]"
}

type ui_GrowSkillList struct {
	*ui.List

	debugData *data_Debug
	logicData *data_Logic

	title   string
	content string
}

func NewUIGrowSkillList(debugData *data_Debug, logicData *data_Logic) *ui_GrowSkillList {
	m := &ui_GrowSkillList{
		List:      ui.NewList(0),
		debugData: debugData,
		logicData: logicData,
		title:     "Skills",
		content:   "this is grow skills model",
	}
	uiGrowSkill1 := NewUIGrowSkill(debugData, logicData, SKILL_1)
	uiGrowSkill2 := NewUIGrowSkill(debugData, logicData, SKILL_2)
	m.AddChoices(uiGrowSkill1, uiGrowSkill2)
	return m
}

func (m *ui_GrowSkillList) Init() tea.Cmd {
	return nil
}

func (m *ui_GrowSkillList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *ui_GrowSkillList) View() string {
	s := fmt.Sprintf("Skill List: %v\n\n", m.GetChosenIndex())
	s += m.List.View()
	return s
}

func (m *ui_GrowSkillList) ChoiceTitle() string {
	return m.title
}
