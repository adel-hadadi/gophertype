package main

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	DefualtLimit = 25
)

var LimitOptions = []int{
	10,
	25,
	50,
	100,
}

type Engine struct {
	words            []string
	target           string
	input            string
	cursorIndex      int
	writedWordsCount int
	limit            int
	finished         bool
	startTime        time.Time
	inputCorrectness map[int]bool
	Stats            Stats
}

type Stats struct {
	WPM float64
	ACC float64
}

func NewEngine() *Engine {
	return &Engine{
		words:            make([]string, 0),
		inputCorrectness: make(map[int]bool),
		limit:            DefualtLimit,
	}
}

func (e *Engine) GenerateRandomText() {
	content, err := os.ReadFile("words.txt")
	if err != nil {
		log.Fatalf("error on reading words file: %v", err)
	}

	list := strings.Split(strings.TrimSpace(string(content)), "\n")
	selectedWords := make([]string, e.limit)

	for i := range e.limit {
		selectedWords[i] = strings.ToLower(list[rand.Intn(len(list))])
	}

	e.words = selectedWords
	e.target = strings.Join(selectedWords, " ")
}

func (e *Engine) ResetGame() {
	e.input = ""
	e.writedWordsCount = 0
	e.finished = false
	e.cursorIndex = 0
}

func (e *Engine) Next() {
	e.ResetGame()
	e.GenerateRandomText()
}

func (e *Engine) StartGame() {
	e.startTime = time.Now()
}

func (e *Engine) GameFinished() {
	e.finished = true

	e.ReportStats()
}

func (e *Engine) ReportStats() {
	e.Stats.WPM = e.CalculateWPM(len(e.input), float64(time.Since(e.startTime).Seconds()))
	e.Stats.ACC = e.CalculateACC(e.target, e.input)
}

func (e *Engine) SetLimit(l int) {
	e.limit = l
	e.GenerateRandomText()
	e.ResetGame()
}

func (e *Engine) CalculateWPM(charsTyped int, timeInSeconds float64) float64 {
	if timeInSeconds == 0 {
		return 0
	}

	return float64(charsTyped) / 5 * (60 / timeInSeconds)
}

func (e *Engine) CalculateACC(target, input string) float64 {
	var wrong int
	for _, w := range e.inputCorrectness {
		if !w {
			wrong++
		}
	}

	return float64(100 - (100 * wrong / len(e.inputCorrectness)))
}

func (e *Engine) Update(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEnter:
		e.Next()
	case tea.KeyBackspace:
		if e.finished {
			return
		}
		if len(e.input) == 0 {
			return
		}
		if e.cursorIndex > 0 {
			e.cursorIndex--
		}
		e.input = e.input[:len(e.input)-1]
	case tea.KeyCtrlR:
		e.ResetGame()
	case tea.KeyDown, tea.KeyUp:
		m := msg.String()
		var currentLimitIndex int
		for k, v := range LimitOptions {
			if v == e.limit {
				currentLimitIndex = k
			}
		}

		if m == "up" && currentLimitIndex != len(LimitOptions)-1 {
			e.SetLimit(LimitOptions[currentLimitIndex+1])
		}

		if m == "down" && currentLimitIndex != 0 {
			e.SetLimit(LimitOptions[currentLimitIndex-1])
		}
	default:
		if e.finished {
			return
		}

		for k := range len(msg.String()) {
			word := string(msg.String()[k])

			if len(e.input) == 0 {
				e.StartGame()
			}

			if word == " " {
				e.writedWordsCount++
			}

			if word != " " || len(e.input) > 0 {
				e.input += word
				e.inputCorrectness[e.cursorIndex] = string(e.target[e.cursorIndex]) == word
			}

			e.cursorIndex++

			if len(e.input) == len(e.target) {
				e.GameFinished()
			}
		}
	}
}
