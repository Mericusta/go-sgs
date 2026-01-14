package goass

import (
	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type ui_GrowSkill struct {
	debugData *data_Debug
	logicData *data_Logic

	skillKey Skill
	title    string
	content  string
}

func NewUIGrowSkill(debugData *data_Debug, logicData *data_Logic, skillKey Skill) *ui_GrowSkill {
	m := &ui_GrowSkill{
		debugData: debugData,
		logicData: logicData,
		skillKey:  skillKey,
	}
	m.title = skillKey.String()
	m.content = "Skill " + m.title
	return m
}

func (m *ui_GrowSkill) Init() tea.Cmd {
	m.logicData.IncreaseSkillLevel(m.skillKey)
	return common.CMD_back()
}

func (m *ui_GrowSkill) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, nil
}

func (m *ui_GrowSkill) View() string {
	return m.content
}

func (m *ui_GrowSkill) ChoiceTitle() string {
	return m.title
}
