package goass

import (
	"fmt"

	"github.com/Mericusta/go-sgs/tool/src/common"
	"github.com/Mericusta/go-sgs/tool/src/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type Attribution int

const (
	ATTRIBUTE_ATK Attribution = iota + 1
	ATTRIBUTE_DEF
	ATTRIBUTE_HP
	ATTRIBUTE_MP
)

func (a Attribution) String() string {
	switch a {
	case ATTRIBUTE_ATK:
		return "ATK"
	case ATTRIBUTE_DEF:
		return "DEF"
	case ATTRIBUTE_HP:
		return "HP"
	case ATTRIBUTE_MP:
		return "MP"
	}
	return "[Attribution]"
}

type ui_GrowAttributeList struct {
	*ui.List

	debugData *data_Debug
	logicData *data_Logic

	title   string
	content string
}

func NewUIGrowAttributes(debugData *data_Debug, logicData *data_Logic) *ui_GrowAttributeList {
	m := &ui_GrowAttributeList{
		List:      ui.NewList(0),
		debugData: debugData,
		logicData: logicData,
		title:     "Attributes",
		content:   "this is grow attributes model",
	}
	uiGrowAttributeATK := NewUIGrowAttribute(debugData, logicData, ATTRIBUTE_ATK)
	uiGrowAttributeDEF := NewUIGrowAttribute(debugData, logicData, ATTRIBUTE_DEF)
	uiGrowAttributeHP := NewUIGrowAttribute(debugData, logicData, ATTRIBUTE_HP)
	uiGrowAttributeMP := NewUIGrowAttribute(debugData, logicData, ATTRIBUTE_MP)
	m.AddChoices(uiGrowAttributeATK, uiGrowAttributeDEF, uiGrowAttributeHP, uiGrowAttributeMP)
	return m
}

func (m *ui_GrowAttributeList) Init() tea.Cmd {
	return nil
}

func (m *ui_GrowAttributeList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m ui_GrowAttributeList) View() string {
	s := fmt.Sprintf("Attribute List: %v\n\n", m.GetChosenIndex())
	s += m.List.View()
	return s
}

func (m ui_GrowAttributeList) ChoiceTitle() string {
	return m.title
}
