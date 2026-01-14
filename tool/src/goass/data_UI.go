package goass

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	Record_TriggerTick = iota + 1
)

type data_UI struct {
	kvMap      map[int]int // key-value 记录
	modelStack []tea.Model // model 栈
}

func newDataUI() *data_UI {
	return &data_UI{kvMap: make(map[int]int)}
}

func (d *data_UI) SetKV(k, v int)  { d.kvMap[k] = v }
func (d *data_UI) GetKV(k int) int { return d.kvMap[k] }
func (d *data_UI) DelKV(k int)     { delete(d.kvMap, k) }

func (d *data_UI) AppendModelStack(m tea.Model) {
	d.modelStack = append(d.modelStack, m)
}

func (d *data_UI) PopModelStack() {
	d.modelStack = d.modelStack[:len(d.modelStack)-1]
}

func (d *data_UI) ModelTop() tea.Model {
	return d.modelStack[len(d.modelStack)-1]
}

func (d *data_UI) ModelStackLen() int {
	return len(d.modelStack)
}

func (d *data_UI) RangeModelStack(f func(tea.Model) bool) {
	for i := len(d.modelStack) - 1; i >= 0; i-- {
		if !f(d.modelStack[i]) {
			return
		}
	}
}

func (d *data_UI) ModelStackDesc() []string {
	modelStackDescSlice := make([]string, 0, len(d.modelStack))
	for _, model := range d.modelStack {
		modelStackDescSlice = append(modelStackDescSlice, reflect.TypeOf(model).String())
	}
	return modelStackDescSlice
}
