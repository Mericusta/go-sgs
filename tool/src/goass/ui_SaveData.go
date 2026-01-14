package goass

import (
	"fmt"

	"github.com/Mericusta/go-sgs/tool/src/common"

	"github.com/Mericusta/go-stp"
	tea "github.com/charmbracelet/bubbletea"
)

type ui_SaveData struct {
	debugData *data_Debug
	logicData *data_Logic

	title   string
	content string
}

func NewUISaveData(debugData *data_Debug, logicData *data_Logic) *ui_SaveData {
	m := &ui_SaveData{
		debugData: debugData,
		logicData: logicData,
		title:     "Save Data",
		content:   "saving data",
	}
	return m
}

func (m *ui_SaveData) Init() tea.Cmd {
	logicDataJson := m.logicData.ToJSON()
	err := stp.WriteFileByOverwriting(m.logicData.saveDataPath, func(b []byte) ([]byte, error) {
		return logicDataJson, nil
	})
	if err != nil {
		m.content = fmt.Sprintf("[ERROR] %v", err.Error())
		return nil
	}
	return common.CMD_back()
}

func (m *ui_SaveData) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, nil
}

func (m *ui_SaveData) View() string {
	return m.content
}

func (m *ui_SaveData) ChoiceTitle() string {
	return m.title
}
