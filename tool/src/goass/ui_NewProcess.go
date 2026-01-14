package goass

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Mericusta/go-extractor"
	"github.com/Mericusta/go-sgs/tool/src/common"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ui_NewProcess struct {
	debugData *data_Debug
	logicData *data_Logic

	title string

	textInput                 textinput.Model // new process form input
	inputItemTitleSlice       []string
	inputItemPlaceholderSlice []string
	inputTextSlice            []string
	inputIndex                int

	spinner          spinner.Model
	progress         progress.Model
	generateItems    []string
	generateIndex    int
	generateUIWidth  int
	generateUIHeight int
	generateDone     bool
	noModule         bool
}

var (
	currentItemNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle            = lipgloss.NewStyle().Margin(1, 2)
	checkMark            = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	packages             = []string{
		"Creating project structure...",
		// - gate [process name]
		//   - main.go [process entry point]
		//   - logic [logic package]
		//     - handler
		//       - handler.go
		//     - interface
		//       - server.go
		//     - object
		//       - server.go
		//   - router
		//     - router.go
		"Initializing go mod...",
		// go mod init [module name]
		"Initializing framework go-sgs...",
		// write main.go
		"Creating main actor...",
		// write logic/interface/server.go
		// write logic/object/server.go
		// write router/router.go
		"Creating websocket dialer...",
		// write logic/handler/dialer.go
		// write logic/object/dialer.go
		"Executing go tidy...",
		// go tidy
		"Executing go mod vendor...",
		// go mod vendor
	}
)

func NewGenerateProcess(debugData *data_Debug, logicData *data_Logic) *ui_NewProcess {
	inputIndex := 0

	// input items
	dataInputItemTitleSlice := []string{
		"Go Module",       // module
		"Process Name",    // gate
		"Main Actor Name", // gateServer
		"WebSocket Port",  // 8080
	}
	dataInputItemPlaceHolderSlice := []string{
		"None",
		"go-sgs",
		"MainServer",
		"8080",
	}
	// input UI model
	modelTextInput := textinput.New()
	modelTextInput.Placeholder = dataInputItemPlaceHolderSlice[inputIndex]
	modelTextInput.Prompt = ": "
	modelTextInput.Focus()
	modelTextInput.CharLimit = 156
	modelTextInput.Width = 20

	m := &ui_NewProcess{
		logicData:                 logicData,
		debugData:                 debugData,
		textInput:                 modelTextInput,
		title:                     "New Process",
		inputItemTitleSlice:       dataInputItemTitleSlice,
		inputItemPlaceholderSlice: dataInputItemPlaceHolderSlice,
		inputTextSlice:            make([]string, 0, 2),
		inputIndex:                inputIndex,
	}

	return m
}

func (m *ui_NewProcess) initGenerateProgressUI() {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	if m.noModule {
		m.generateItems = append(m.generateItems, packages[0])
		m.generateItems = append(m.generateItems, packages[2:]...)
	} else {
		m.generateItems = packages
	}
	m.spinner = s
	m.progress = p
}

func (m *ui_NewProcess) Init() tea.Cmd {
	m.textInput.SetValue("")
	m.inputIndex = 0
	m.inputTextSlice = make([]string, 0, 2)
	return textinput.Blink
}

func (m *ui_NewProcess) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd // 优先更新子 model
	_textInput, _cmd := m.textInput.Update(msg)
	m.textInput, cmd = _textInput, tea.Batch(_cmd)
	_spinner, _cmd := m.spinner.Update(msg)
	m.spinner, cmd = _spinner, tea.Batch(cmd, _cmd)
	_progress, _cmd := m.progress.Update(msg)
	if _progress, ok := _progress.(progress.Model); ok {
		m.progress, cmd = _progress, tea.Batch(cmd, _cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.inputIndex >= len(m.inputItemTitleSlice) {
				break
			}

			// logic
			inputValue := m.textInput.Value()
			if len(inputValue) > 0 {
				m.inputTextSlice = append(m.inputTextSlice, inputValue)
			} else {
				m.inputTextSlice = append(m.inputTextSlice, m.textInput.Placeholder)
			}
			if m.inputIndex == 0 && m.inputTextSlice[m.inputIndex] == "None" {
				m.noModule = true
				m.debugData.WriteLog("ui_NewProcess.Update: no need module")
			}
			m.debugData.WriteLog("ui_NewProcess.Update: item %v use %v input %v", m.inputItemTitleSlice[m.inputIndex], m.inputTextSlice[m.inputIndex], inputValue)

			// cmd
			m.inputIndex++
			if m.inputIndex >= len(m.inputItemTitleSlice) {
				m.debugData.WriteLog("ui_NewProcess.Update: generate process %v with main actor %v", m.inputTextSlice[0], m.inputTextSlice[1])
				m.initGenerateProgressUI()
				cmd = tea.Batch(cmd, m.generateProcessItem(), m.spinner.Tick)
			} else {
				m.textInput.SetValue("")
				m.textInput.Placeholder = m.inputItemPlaceholderSlice[m.inputIndex]
			}
		case tea.KeyEsc:
			return m, common.CMD_back()
		}
	case tea.WindowSizeMsg:
		m.generateUIWidth, m.generateUIHeight = msg.Width, msg.Height
	case generatedProcessItemMsg:
		if m.generateIndex >= len(m.generateItems) {
			m.generateDone = true
			break
		}

		// logic
		m.debugData.WriteLog("ui_NewProcess.Update: generatedProcessItemMsg, %v", m.generateItems[m.generateIndex])

		m.generateIndex++
		if m.generateIndex >= len(m.generateItems) {
			m.generateDone = true
			break
		}

		_cmd := m.progress.SetPercent(float64(m.generateIndex) / float64(len(m.generateItems)))
		cmd = tea.Batch(_cmd, m.generateProcessItem())
	}

	return m, cmd
}

type generatedProcessItemMsg struct{}

func (m *ui_NewProcess) generateProcessItem() tea.Cmd {
	var err error
	m.debugData.WriteLog("ui_NewProcess.generateProcessItem: %v", m.generateItems[m.generateIndex])
	switch m.generateIndex {
	case 0:
		err = m.creatingProjectStructure()
	case 1:
		err = m.initializingGoMod()
	case 2:
		err = m.initializingFrameworkGoSGS()
	}

	if err != nil {
		m.debugData.WriteLog("ui_NewProgress.generateProcessItem: generate item %v occurs error %v", m.generateIndex, err)
		panic(err)
	}

	d := time.Millisecond * time.Duration(500)
	return tea.Tick(d, func(t time.Time) tea.Msg { return generatedProcessItemMsg{} })
}

func (m *ui_NewProcess) creatingProjectStructure() error {
	var (
		toCreateFolderSlice = make([]string, 0, 8)
		toCreateFileMap     = make(map[string]string)
		processDir          = m.inputTextSlice[1]
	)
	toCreateFolderSlice = append(toCreateFolderSlice, processDir)
	toCreateFileMap[fmt.Sprintf("%s/main.go", processDir)] = "main"
	toCreateFolderSlice = append(toCreateFolderSlice, fmt.Sprintf("%s/logic", processDir))
	toCreateFolderSlice = append(toCreateFolderSlice, fmt.Sprintf("%s/logic/handler", processDir))
	toCreateFileMap[fmt.Sprintf("%s/logic/handler/handler.go", processDir)] = "handler"
	toCreateFolderSlice = append(toCreateFolderSlice, fmt.Sprintf("%s/logic/interface", processDir))
	toCreateFileMap[fmt.Sprintf("%s/logic/interface/server.go", processDir)] = "inObj"
	toCreateFolderSlice = append(toCreateFolderSlice, fmt.Sprintf("%s/logic/object", processDir))
	toCreateFileMap[fmt.Sprintf("%s/logic/object/server.go", processDir)] = "obj"
	toCreateFolderSlice = append(toCreateFolderSlice, fmt.Sprintf("%s/router", processDir))
	toCreateFileMap[fmt.Sprintf("%s/router/router.go", processDir)] = "router"

	for _, toCreateFolder := range toCreateFolderSlice {
		if _, err := os.Stat(toCreateFolder); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("os.Stat, %v, %v", toCreateFolder, err)
		}
		if err := os.Mkdir(toCreateFolder, os.ModePerm); err != nil && !os.IsExist(err) {
			return fmt.Errorf("os.Mkdir, %v, %v", toCreateFolder, err)
		}
	}

	for toCreateFile, pkgName := range toCreateFileMap {
		if _, err := os.Stat(toCreateFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("os.Stat, %v, %v", toCreateFile, err)
		}
		if fs, err := os.Create(toCreateFile); err != nil && !os.IsExist(err) {
			return fmt.Errorf("os.Create, %v, %v", toCreateFile, err)
		} else {
			if _, err = fs.WriteString("package " + pkgName + "\n"); err != nil {
				return fmt.Errorf("fs.WriteString, %v, %v", toCreateFile, err)
			}
		}
	}

	return nil
}

func (m *ui_NewProcess) initializingGoMod() error {
	if m.noModule {
		m.debugData.WriteLog("ui_NewProgress.generateProcessItem: skip go mod init")
		return nil
	}
	dirName := m.inputTextSlice[1]
	moduleName := m.inputTextSlice[0]
	cmd := fmt.Sprintf("go mod init %s", moduleName)
	if err := common.ExecShellCommandInDir(cmd, dirName); err != nil {
		return fmt.Errorf("common.ExecShellCommandInDir, %v", err)
	}
	return nil
}

func (m *ui_NewProcess) initializingFrameworkGoSGS() error {
	processDir := m.inputTextSlice[1]
	gfm, err := extractor.ExtractGoFileMeta(fmt.Sprintf("%s/main.go", processDir))
	if err != nil {
		return fmt.Errorf("extractor.ExtractGoFileMeta, %v", err)
	}

}

func (m *ui_NewProcess) View() string {
	content := ""
	inputDone := true
	for index := 0; index <= m.inputIndex && index < len(m.inputItemTitleSlice); index++ {
		if index >= len(m.inputTextSlice) {
			content += fmt.Sprintf("%s%s\n", m.inputItemTitleSlice[index], m.textInput.View())
			inputDone = false
			break
		} else {
			content += fmt.Sprintf("%s: %s\n", m.inputItemTitleSlice[index], m.inputTextSlice[index])
		}
	}

	if !inputDone {
		return content
	}

	n := len(m.generateItems)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	completedItem := ""
	for index := 0; index < m.generateIndex; index++ {
		completedItem += fmt.Sprintf("%s %s\n", checkMark, m.generateItems[index])
	}

	doneInfo := ""
	if m.generateDone {
		doneInfo = doneStyle.Render(fmt.Sprintf("Done! Generated new process %v\n", m.inputTextSlice[1]))
		return completedItem + doneInfo
	}

	generateCount := fmt.Sprintf(" %*d/%*d", w, m.generateIndex, w, n)
	spinView := m.spinner.View() + " "
	progressView := m.progress.View()
	cellsAvail := max(0, m.generateUIWidth-lipgloss.Width(spinView+progressView+generateCount))

	generateItem := currentItemNameStyle.Render(m.generateItems[m.generateIndex])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(generateItem)

	cellsRemaining := max(0, m.generateUIWidth-lipgloss.Width(spinView+info+progressView+generateCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return completedItem + spinView + info + gap + progressView + generateCount
}

func (m *ui_NewProcess) ChoiceTitle() string {
	return m.title
}
