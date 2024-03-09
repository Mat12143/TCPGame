package ui

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var choices = []string{}

type model struct {
	cursor int
	choice string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

        case "ctrl+c":
            return m, tea.Quit

		case "enter":
			m.choice = choices[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
    s.WriteString("Choose the zone: \n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("( + ) ")
		} else {
			s.WriteString("(   ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	return s.String()
}

func ZoneSelect(op []string) string {

    choices = op

	p := tea.NewProgram(model{})

	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
        return ""
	}

	if m, ok := m.(model); ok && m.choice != "" {
        return m.choice
	}

    return ""
}
