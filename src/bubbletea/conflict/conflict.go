package conflict

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/bubbletea/custom/textarea"
	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/resolver"
	"github.com/costaluu/flag/types"
)

var conflictMutex sync.Mutex = sync.Mutex{}
var hardQuit bool = false

// AcceptBothChanges returns the union of the changes from both branches
func acceptBothChanges(conflict string) (string, error) {
	sections, err := splitConflict(conflict)
	
	if err != nil {
		return "", err
	}

	var list []string

	filesystem.FileWriteContentToFile("output", fmt.Sprintf("%+v", sections))

	if len(sections.before) > 0 {
		list = append(list, sections.before)
	}

	if len(sections.current) > 0 {
		list = append(list, sections.current)
	}

	if len(sections.incoming) > 0 {
		list = append(list, sections.incoming)
	}

	if len(sections.after) > 0 {
		list = append(list, sections.after)
	}

	return strings.Join(list, "\n"), nil
}

// AcceptCurrentChanges returns the changes from the current branch
func acceptCurrentChanges(conflict string) (string, error) {
	sections, err := splitConflict(conflict)

	if err != nil {
		return "", err
	}

	var list []string

	if len(sections.before) > 0 {
		list = append(list, sections.before)
	}

	if len(sections.current) > 0 {
		list = append(list, sections.current)
	}

	if len(sections.after) > 0 {
		list = append(list, sections.after)
	}
	
	return strings.Join(list, "\n"), nil
}

// AcceptIncomingChanges returns the changes from the incoming branch
func acceptIncomingChanges(conflict string) (string, error) {
	sections, err := splitConflict(conflict)
	
	if err != nil {
		return "", err
	}
	
	var list []string

	if len(sections.before) > 0 {
		list = append(list, sections.before)
	}

	if len(sections.incoming) > 0 {
		list = append(list, sections.incoming)
	}

	if len(sections.after) > 0 {
		list = append(list, sections.after)
	}

	return strings.Join(list, "\n"), nil
}

// Helper function to split the conflict into its components
type conflictSections struct {
	before string
	current  string
	incoming string
	after string
}

func splitConflict(conflict string) (conflictSections, error) {
	// Split the conflict string into its components
	parts := strings.Split(conflict, "\n")

	var sections conflictSections
	var beforeSection []string
	var afterSection []string
	var currentSection []string
	var incomingSection []string
	var inCurrent, inIncoming, afterInIncoming bool

	for _, line := range parts {
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			inCurrent = true
		case strings.HasPrefix(line, "======="):
			inCurrent = false
			inIncoming = true
		case strings.HasPrefix(line, ">>>>>>>"):
			inIncoming = false
			afterInIncoming = true
		default:
			if !inCurrent && !inIncoming && !afterInIncoming {
				beforeSection = append(beforeSection, line)
			} else if inCurrent {
				currentSection = append(currentSection, line)
			} else if inIncoming {
				incomingSection = append(incomingSection, line)
			} else if afterInIncoming {
				afterSection = append(afterSection, line)
			}
		}
	}

	if len(currentSection) == 0 && len(incomingSection) == 0 {
		return sections, errors.New("invalid conflict: no content detected")
	}

	sections.current = strings.Join(currentSection, "\n")
	sections.incoming = strings.Join(incomingSection, "\n")
	sections.after = strings.Join(afterSection, "\n")
	sections.before = strings.Join(beforeSection, "\n")

	return sections, nil
}

var (
	regexStr = `<<<<<<<|=======|>>>>>>>`
	regexPattern = regexp.MustCompile(regexStr)
	solved = lipgloss.NewStyle().Background(lipgloss.Color("42")).Foreground(lipgloss.Color("255")).SetString(" SOLVED ").Bold(true).Render()
	notSolved = lipgloss.NewStyle().Background(lipgloss.Color("160")).Foreground(lipgloss.Color("255")).SetString(" NOT SOLVED ").Bold(true).Render()
	ResolvedConflicts []string
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	cursorLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

type keymap = struct {
	next, prev, current, incoming, both, contextup, contextdown, undo, redo, exit, quit key.Binding
}

func newTextarea(content string) textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	
	t.SetValue(content)
	t.Viewport.GotoTop()

	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	
	t.Blur()
	
	return t
}

type model struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	input textarea.Model
	conflictIndex int
	conflicts []resolver.ConflictRecord
	conflictPath string
	title string
	fixViewport bool
}

func newModel(conflicts []resolver.ConflictRecord, conflictPath string, title string) model {
	m := model{
		fixViewport: false,
		title: title,
		conflictPath: conflictPath,
		input: textarea.Model{},
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("ctrl+down"),
				key.WithHelp("ctrl + ↓", "next conflict"),
			),
			prev: key.NewBinding(
				key.WithKeys("ctrl+up"),
				key.WithHelp("ctrl + ↑", "previous conflict"),
			),
			current: key.NewBinding(
				key.WithKeys("ctrl+left"),
				key.WithHelp("ctrl + ←", "accept current"),
			),
			incoming: key.NewBinding(
				key.WithKeys("ctrl+right"),
				key.WithHelp("ctrl + →", "accept incoming"),
			),
			both: key.NewBinding(
				key.WithKeys("ctrl+b"),
				key.WithHelp("ctrl + b", "accept both"),
			),
			contextup:  key.NewBinding(
				key.WithKeys("shift+up"),
				key.WithHelp("shift + ↑", "add context up"),
			),
			contextdown:  key.NewBinding(
				key.WithKeys("shift+down"),
				key.WithHelp("shift + ↓", "add context down"),
			),
			undo: key.NewBinding(
				key.WithKeys("ctrl+z"),
				key.WithHelp("ctrl + z", "undo"),
			),
			redo: key.NewBinding(
				key.WithKeys("ctrl+y"),
				key.WithHelp("ctrl + y", "redo"),
			),
			exit: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "save and exit"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
		conflicts: conflicts,
		conflictIndex: 0,
	}

	m.input = newTextarea(conflicts[0].Current.Content)
	m.input.Focus()
	
	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentContent := m.input.Value()
	result := regexPattern.Find([]byte(currentContent))

	if len(result) > 0 {
		m.conflicts[m.conflictIndex].Current.Resolved = false
	} else {
		m.conflicts[m.conflictIndex].Current.Resolved = true
	}

	tempConflict := m.conflicts[m.conflictIndex]
	tempConflict.Current.Content = currentContent

	m.conflicts[m.conflictIndex].RecordChange(tempConflict.Current)

	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
				case key.Matches(msg, m.keymap.quit):
					hardQuit = true
					m.input.Blur()
					return m, tea.Quit
				case key.Matches(msg, m.keymap.exit):
					m.input.Blur()
					return m, tea.Quit
				case key.Matches(msg, m.keymap.next):
					if m.conflictIndex + 1 < len(m.conflicts) {
						m.conflictIndex++
						m.input.SetValue(m.conflicts[m.conflictIndex].Current.Content)
						m.fixViewport = false
					}
				case key.Matches(msg, m.keymap.prev):
					if m.conflictIndex > 0 {
						m.conflictIndex--
						m.input.SetValue(m.conflicts[m.conflictIndex].Current.Content)
						m.fixViewport = false
					}
				case key.Matches(msg, m.keymap.both):
					if !m.conflicts[m.conflictIndex].Current.Resolved {
						resolvedConflict, err := acceptBothChanges(m.input.Value())

						if err == nil {
							temp := m.conflicts[m.conflictIndex].Current
							temp.Content = resolvedConflict

							m.conflicts[m.conflictIndex].RecordChange(temp)
							m.input.SetValue(resolvedConflict)
							m.conflicts[m.conflictIndex].Current.Resolved = true
						}
					}
				case key.Matches(msg, m.keymap.current):
					if !m.conflicts[m.conflictIndex].Current.Resolved {
						resolvedConflict, err := acceptCurrentChanges(m.input.Value())

						if err == nil {
							temp := m.conflicts[m.conflictIndex].Current
							temp.Content = resolvedConflict

							m.conflicts[m.conflictIndex].RecordChange(temp)
							m.input.SetValue(resolvedConflict)
							m.conflicts[m.conflictIndex].Current.Resolved = true
						}
					}
				case key.Matches(msg, m.keymap.incoming):
					if !m.conflicts[m.conflictIndex].Current.Resolved {
						resolvedConflict, err := acceptIncomingChanges(m.input.Value())

						if err == nil {
							temp := m.conflicts[m.conflictIndex].Current
							temp.Content = resolvedConflict

							m.conflicts[m.conflictIndex].RecordChange(temp)
							m.input.SetValue(resolvedConflict)
							m.conflicts[m.conflictIndex].Current.Resolved = true
						}
					}
				case key.Matches(msg, m.keymap.undo):
					m.conflicts[m.conflictIndex].Undo()
					m.input.SetValue(m.conflicts[m.conflictIndex].Current.Content)
					return m, nil
				case key.Matches(msg, m.keymap.redo):
					m.conflicts[m.conflictIndex].Redo()
					m.input.SetValue(m.conflicts[m.conflictIndex].Current.Content)
					return m, nil
				case key.Matches(msg, m.keymap.contextup):
					reader, err := resolver.NewFileLineReader(m.conflictPath)
					defer reader.Close()

					if err != nil {
						logger.Fatal[error](err)
					}

					if !m.conflicts[m.conflictIndex].Current.Resolved {
						lineFetched, err := reader.ReadLine(m.conflicts[m.conflictIndex].Current.LineStart - 1)

						if err == nil {
							temp := m.conflicts[m.conflictIndex].Current
							temp.LineStart -= 1
							temp.Content = strings.Join([]string{lineFetched, m.conflicts[m.conflictIndex].Current.Content}, "\n")
							m.input.SetValue(temp.Content)
							
							m.conflicts[m.conflictIndex].RecordChange(temp)
						}
					}
				case key.Matches(msg, m.keymap.contextdown):
					reader, err := resolver.NewFileLineReader(m.conflictPath)
					defer reader.Close()

					if err != nil {
						logger.Fatal[error](err)
					}

					if !m.conflicts[m.conflictIndex].Current.Resolved {
						lineFetched, err := reader.ReadLine(m.conflicts[m.conflictIndex].Current.LineEnd + 1)

						if err == nil {
							temp := m.conflicts[m.conflictIndex].Current
							temp.LineEnd += 1
							temp.Content = strings.Join([]string{m.conflicts[m.conflictIndex].Current.Content, lineFetched}, "\n")
							m.input.SetValue(temp.Content)

							m.conflicts[m.conflictIndex].RecordChange(temp)
						}
					}		
			}
		case tea.WindowSizeMsg:
			m.height = msg.Height
			m.width = msg.Width			
	}

	m.sizeInputs()
	newModel, cmd := m.input.Update(msg)
	m.input = newModel

	if !m.fixViewport {
		m.input.Viewport.GotoTop()
		m.fixViewport = true
	}

	return m, tea.Batch(cmd)
}

func (m *model) sizeInputs() {
		m.input.SetWidth(m.width - 35)
		m.input.SetHeight(m.height - 3)
}

func (m model) View() string {
	var helpText string

	var keys []key.Binding = []key.Binding{
		m.keymap.next,
		m.keymap.prev,
	}

	if m.conflicts[m.conflictIndex].Current.Resolved {
		helpText = fmt.Sprintf("\nConflict %d/%d %s\n", m.conflictIndex + 1, len(m.conflicts), solved)
	} else {
		keys = append(keys, m.keymap.current, m.keymap.incoming, m.keymap.both, m.keymap.contextup, m.keymap.contextdown)
		helpText = fmt.Sprintf("\nConflict %d/%d %s\n", m.conflictIndex + 1, len(m.conflicts), notSolved)
	}

	keys = append(keys, m.keymap.undo, m.keymap.redo, m.keymap.exit, m.keymap.quit)

	for _, key := range keys {
		helpText += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).SetString("• " + key.Help().Key).Render() + " " + key.Help().Desc + "\n"
	}

	return constants.MergeMark.Render() + " " + lipgloss.NewStyle().SetString(m.title).Bold(true).Render() + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, m.input.View(), " ", helpText)
}

func SolveConflicts(content []resolver.ConflictRecord, conflictPath string, title string) []resolver.ConflictRecord {
	if len(content) > 0 {
		model := newModel(content, conflictPath, title)
		
		if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
			logger.Fatal[error](err)
		}

		if hardQuit {
			os.Exit(0)
		}
		
		return model.conflicts
	} else {
		return content
	}
}

// FindGitConflicts reads a file line by line and prints the lines containing git conflicts
func FindGitConflicts(filePath string) ([]resolver.ConflictRecord) {
	file, err := os.Open(filePath)

	if err != nil {
		logger.Fatal[error](err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var conflictBlock []string
	inConflict := false
	var lineStart int
	var conflicts []resolver.ConflictRecord = []resolver.ConflictRecord{}
	var currentLineNumber int = 0

	for scanner.Scan() {
		currentLineNumber++
		line := scanner.Text()

		if strings.HasPrefix(line, "<<<<<<<") {
			inConflict = true
			lineStart = currentLineNumber
		}

		if inConflict {
			conflictBlock = append(conflictBlock, line)
		}

		if strings.HasPrefix(line, ">>>>>>>") {
			inConflict = false

			var conflict types.Conflict = types.Conflict{
				LineStart: lineStart,
				LineEnd: currentLineNumber,
				Content: strings.Join(conflictBlock, "\n"),
				Resolved: false,
			}

			conflicts = append(conflicts, resolver.ConflictRecord{
				Current: conflict,
				UndoStack: resolver.NewStack[types.Conflict](),
				RedoStack: resolver.NewStack[types.Conflict](),
			})

			conflictBlock = nil // Clear the block after processing
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal[error](err)
	}

	return conflicts
}

func Resolve(title string) {
	// var iterator int = 0
	var allConflictsSolved bool = false
	
	for !allConflictsSolved {
		// iterator++
	
		conflicts := FindGitConflicts("../.features/merge-tmp")
		processedConflicts := SolveConflicts(conflicts, "../.features/merge-tmp", title)
		var solvedConflicts []resolver.ConflictRecord
		var unSolvedConflicts []resolver.ConflictRecord

		for _, conflict := range processedConflicts {
			if !conflict.Current.Resolved {
				unSolvedConflicts = append(unSolvedConflicts, conflict)
			} else {
				solvedConflicts = append(solvedConflicts, conflict)
			}
		}

		var lineOffset int = 0

		for _, solvedConflict := range solvedConflicts {
			stringContent := strings.Split(solvedConflict.Current.Content, "\n")

			err := filesystem.FileReplaceLinesInFile("../.features/merge-tmp", solvedConflict.Current.LineStart + lineOffset, solvedConflict.Current.LineEnd + lineOffset, stringContent)
			
			if err != nil {
				logger.Fatal[error](err)
			}

			linesCountBefore := (solvedConflict.Current.LineEnd + 1) - solvedConflict.Current.LineStart
			linesCountAfter := len(stringContent)
			
			if linesCountBefore > linesCountAfter {
				lineOffset -= int(math.Abs(float64(linesCountBefore) - float64(linesCountAfter)))
			} else if linesCountBefore < linesCountAfter {
				lineOffset += int(math.Abs(float64(linesCountBefore) - float64(linesCountAfter)))
			}
		}
		
		if len(unSolvedConflicts) == 0 {
			allConflictsSolved = true
		}
	}
}