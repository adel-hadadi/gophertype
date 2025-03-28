package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
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

//go:embed words.txt
var words embed.FS

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["./debug.log"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger := zap.Must(cfg.Build())
	defer logger.Sync()

	logger.Info("logger construction succeeded")

	p := tea.NewProgram(New(logger), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func New(logger *zap.Logger) model {
	m := model{
		styledInput:       make(map[int]bool, 0),
		logger:            logger,
		enteredWordsCount: 1,
	}

	m.generateRandomText()

	return m
}

type Styles struct {
	BorderColor lipgloss.Color
}

type model struct {
	logger            *zap.Logger
	width             int
	height            int
	finished          bool
	input             string
	styledInput       map[int]bool
	words             []string
	enteredWordsCount int8
	wpm               float64
	startTime         time.Time
	cursorIndex       int
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
		case tea.KeyEnter:
			m.input = ""
			m.enteredWordsCount = 1
			m.finished = false
			m.styledInput = make(map[int]bool, 0)
			m.cursorIndex = 0
			m.generateRandomText()

			return m, nil
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				if m.cursorIndex > 0 {
					m.cursorIndex--
				}
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlR:
			m.input = ""
			m.enteredWordsCount = 1
			m.finished = false
			m.styledInput = make(map[int]bool, 0)
			m.cursorIndex = 0

			return m, nil
		default:
			if m.finished {
				return m, nil
			}

			for k := range len(msg.String()) {
				word := string(msg.String()[k])

				if len(m.input) == 0 {
					m.startTime = time.Now()
				}

				if word == " " {
					m.enteredWordsCount++
				}

				if word != " " || len(m.input) > 0 {
					target := strings.Join(m.words, " ")
					m.input += word

					m.styledInput[m.cursorIndex] = string(target[m.cursorIndex]) == word
				}

				m.cursorIndex++

				if len(m.input) == len(strings.Join(m.words, " ")) {
					m.finished = true
					m.wpm = CalculateWPM(len(m.input), float64(time.Since(m.startTime).Seconds()))
				}
			}

			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	var output string
	output += string(Yellow) + fmt.Sprintf("%v/%v", m.enteredWordsCount, len(m.words)) + string(Reset) + "\n"

	target := strings.Join(m.words, " ")

	if m.input != "" {
		for index := range len(m.input) {
			correct, exists := m.styledInput[index]
			if !exists || correct {
				output += string(m.input[index])
			} else {
				output += string(Red) + string(target[index]) + string(Reset)
			}
		}
	}

	output += colorizedText(Gray, target[len(m.input):])

	if m.finished {
		var wrong int
		for _, w := range m.styledInput {
			if !w {
				wrong++
			}
		}

		output += fmt.Sprintf(
			"%s\n\n mpm: %.2f acc: %%%v\n Press Entere for next and ctrl+r to play again%s",
			string(Blue),
			m.wpm,
			100-(100*wrong/len(m.styledInput)),
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
			output,
		),
	)
}

func (m *model) generateRandomText() {
	content, err := words.ReadFile("words.txt")
	if err != nil {
		log.Fatalf("error on reading words file: %v", err)
	}

	list := strings.Split(strings.TrimSpace(string(content)), "\n")

	selectedWords := make([]string, 10)

	for i := range 10 {
		selectedWords[i] = strings.ToLower(list[rand.Intn(len(list))])
	}

	m.words = selectedWords
}

func colorizedText(color Color, text string) string {
	return fmt.Sprintf("%s%s\033[0m", color, text)
}

func CalculateWPM(charsTyped int, timeInSeconds float64) float64 {
	if timeInSeconds == 0 {
		return 0
	}
	return (float64(charsTyped) / 5) * (60 / timeInSeconds)
}
