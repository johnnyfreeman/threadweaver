package editor

import (
	"strings"
)

/*
Editor coordinates the editing state, integrating buffer management, cursor control,
selection handling, and mode switching. Acts as the central state machine for all
editing operations, maintaining consistency between cursor, selection, and buffer state.
*/
type Editor struct {
	buffer    *Buffer
	cursor    Position
	selection Selection
	mode      Mode
	command   string
	clipboard string
}

func New() *Editor {
	return &Editor{
		buffer:    NewBuffer(),
		cursor:    Position{Line: 0, Col: 0},
		selection: NewSelection(Position{Line: 0, Col: 0}),
		mode:      ModeNormal,
	}
}

func (e *Editor) LoadFile(filename string) error {
	return e.buffer.LoadFile(filename)
}

func (e *Editor) SaveFile() error {
	return e.buffer.SaveFile()
}

func (e *Editor) GetMode() Mode {
	return e.mode
}

func (e *Editor) SetMode(mode Mode) {
	e.mode = mode
	if mode != ModeVisual {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) GetCursor() Position {
	return e.cursor
}

func (e *Editor) GetSelection() Selection {
	return e.selection
}

func (e *Editor) GetBuffer() *Buffer {
	return e.buffer
}

func (e *Editor) GetCommand() string {
	return e.command
}

/*
clampCursor ensures cursor stays within valid buffer bounds. In normal mode,
cursor cannot move past the last character. In insert mode, cursor can be
positioned after the last character for end-of-line insertion.
*/
func (e *Editor) clampCursor() {
	if e.cursor.Line < 0 {
		e.cursor.Line = 0
	}
	if e.cursor.Line >= e.buffer.LineCount() {
		e.cursor.Line = e.buffer.LineCount() - 1
	}
	if e.cursor.Line < 0 {
		e.cursor.Line = 0
	}

	lineLen := len(e.buffer.GetLine(e.cursor.Line))
	if e.mode == ModeNormal && lineLen > 0 {
		if e.cursor.Col >= lineLen {
			e.cursor.Col = lineLen - 1
		}
	} else {
		if e.cursor.Col > lineLen {
			e.cursor.Col = lineLen
		}
	}
	if e.cursor.Col < 0 {
		e.cursor.Col = 0
	}
}

func (e *Editor) MoveCursor(dLine, dCol int) {
	e.cursor.Line += dLine
	e.cursor.Col += dCol
	e.clampCursor()

	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) MoveCursorTo(pos Position) {
	e.cursor = pos
	e.clampCursor()

	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) MoveToLineStart() {
	e.cursor.Col = 0
	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) MoveToLineEnd() {
	lineLen := len(e.buffer.GetLine(e.cursor.Line))
	if e.mode == ModeNormal && lineLen > 0 {
		e.cursor.Col = lineLen - 1
	} else {
		e.cursor.Col = lineLen
	}
	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

/*
MoveWordForward implements word-wise forward navigation. Words are defined as
contiguous alphanumeric characters plus underscore. Moves to next line if at
end of current line. Updates selection in visual mode.
*/
func (e *Editor) MoveWordForward() {
	line := e.buffer.GetLine(e.cursor.Line)
	col := e.cursor.Col

	inWord := false
	for col < len(line) {
		isAlphaNum := (line[col] >= 'a' && line[col] <= 'z') ||
			(line[col] >= 'A' && line[col] <= 'Z') ||
			(line[col] >= '0' && line[col] <= '9') ||
			line[col] == '_'

		if !inWord && isAlphaNum {
			inWord = true
		} else if inWord && !isAlphaNum {
			break
		}
		col++
	}

	if col == e.cursor.Col && e.cursor.Line < e.buffer.LineCount()-1 {
		e.cursor.Line++
		e.cursor.Col = 0
	} else {
		e.cursor.Col = col
	}

	e.clampCursor()
	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) MoveWordBackward() {
	if e.cursor.Col == 0 {
		if e.cursor.Line > 0 {
			e.cursor.Line--
			e.cursor.Col = len(e.buffer.GetLine(e.cursor.Line))
		}
	} else {
		line := e.buffer.GetLine(e.cursor.Line)
		col := e.cursor.Col - 1

		for col > 0 && line[col] == ' ' {
			col--
		}

		for col > 0 {
			isAlphaNum := (line[col-1] >= 'a' && line[col-1] <= 'z') ||
				(line[col-1] >= 'A' && line[col-1] <= 'Z') ||
				(line[col-1] >= '0' && line[col-1] <= '9') ||
				line[col-1] == '_'

			currentIsAlphaNum := (line[col] >= 'a' && line[col] <= 'z') ||
				(line[col] >= 'A' && line[col] <= 'Z') ||
				(line[col] >= '0' && line[col] <= '9') ||
				line[col] == '_'

			if !isAlphaNum && currentIsAlphaNum {
				break
			}
			col--
		}

		e.cursor.Col = col
	}

	e.clampCursor()
	if e.mode == ModeVisual {
		e.selection.ExtendTo(e.cursor)
	} else {
		e.selection = NewSelection(e.cursor)
	}
}

func (e *Editor) SelectLine() {
	e.selection.Anchor = Position{Line: e.cursor.Line, Col: 0}
	lineLen := len(e.buffer.GetLine(e.cursor.Line))
	e.selection.Head = Position{Line: e.cursor.Line, Col: lineLen}
	e.cursor = e.selection.Head
}

func (e *Editor) SelectWord() {
	line := e.buffer.GetLine(e.cursor.Line)
	if e.cursor.Col >= len(line) {
		return
	}

	start := e.cursor.Col
	end := e.cursor.Col

	isAlphaNum := func(ch byte) bool {
		return (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_'
	}

	if !isAlphaNum(line[start]) {
		return
	}

	for start > 0 && isAlphaNum(line[start-1]) {
		start--
	}

	for end < len(line) && isAlphaNum(line[end]) {
		end++
	}

	e.selection.Anchor = Position{Line: e.cursor.Line, Col: start}
	e.selection.Head = Position{Line: e.cursor.Line, Col: end}
	e.cursor = e.selection.Head
}

func (e *Editor) DeleteSelection() {
	if !e.selection.IsEmpty() {
		e.cursor = e.buffer.DeleteSelection(e.selection)
		e.selection = NewSelection(e.cursor)
		e.clampCursor()
	}
}

func (e *Editor) YankSelection() {
	if !e.selection.IsEmpty() {
		e.clipboard = e.buffer.GetSelectedText(e.selection)
	}
}

func (e *Editor) Paste() {
	if e.clipboard == "" {
		return
	}

	lines := strings.Split(e.clipboard, "\n")
	if len(lines) == 1 {
		for _, ch := range e.clipboard {
			e.cursor = e.buffer.InsertChar(e.cursor, ch)
		}
	} else {
		for i, line := range lines {
			for _, ch := range line {
				e.cursor = e.buffer.InsertChar(e.cursor, ch)
			}
			if i < len(lines)-1 {
				e.cursor = e.buffer.InsertNewline(e.cursor)
			}
		}
	}
	e.selection = NewSelection(e.cursor)
}

func (e *Editor) InsertChar(ch rune) {
	e.cursor = e.buffer.InsertChar(e.cursor, ch)
	e.selection = NewSelection(e.cursor)
}

func (e *Editor) InsertNewline() {
	e.cursor = e.buffer.InsertNewline(e.cursor)
	e.selection = NewSelection(e.cursor)
}

func (e *Editor) Backspace() {
	e.cursor = e.buffer.DeleteChar(e.cursor)
	e.selection = NewSelection(e.cursor)
	e.clampCursor()
}

func (e *Editor) AppendCommand(ch rune) {
	e.command += string(ch)
}

func (e *Editor) BackspaceCommand() {
	if len(e.command) > 0 {
		e.command = e.command[:len(e.command)-1]
	}
}

func (e *Editor) ClearCommand() {
	e.command = ""
}

/*
ExecuteCommand processes command-line input. Returns true if the command
requests editor termination. Implements basic vim-style commands for file
operations and quitting.
*/
func (e *Editor) ExecuteCommand() bool {
	cmd := strings.TrimSpace(e.command)
	e.command = ""

	switch cmd {
	case "w":
		e.SaveFile()
		return false
	case "q":
		if !e.buffer.IsDirty() {
			return true
		}
	case "q!":
		return true
	case "wq":
		e.SaveFile()
		return true
	}

	return false
}