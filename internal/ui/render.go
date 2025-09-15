package ui

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/user/editor/internal/editor"
)

/*
Renderer handles terminal output using Ultraviolet's cell-based rendering.
Converts editor state into visual representation with selection highlighting
and cursor positioning using screen buffers for precise cell manipulation.
*/
type Renderer struct {
	width  int
	height int
}

func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		width:  width,
		height: height,
	}
}

func (r *Renderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

/*
Render transforms editor state into terminal output. Creates viewport of visible
lines, applies selection highlighting in visual mode, and adds block cursor for
normal/visual modes. Uses Ultraviolet screen buffers for cell-level styling.
*/
func (r *Renderer) Render(ed *editor.Editor, scrollOffset int) string {
	buffer := ed.GetBuffer()
	cursor := ed.GetCursor()
	selection := ed.GetSelection()

	viewportHeight := r.height - 2 // Leave room for status bar

	var lines []string

	for y := 0; y < viewportHeight; y++ {
		lineNum := y + scrollOffset
		if lineNum >= buffer.LineCount() {
			// Empty line
			lines = append(lines, "")
			continue
		}

		line := buffer.GetLine(lineNum)

		// Pad line to ensure cursor can be rendered anywhere
		// In normal mode, cursor can be at most len(line)-1
		// but we need to ensure there's at least one character
		if line == "" {
			line = " " // Ensure empty lines have at least a space
		} else if ed.GetMode() == editor.ModeNormal && cursor.Line == lineNum {
			// Make sure line is long enough for cursor position
			for len(line) <= cursor.Col {
				line += " "
			}
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	// Apply selection highlighting if in visual mode
	if ed.GetMode() == editor.ModeVisual && !selection.IsEmpty() {
		content = r.applySelection(content, selection, scrollOffset)
		// Also show cursor at the head of selection in visual mode
		content = r.applyCursor(content, cursor, scrollOffset, ed.GetMode())
	} else {
		// Apply cursor highlighting in normal mode
		content = r.applyCursor(content, cursor, scrollOffset, ed.GetMode())
	}

	return content
}

/*
applySelection overlays reverse-video styling on selected text regions.
Handles both single-line and multi-line selections by iterating through
affected cells and applying style changes at the cell level.
*/
func (r *Renderer) applySelection(content string, sel editor.Selection, scrollOffset int) string {
	// Create a screen buffer
	area := uv.Rect(0, 0, r.width, r.height-2)
	scr := uv.NewScreenBuffer(area.Dx(), area.Dy())

	// Draw the content
	uv.NewStyledString(content).Draw(scr, area)

	// Calculate selection area adjusted for scroll
	start, end := sel.Start(), sel.End()
	selArea := uv.Rectangle{
		Min: uv.Pos(start.Col, start.Line-scrollOffset),
		Max: uv.Pos(end.Col, end.Line-scrollOffset),
	}
	selArea = selArea.Canon()

	// Apply reverse style to selected cells
	reverseStyle := uv.NewStyle().Reverse(true)
	for y := range scr.Height() {
		if y >= selArea.Min.Y && y <= selArea.Max.Y {
			startX := 0
			endX := scr.Width()

			if selArea.Min.Y == selArea.Max.Y {
				// Single line selection
				startX = selArea.Min.X
				endX = selArea.Max.X
			} else if y == selArea.Min.Y {
				// First line of multi-line selection
				startX = selArea.Min.X
			} else if y == selArea.Max.Y {
				// Last line of multi-line selection
				endX = selArea.Max.X
			}

			for x := startX; x < endX && x < scr.Width(); x++ {
				cell := scr.CellAt(x, y)
				if cell != nil {
					cell.Style = reverseStyle
					scr.SetCell(x, y, cell)
				}
			}
		}
	}

	return scr.Render()
}

/*
applyCursor renders block cursor for normal/visual modes by reversing the cell
at cursor position. Insert/command modes use native terminal line cursor instead.
Creates empty cell with space if cursor is beyond line content.
*/
func (r *Renderer) applyCursor(content string, cursor editor.Position, scrollOffset int, mode editor.Mode) string {
	// Don't show block cursor in insert/command mode (they use line cursor)
	if mode == editor.ModeInsert || mode == editor.ModeCommand {
		return content
	}

	// Create a screen buffer
	area := uv.Rect(0, 0, r.width, r.height-2)
	scr := uv.NewScreenBuffer(area.Dx(), area.Dy())

	// Draw the content
	uv.NewStyledString(content).Draw(scr, area)

	// Apply cursor
	cursorY := cursor.Line - scrollOffset
	if cursorY >= 0 && cursorY < scr.Height() {
		cell := scr.CellAt(cursor.Col, cursorY)
		if cell != nil {
			// Clone the cell and apply reverse style
			cell = cell.Clone()
			cell.Style = uv.NewStyle().Reverse(true)
			scr.SetCell(cursor.Col, cursorY, cell)
		} else {
			// No cell at cursor position, create one with a space
			newCell := &uv.Cell{
				Content: " ",
				Width:   1,
				Style:   uv.NewStyle().Reverse(true),
			}
			scr.SetCell(cursor.Col, cursorY, newCell)
		}
	}

	return scr.Render()
}

func (r *Renderer) CalculateScrollOffset(cursor editor.Position, currentOffset int) int {
	viewportHeight := r.height - 2

	if cursor.Line < currentOffset {
		return cursor.Line
	} else if cursor.Line >= currentOffset+viewportHeight {
		return cursor.Line - viewportHeight + 1
	}

	return currentOffset
}