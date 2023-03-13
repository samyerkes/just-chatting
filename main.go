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
	ready     bool
	textInput textinput.Model
	viewport  viewport.Model
	width     int
}

const useHighPerformanceRenderer = false

func initialApp() app {
	ti := textinput.New()
	ti.Placeholder = "Ask a question..."
	ti.Focus()

	return app{
		altscreen: true,
		qas:       []qa{},
		ready:     false,
		textInput: ti,
	}
}

func (m app) Init() tea.Cmd {
	return textinput.Blink
}

func (m app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}

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
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
