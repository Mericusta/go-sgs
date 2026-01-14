package goass

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/Mericusta/go-sgs/tool/src/common"
	tea "github.com/charmbracelet/bubbletea"
)

type logic_MainProgram struct {
	// debug data
	debugData *data_Debug

	// logic data
	logicData *data_Logic

	// ui data
	uiData *data_UI
}

func NewLogicMainProgram(saveDataPath, logFilePath string, debugMode bool) *logic_MainProgram {
	var debugData *data_Debug
	if debugMode {
		debugData = newDebugData(logFilePath)
	}
	logicData := newLogicData(saveDataPath)
	uiData := newDataUI()
	logicMainProgram := &logic_MainProgram{
		debugData: debugData,
		logicData: logicData,
		uiData:    uiData,
	}
	logicMainProgram.uiData.AppendModelStack(
		NewUIMainOperationList(debugData, logicData, uiData),
	)
	return logicMainProgram
}

func (m *logic_MainProgram) String() string {
	if m.debugData == nil {
		return ""
	}
	return fmt.Sprintf(`time: %v
- ptr %p
- exp %v
- level %v
- dungeon %v
- attributes level %v
- skills level %v
- idle begin at %v
- cumulative rewards %v
- modelStack: %v
- triggerMap: %v
`, time.Now().Format(time.DateTime), m,
		m.logicData.exp, m.logicData.level, m.logicData.dungeonProgress,
		m.logicData.attributesLevel, m.logicData.skillsLevel,
		time.Unix(m.logicData.idleTS, 0).Format(time.DateTime),
		m.logicData.cumulativeRewards,
		m.uiData.ModelStackDesc(), m.debugData.triggerCmdTickMap,
	)
}

func (m *logic_MainProgram) Init() tea.Cmd {
	m.debugData.WriteLog("logic_MainProgram.Init, top model %v", reflect.TypeOf(m.uiData.ModelTop()).String())
	return m.uiData.ModelTop().Init()
}

func (m *logic_MainProgram) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.debugData.WriteLog("logic_MainProgram.Update, top model %v, handle msg %v", reflect.TypeOf(m.uiData.ModelTop()).String(), reflect.TypeOf(msg).String())

	// 从栈顶开始处理消息，如果某一层不拦截，则需要在 model 的 update 中实现

	// 消息
	// - 按键消息：只有栈顶处理
	// - 非按键消息，层层传递

	var cmds []tea.Cmd
	switch msg.(type) {
	case tea.KeyMsg:
		_, _cmd := m.uiData.ModelTop().Update(msg)
		if _cmd != nil {
			cmds = append(cmds, _cmd)
		}
	default:
		// 迭代 model 栈
		m.uiData.RangeModelStack(func(m tea.Model) bool {
			_, _cmd := m.Update(msg)
			cmds = append(cmds, _cmd)
			return true
		})

		// 栈底处理消息
		switch msg := msg.(type) {
		case common.MSG_enter:
			m.uiData.AppendModelStack(msg.Model())
			m.debugData.WriteLog("common.MSG_enter, model %v", reflect.TypeOf(msg.Model()).String())
			cmds = append(cmds, msg.Model().Init())
		case common.MSG_back:
			if m.uiData.ModelStackLen() == 1 {
				return m, tea.Quit
			}
			m.debugData.WriteLog("common.MSG_back, model %v", reflect.TypeOf(m.uiData.ModelTop()).String())
			m.uiData.PopModelStack()
			cmds = append(cmds, m.uiData.ModelTop().Init())
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *logic_MainProgram) View() string {
	modelView := m.uiData.ModelTop().View()
	return fmt.Sprintf("logic main program\n%v\ntop model\n\n%v\n", m, modelView)
}

func GoAssistant() {
	p := tea.NewProgram(
		NewLogicMainProgram(
			os.Getenv("SAVE_DATA_PATH"),
			os.Getenv("LOG_FILE_PATH"),
			os.Getenv("DEBUG_MODE") == "true", // DEBUG_MODE 模式有问题
		),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
