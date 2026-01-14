package goass

import (
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs/tool/src/ui"

	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type ui_MainOperationList struct {
	*ui.List

	debugData *data_Debug
	logicData *data_Logic
	uiData    *data_UI
}

func NewUIMainOperationList(debugData *data_Debug, logicData *data_Logic, uiData *data_UI) *ui_MainOperationList {
	m := &ui_MainOperationList{
		List:      ui.NewList(0),
		debugData: debugData,
		logicData: logicData,
		uiData:    uiData,
	}
	uiGenerateProject := NewGenerateProcess(debugData, logicData)
	uiEnterBattle := NewUIBattleDungeon(debugData, logicData, uiData)
	uiClaimRewards := NewUIClaimRewards(debugData, logicData)
	uiGrowList := NewUIGrowList(debugData, logicData)
	uiSaveData := NewUISaveData(debugData, logicData)
	m.AddChoices(
		uiGenerateProject, uiEnterBattle,
		uiClaimRewards, uiGrowList,
		uiSaveData,
	)
	return m
}

func (m *ui_MainOperationList) Init() tea.Cmd {
	var cmds tea.Cmd
	// if m.uiData.GetKV(Record_TriggerTick) == 0 {
	// 	m.debugData.triggerCmdTickMap[reflect.TypeOf(m).String()]++
	// 	m.uiData.SetKV(Record_TriggerTick, 1)
	// 	cmds = tea.Batch(common.CMD_tick())
	// }
	return cmds
}

func (m *ui_MainOperationList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds tea.Cmd
	_list, _cmd := m.List.Update(msg)
	m.List, cmds = _list.(*ui.List), tea.Batch(_cmd)

	switch msg := msg.(type) {
	case common.MSG_tick:
		m.logicData.cumulativeRewards = time.Now().Unix() - m.logicData.idleTS
		return m, common.CMD_tick()
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_interrupt, common.KEY_quit:
			return m, common.CMD_back()
		}
	}
	return m, cmds
}

func (m ui_MainOperationList) View() string {
	s := fmt.Sprintf("Main Operation List: %v\n\n", m.GetChosenIndex())
	s += m.List.View()
	return s
}
