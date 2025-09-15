package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/user/editor/internal/editor"
	"github.com/user/editor/internal/ui"
)

type model struct {
	editor       *editor.Editor
	renderer     *ui.Renderer
	width        int
	height       int
	scrollOffset int
	quitting     bool
}

func initialModel(filename string) model {
	ed := editor.New()
	if filename != "" {
		if err := ed.LoadFile(filename); err != nil {
			log.Printf("Error loading file: %v", err)
		}
	}

	return model{
		editor:       ed,
		renderer:     ui.NewRenderer(80, 24),
		width:        80,
		height:       24,
		scrollOffset: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.renderer.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mode := m.editor.GetMode()

	switch mode {
	case editor.ModeNormal:
		return m.handleNormalMode(msg)
	case editor.ModeInsert:
		return m.handleInsertMode(msg)
	case editor.ModeVisual:
		return m.handleVisualMode(msg)
	case editor.ModeCommand:
		return m.handleCommandMode(msg)
	}

	return m, nil
}

func (m model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if !m.editor.GetBuffer().IsDirty() {
			m.quitting = true
			return m, tea.Quit
		}

	case "h", "left":
		m.editor.MoveCursor(0, -1)
	case "j", "down":
		m.editor.MoveCursor(1, 0)
	case "k", "up":
		m.editor.MoveCursor(-1, 0)
	case "l", "right":
		m.editor.MoveCursor(0, 1)

	case "w":
		m.editor.MoveWordForward()
	case "b":
		m.editor.MoveWordBackward()

	case "0":
		m.editor.MoveToLineStart()
	case "$":
		m.editor.MoveToLineEnd()

	case "i":
		m.editor.SetMode(editor.ModeInsert)
	case "a":
		m.editor.MoveCursor(0, 1)
		m.editor.SetMode(editor.ModeInsert)

	case "v":
		m.editor.SetMode(editor.ModeVisual)

	case "x":
		m.editor.SelectLine()
		m.editor.SetMode(editor.ModeVisual)

	case "d":
		if !m.editor.GetSelection().IsEmpty() {
			m.editor.DeleteSelection()
		}
	case "c":
		if !m.editor.GetSelection().IsEmpty() {
			m.editor.DeleteSelection()
			m.editor.SetMode(editor.ModeInsert)
		}
	case "y":
		m.editor.YankSelection()
	case "p":
		m.editor.Paste()

	case ":":
		m.editor.SetMode(editor.ModeCommand)
		m.editor.ClearCommand()
	}

	m.scrollOffset = m.renderer.CalculateScrollOffset(m.editor.GetCursor(), m.scrollOffset)
	return m, nil
}

func (m model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editor.MoveCursor(0, -1)
		m.editor.SetMode(editor.ModeNormal)

	case "enter":
		m.editor.InsertNewline()

	case "backspace":
		m.editor.Backspace()

	case "left":
		m.editor.MoveCursor(0, -1)
	case "right":
		m.editor.MoveCursor(0, 1)
	case "up":
		m.editor.MoveCursor(-1, 0)
	case "down":
		m.editor.MoveCursor(1, 0)

	default:
		// Handle all single character input including space
		runes := []rune(msg.String())
		if len(runes) == 1 {
			m.editor.InsertChar(runes[0])
		}
	}

	m.scrollOffset = m.renderer.CalculateScrollOffset(m.editor.GetCursor(), m.scrollOffset)
	return m, nil
}

func (m model) handleVisualMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editor.SetMode(editor.ModeNormal)

	case "h", "left":
		m.editor.MoveCursor(0, -1)
	case "j", "down":
		m.editor.MoveCursor(1, 0)
	case "k", "up":
		m.editor.MoveCursor(-1, 0)
	case "l", "right":
		m.editor.MoveCursor(0, 1)

	case "w":
		m.editor.MoveWordForward()
	case "b":
		m.editor.MoveWordBackward()

	case "d":
		m.editor.DeleteSelection()
		m.editor.SetMode(editor.ModeNormal)
	case "c":
		m.editor.DeleteSelection()
		m.editor.SetMode(editor.ModeInsert)
	case "y":
		m.editor.YankSelection()
		m.editor.SetMode(editor.ModeNormal)
	}

	m.scrollOffset = m.renderer.CalculateScrollOffset(m.editor.GetCursor(), m.scrollOffset)
	return m, nil
}

func (m model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editor.SetMode(editor.ModeNormal)
		m.editor.ClearCommand()

	case "enter":
		if m.editor.ExecuteCommand() {
			m.quitting = true
			return m, tea.Quit
		}
		m.editor.SetMode(editor.ModeNormal)

	case "backspace":
		m.editor.BackspaceCommand()

	default:
		if len(msg.String()) == 1 {
			m.editor.AppendCommand(rune(msg.String()[0]))
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.View{
			Layer: lipgloss.NewCanvas(
				lipgloss.NewLayer(""),
			),
		}
	}

	// Render the editor content
	editorContent := m.renderer.Render(m.editor, m.scrollOffset)

	// Render the status line
	statusLine := ui.RenderStatusLine(m.width, m.editor)

	// Handle command mode display
	if m.editor.GetMode() == editor.ModeCommand {
		commandLine := ":" + m.editor.GetCommand()
		// Add cursor to command line
		commandLine += "█"
		// Replace last line with command
		statusLine = lipgloss.NewStyle().
			Width(m.width).
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Render(commandLine)
	}

	// Combine editor content and status line
	fullView := lipgloss.JoinVertical(
		lipgloss.Top,
		editorContent,
		statusLine,
	)

	// Create the view with layers
	view := tea.View{
		Layer: lipgloss.NewCanvas(
			lipgloss.NewLayer(fullView),
		),
	}

	// Add cursor for insert mode
	if m.editor.GetMode() == editor.ModeInsert {
		cursor := m.editor.GetCursor()
		cursorY := cursor.Line - m.scrollOffset
		if cursorY >= 0 && cursorY < m.height-2 {
			view.Cursor = tea.NewCursor(cursor.Col, cursorY)
		}
	}

	return view
}

func main() {
	filename := ""
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	p := tea.NewProgram(
		initialModel(filename),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}