package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/types"
)

type IteratorModel struct {
	paths   []types.FilePathCategory
	index   int
	width   int
	height  int
	spinner spinner.Model
	done    bool
	runner func (path types.FilePathCategory) tea.Cmd
}

var (
	currentPathNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
)

func newIteratorModel(pathList []types.FilePathCategory, runner func(path types.FilePathCategory) tea.Cmd) IteratorModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return IteratorModel{
		paths:   pathList,
		spinner: s,
		runner: runner,
	}
}

func (m IteratorModel) Init() tea.Cmd {
	return tea.Batch(m.runner(m.paths[m.index]), m.spinner.Tick)
}

func (m IteratorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case finishedSyncMsg:
		path := m.paths[m.index]
		if m.index >= len(m.paths)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s %s", constants.CheckMark.Render(), path.Path), // print the last success message
				tea.Quit,                             // exit the program
			)
		}

		// Update progress bar
		m.index++

		return m, tea.Batch(
			tea.Printf("%s %s", constants.CheckMark.Render(), path.Path), // print success message above our program
			m.runner(m.paths[m.index]),  // process the next path
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m IteratorModel) View() string {
	n := len(m.paths)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	pathCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	cellsAvail := max(0, m.width-lipgloss.Width(spin+pathCount))

	path := currentPathNameStyle.Render(m.paths[m.index].Path)
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Syncing " + path)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+pathCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + pathCount
}

type finishedSyncMsg types.FilePathCategory

func FileIterator(list []types.FilePathCategory, run func (parameter types.FilePathCategory)) {
	runner := func(path types.FilePathCategory) tea.Cmd {
		run(path)

		return tea.Tick(time.Millisecond, func(t time.Time) tea.Msg {
			return finishedSyncMsg(path)
		})
	}

	if len(list) == 0 {
		return
	}

	model := newIteratorModel(list, runner)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		logger.Fatal[error](err)
	}

}