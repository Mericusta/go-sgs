package goass

import (
	"strings"

	"github.com/Mericusta/go-sgs/tool/src/common"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	DUNGEON_REQUIREMENT_ATTRIBUTE = iota + 1
	DUNGEON_REQUIREMENT_SKILL
)

var dungeonProgressPassRequirements map[int64]map[int64]map[int64]int64 = map[int64]map[int64]map[int64]int64{
	1:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 1, int64(ATTRIBUTE_DEF): 1}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 1, int64(SKILL_2): 1}},
	2:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 2, int64(ATTRIBUTE_DEF): 2}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 2, int64(SKILL_2): 2}},
	3:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 3, int64(ATTRIBUTE_DEF): 3}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 3, int64(SKILL_2): 3}},
	4:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 4, int64(ATTRIBUTE_DEF): 4}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 4, int64(SKILL_2): 4}},
	5:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 5, int64(ATTRIBUTE_DEF): 5}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 5, int64(SKILL_2): 5}},
	6:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 6, int64(ATTRIBUTE_DEF): 6}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 6, int64(SKILL_2): 6}},
	7:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 7, int64(ATTRIBUTE_DEF): 7}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 7, int64(SKILL_2): 7}},
	8:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 8, int64(ATTRIBUTE_DEF): 8}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 8, int64(SKILL_2): 8}},
	9:  {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 9, int64(ATTRIBUTE_DEF): 9}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 9, int64(SKILL_2): 9}},
	10: {DUNGEON_REQUIREMENT_ATTRIBUTE: {int64(ATTRIBUTE_ATK): 10, int64(ATTRIBUTE_DEF): 10}, DUNGEON_REQUIREMENT_SKILL: {int64(SKILL_1): 10, int64(SKILL_2): 10}},
}

type ui_BattleDungeon struct {
	progress.Model

	debugData *data_Debug
	logicData *data_Logic
	uiData    *data_UI

	title   string
	content string
	percent float64 // 百分比 [0,1]
	padding int     // 左边距

	dungeonRequirements map[int64]map[int64]int64
	dungeonResults      map[int64]map[int64]bool
	tickIndex           int
}

func NewUIBattleDungeon(debugData *data_Debug, logicData *data_Logic, uiData *data_UI) *ui_BattleDungeon {
	const (
		initPercent = 0
		padding     = 0
		maxWidth    = 50
	)
	m := &ui_BattleDungeon{
		debugData:           debugData,
		logicData:           logicData,
		uiData:              uiData,
		title:               "Battle Dungeon",
		content:             "this is battle dungeon model",
		Model:               progress.New(progress.WithWidth(maxWidth)),
		padding:             padding,
		dungeonRequirements: make(map[int64]map[int64]int64),
		dungeonResults:      make(map[int64]map[int64]bool),
		tickIndex:           0,
	}
	_dungeonRequirements, has := dungeonProgressPassRequirements[m.logicData.dungeonProgress]
	if has {
		for requirementType, requirementTypeMap := range _dungeonRequirements {
			m.dungeonRequirements[requirementType] = make(map[int64]int64)
			m.dungeonResults[requirementType] = make(map[int64]bool)
			for requirementKey, requirementValue := range requirementTypeMap {
				m.dungeonRequirements[requirementType][requirementKey] = requirementValue
				m.dungeonResults[requirementType][requirementKey] = false
			}
		}
	}

	return m
}

func (m *ui_BattleDungeon) Init() tea.Cmd {
	m.debugData.WriteLog("ui_BattleDungeon.Init")

	m.percent = 0
	if len(m.dungeonRequirements) == 0 || len(m.dungeonResults) == 0 {
		m.debugData.WriteLog("ui_BattleDungeon.Init, invalid requirements and results, back")
		return common.CMD_back()
	}
	return nil
}

func (m *ui_BattleDungeon) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds tea.Cmd
	_progress, _cmd := m.Model.Update(msg)
	m.Model, cmds = _progress.(progress.Model), tea.Batch(_cmd)

	switch msg := msg.(type) {
	case common.MSG_tick:
		if m.percent >= 1.0 || !m.TickDungeon() {
			return m, common.CMD_back()
		}

		resultCount := 0
		for _, typeResults := range m.dungeonResults {
			for _, v := range typeResults {
				if v {
					resultCount++
				}
			}
		}
		targetCount := 0
		for _, v := range m.dungeonRequirements {
			targetCount += len(v)
		}

		m.debugData.WriteLog("ui_BattleDungeon.Update, resultCount %v, targetCount %v", resultCount, targetCount)

		m.tickIndex++
		m.percent = float64(resultCount) / float64(targetCount)
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_backspace:
			return m, common.CMD_back()
		}
	}
	return m, cmds
}

func (m *ui_BattleDungeon) View() string {
	s := "Battle Dungeon Progress:\n"
	s += strings.Repeat(" ", m.padding) + m.Model.ViewAs(m.percent) + ""
	return s
}

func (m *ui_BattleDungeon) ChoiceTitle() string {
	return m.title
}

func (m *ui_BattleDungeon) TickDungeon() bool {
	m.debugData.WriteLog("ui_BattleDungeon.TickDungeon")

	for requirementType, requirementTypeResults := range m.dungeonResults {
		requirementTypeMap, has := m.dungeonRequirements[requirementType]
		if !has {
			m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, not has requirementType %v", requirementType)
			return false
		}
		for requirementKey, done := range requirementTypeResults {
			if done {
				m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v requirementKey %v done", requirementType, requirementKey)
				continue
			}
			requirementValue, has := requirementTypeMap[requirementKey]
			if !has {
				m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v not has requirementKey %v", requirementType, requirementKey)
				return false
			}
			m.dungeonResults[requirementType][requirementKey] = true
			switch requirementType {
			case DUNGEON_REQUIREMENT_ATTRIBUTE:
				if m.logicData.attributesLevel[Attribution(requirementKey)] < requirementValue {
					m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v requirementKey %v not match requirementValue %v", requirementType, requirementKey, requirementValue)
					return false
				}
			case DUNGEON_REQUIREMENT_SKILL:
				if m.logicData.skillsLevel[Skill(requirementKey)] < requirementValue {
					m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v requirementKey %v not match requirementValue %v", requirementType, requirementKey, requirementValue)
					return false
				}
			default:
				m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v unknown", requirementType)
				return false
			}
			// just check once
			m.debugData.WriteLog("ui_BattleDungeon.TickDungeon, requirementType %v requirementKey %v pass", requirementType, requirementKey)
			return true
		}
	}

	return false
}
