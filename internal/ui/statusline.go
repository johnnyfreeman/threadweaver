package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/user/editor/internal/editor"
)

/*
Status line styles using Lipgloss for terminal coloring. Each mode gets distinct
colors for visual feedback. Styles are pre-computed for performance.
*/
var (
	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252"))

	modeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	modeInsertStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("29")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	modeVisualStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("172")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			PaddingLeft(1)

	positionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Right)

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
)

/*
RenderStatusLine creates the bottom status bar showing mode, filename, and cursor position.
Layouts components with mode indicator on left, filename center-left, and position on right.
Highlights dirty state with color change and [+] indicator.
*/
func RenderStatusLine(width int, ed *editor.Editor) string {
	mode := ed.GetMode()
	buffer := ed.GetBuffer()
	cursor := ed.GetCursor()

	var modeText string
	var style lipgloss.Style

	switch mode {
	case editor.ModeNormal:
		modeText = " NORMAL "
		style = modeStyle
	case editor.ModeInsert:
		modeText = " INSERT "
		style = modeInsertStyle
	case editor.ModeVisual:
		modeText = " VISUAL "
		style = modeVisualStyle
	case editor.ModeCommand:
		modeText = " COMMAND "
		style = modeStyle
	}

	modeBlock := style.Render(modeText)

	filename := buffer.GetFilename()
	if buffer.IsDirty() {
		filename = dirtyStyle.Render(filename + " [+]")
	}
	fileBlock := fileStyle.Render(filename)

	position := fmt.Sprintf("%d:%d", cursor.Line+1, cursor.Col+1)
	posBlock := positionStyle.Render(position)

	leftContent := lipgloss.JoinHorizontal(lipgloss.Top, modeBlock, fileBlock)
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(posBlock)

	padding := width - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	middleSpace := lipgloss.NewStyle().Width(padding).Render("")

	fullLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		middleSpace,
		posBlock,
	)

	return statusStyle.Width(width).Render(fullLine)
}