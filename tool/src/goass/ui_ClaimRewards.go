package goass

import (
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type ui_ClaimRewards struct {
	debug_Data *data_Logic
	logicData  *data_Logic

	title   string
	content string
}

func NewUIClaimRewards(debugData *data_Debug, logicData *data_Logic) *ui_ClaimRewards {
	m := &ui_ClaimRewards{
		logicData: logicData,
		title:     "Claim Rewards",
		content:   "this is claim rewards model",
	}
	return m
}

func (m *ui_ClaimRewards) Init() tea.Cmd { return nil }

func (m *ui_ClaimRewards) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_enter, common.KEY_space:
			m.logicData.idleTS = time.Now().Unix()
			m.logicData.AddExp(m.logicData.cumulativeRewards) // 被动分配
			m.logicData.cumulativeRewards = 0
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, nil
}

func (m *ui_ClaimRewards) View() string {
	dungeonProgressBasedRewards := int64(float64(m.logicData.cumulativeRewards) * (float64(m.logicData.dungeonProgress) / float64(len(dungeonProgressPassRequirements))))
	return fmt.Sprintf("cumulative rewards: %v, claim?", m.logicData.cumulativeRewards+dungeonProgressBasedRewards)
}

func (m *ui_ClaimRewards) ChoiceTitle() string {
	return m.title
}
