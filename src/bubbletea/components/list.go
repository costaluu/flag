package components

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
)

var ListResult ListItem

type ListItem struct {
	ItemTitle string
	ItemDesc string
	ItemValue string
}

func (i ListItem) Title() string {
	return i.ItemTitle
}

func (i ListItem) Description() string { return i.ItemDesc }
func (i ListItem) FilterValue() string { return i.ItemTitle }

type ListModel struct {
	list list.Model
	choice string
	quitting bool
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		} else if msg.String() == "enter" {
			m.quitting = true
			item, ok := m.list.SelectedItem().(ListItem)
			
			if ok {
				m.choice = item.ItemTitle
				ListResult = item
			}

			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return docStyle.Render(m.list.View())
}

func PickerList(title string, items []ListItem) ListItem {
	var customDelegate list.DefaultDelegate = list.NewDefaultDelegate()
	customDelegate.SetSpacing(0)

	customDelegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color(constants.AccentColor)).
		Padding(0, 0, 0, 1)

	customDelegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color(constants.AccentDarkerColor)).
		Padding(0, 0, 0, 1)
	
	var parsedItems []list.Item = []list.Item{}

	for _, listItem := range items {
		parsedItems = append(parsedItems, listItem)
	}

	model := ListModel{list: list.New(parsedItems, customDelegate, 0, 0) }

	model.list.Title = title
	model.list.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("166")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	model.list.SetShowStatusBar(false)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}

	ListResultTemp := ListResult 
	
	ListResult.ItemDesc = ""
	ListResult.ItemTitle = ""
	ListResult.ItemValue = ""

	return ListResultTemp
}