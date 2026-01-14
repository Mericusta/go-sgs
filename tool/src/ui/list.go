package ui

import (
	"fmt"
	"slices"

	"github.com/Mericusta/go-sgs/tool/src/common"

	tea "github.com/charmbracelet/bubbletea"
)

type IOperation interface {
	tea.Model

	ChoiceTitle() string // 操作列表名称
}

type List struct {
	chosen       int          // 光标所在的位置
	choiceModels []IOperation // 待选择的操作
}

func NewList(i int) *List {
	return &List{chosen: i}
}

func (m *List) GetChosenIndex() int         { return m.chosen }
func (m *List) SetChosenIndex(i int)        { m.chosen = i }
func (m *List) AddChoices(ms ...IOperation) { m.choiceModels = append(m.choiceModels, ms...) }
func (m *List) DelChoices(i int)            { m.choiceModels = slices.Delete(m.choiceModels, i, 0) }

func (m *List) Init() tea.Cmd { return nil }

func (m *List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case common.KEY_up:
			if m.chosen > 0 {
				m.chosen--
			}
		case common.KEY_down:
			if m.chosen < len(m.choiceModels)-1 {
				m.chosen++
			}
		case common.KEY_enter, common.KEY_space:
			return m, common.CMD_enter(m.choiceModels[m.chosen])
		}
	}
	return m, nil
}

func (m List) View() string {
	var s string
	for i, choice := range m.choiceModels {
		cursor := " "
		if m.chosen == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice.ChoiceTitle())
	}
	return s
}
