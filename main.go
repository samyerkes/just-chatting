package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	myRequest = Request{
		Bearer:      "Bearer " + os.Getenv("OPENAI_API_KEY"),
		ContentType: "application/json",
		Endpoint:    "https://api.openai.com/v1/chat/completions",
		Method:      "POST",
	}
	myData = Data{
		Model:    "gpt-3.5-turbo",
		Messages: []Message{},
	}
	header = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render
	help   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render
)

func main() {
	p := tea.NewProgram(initialApp(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("could not start program: %v", err)
		os.Exit(1)
	}
}

type qa struct {
	question string
	answer   string
}

type app struct {
	altscreen bool
	qas       []qa
	quitting  bool
	textInput textinput.Model
	viewport  viewport.Model
}

func initialApp() app {
	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return app{
		textInput: ti,
		altscreen: true,
		qas:       []qa{},
	}
}

func (m app) Init() tea.Cmd {
	return textinput.Blink
}

func (m app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "esc", "tab":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+f":
			var cmd tea.Cmd
			if m.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}
			m.altscreen = !m.altscreen
			return m, cmd

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter":
			newQa := qa{
				question: m.textInput.Value(),
			}
			response := SendPrompt(m.textInput.Value())
			newQa.answer = response
			m.qas = append(m.qas, newQa)
			m.textInput.SetValue("")
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m app) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	// header
	s := header("OpenAI GPT-3 Chatbot")
	s += "\n"
	// footer
	s += help("Quit (ESC / CTRL+C) | Fullscreen (CTRL+F)\n")
	s += "\n"

	// Display the questions and answers
	for _, qa := range m.qas {
		if len(qa.question) > 1 {
			s += "\n"
		}
		s += fmt.Sprintf("YOU: %s", qa.question)
		s += fmt.Sprintf("\nAI: %s", qa.answer)
	}
	if len(m.qas) > 0 {
		s += "\n\n"
	}

	// body
	s += m.textInput.View()
	s += "\n\n"

	return s
}
