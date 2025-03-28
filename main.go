package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Color string

const (
	Reset   Color = "\033[0m"
	Red     Color = "\033[31m"
	Green   Color = "\033[32m"
	Yellow  Color = "\033[33m"
	Blue    Color = "\033[34m"
	Magenta Color = "\033[35m"
	Cyan    Color = "\033[36m"
	Gray    Color = "\033[90m"
	White   Color = "\033[97m"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	p := tea.NewProgram(New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	width      int
	height     int
	GameEngine *Engine
}

func New() model {
	e := NewEngine()
	e.GenerateRandomText()

	return model{
		GameEngine: e,
	}
}

type Styles struct {
	BorderColor lipgloss.Color
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		default:
			m.GameEngine.Update(msg)
		}
	}

	return m, nil
}

func (m model) GenerateHeader() string {
	wordCount := fmt.Sprintf(
		"%v/%v",
		m.GameEngine.writedWordsCount,
		len(m.GameEngine.words),
	)

	var limitOptions []string
	var limitOptionsOutput string
	for _, o := range LimitOptions {
		s := strconv.Itoa(o)
		limitOptions = append(limitOptions, s)

		if m.GameEngine.limit == o {
			limitOptionsOutput += string(Yellow) + s + string(Reset) + " "
		} else {
			limitOptionsOutput += s + " "
		}
	}
	limitOptionsOutput += "\n"

	var spaceCount int
	if m.width > 0 {
		spaceCount = m.width - 20 - lipgloss.Width(wordCount) - lipgloss.Width(strings.Join(limitOptions, " "))
	}

	return string(Yellow) + wordCount + string(Reset) + strings.Repeat(" ", spaceCount) + limitOptionsOutput
}

func (m model) View() string {
	var output string

	output += m.GenerateHeader()

	input := m.GameEngine.input
	if input != "" {
		for index := range len(input) {
			correct, exists := m.GameEngine.inputCorrectness[index]
			if !exists || correct {
				output += string(White) + string(input[index]) + string(Reset)
			} else {
				output += string(Red) + string(m.GameEngine.target[index]) + string(Reset)
			}
		}
	}

	output += colorizedText(Gray, m.GameEngine.target[len(input):])

	if m.GameEngine.finished {
		output += fmt.Sprintf(
			"%s\n\n mpm: %.2f acc: %%%v\n Press Entere for next and ctrl+r to play again%s",
			string(Blue),
			m.GameEngine.Stats.WPM,
			m.GameEngine.Stats.ACC,
			string(Reset),
		)
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().Width(m.width-20).Render(output),
		),
	)
}

func colorizedText(color Color, text string) string {
	return fmt.Sprintf("%s%s\033[0m", color, text)
}
