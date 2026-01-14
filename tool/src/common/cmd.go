package common

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	MSG_enter struct{ m tea.Model }
	MSG_back  struct{}
	MSG_tick  struct{}
)

var (
	CMD_enter = func(m tea.Model) tea.Cmd { return func() tea.Msg { return MSG_enter{m: m} } }
	CMD_back  = func() tea.Cmd { return func() tea.Msg { return MSG_back{} } }
	CMD_tick  = func() tea.Cmd { return tea.Tick(time.Second, func(t time.Time) tea.Msg { return MSG_tick{} }) }
)

func (m MSG_enter) Model() tea.Model {
	return m.m
}
