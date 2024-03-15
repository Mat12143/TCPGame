package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var items map[int]string

type model struct {
	cursor int
	choice int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "q":
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.choice = m.cursor
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(items) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(items) - 1
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("Select the zone:\n\n")

	for i := 0; i < len(items); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(items[i])
		s.WriteString("\n")
	}

	return s.String()
}

func SelectScreen(i map[int]string) int {
	items = i

	p := tea.NewProgram(model{})

	go func(p *tea.Program) {
		time.Sleep(10 * time.Second)
		p.Quit()
	}(p)

	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		return 0
	}

	if m, ok := m.(model); ok {
		return m.choice
	}
	return 0
}
