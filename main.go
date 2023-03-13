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
	green   = lipgloss.Color("36")
	gray    = lipgloss.Color("8")
	purple  = lipgloss.Color("12")
	pink    = lipgloss.Color("201")
	white   = lipgloss.Color("#ffffff")
	header  = lipgloss.NewStyle().Bold(true).Foreground(green).Render
	help    = lipgloss.NewStyle().Foreground(gray).Render
	chat    = lipgloss.NewStyle().BorderForeground(green).BorderStyle(lipgloss.NormalBorder()).Padding(1).Width(90)
	ai      = lipgloss.NewStyle().Bold(true).Foreground(purple)
	aiText  = lipgloss.NewStyle().Foreground(white)
	you     = lipgloss.NewStyle().Bold(true).Foreground(pink)
	youText = lipgloss.NewStyle().Foreground(pink)
)

func main() {
	p := tea.NewProgram(initialApp(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("could not start program: %v", err)
		os.Exit(1)
	}
}

type qa struct {
	answer   string
	question string
}

type app struct {
	altscreen bool
	height    int
	qas       []qa
	quitting  bool
	textInput textinput.Model
	viewport  viewport.Model
	width     int
}

func initialApp() app {
	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()

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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "esc", "tab":
			m.quitting = true
			return m, tea.Quit

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter":
			newQa := qa{
				question: m.textInput.Value(),
			}
			response := SendPrompt(m.textInput.Value())
			newQa.answer = aiText.Render(response)
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

	var s string
	// Display the questions and answers
	for _, qa := range m.qas {
		if len(qa.question) > 1 {
			s += "\n"
		}
		s += you.Render("YOU:") + " " + youText.Render(qa.question) + "\n"
		s += ai.Render("AI:") + " " + qa.answer
	}
	if len(m.qas) > 0 {
		s += "\n\n"
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		lipgloss.JoinVertical(
			lipgloss.Left,
			header("OpenAI GPT-3 Chatbot"),
			chat.Render(s),
			chat.Render(m.textInput.View()),
			help("Quit (ESC / CTRL+C)"),
		),
	)
}
