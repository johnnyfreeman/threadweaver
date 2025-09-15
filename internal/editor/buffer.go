package editor

import (
	"os"
	"strings"
)

/*
Buffer manages text content as a slice of lines, tracking modifications and file I/O.
Line-based storage simplifies newline handling and line-oriented operations while
maintaining reasonable performance for typical editing tasks.
*/
type Buffer struct {
	lines    []string
	filename string
	dirty    bool
}

func NewBuffer() *Buffer {
	return &Buffer{
		lines: []string{""},
		dirty: false,
	}
}

func (b *Buffer) LoadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			b.lines = []string{""}
			b.filename = filename
			b.dirty = false
			return nil
		}
		return err
	}

	text := string(content)
	if text == "" {
		b.lines = []string{""}
	} else {
		b.lines = strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	}
	b.filename = filename
	b.dirty = false
	return nil
}

func (b *Buffer) SaveFile() error {
	if b.filename == "" {
		return nil
	}
	content := strings.Join(b.lines, "\n")
	err := os.WriteFile(b.filename, []byte(content), 0644)
	if err == nil {
		b.dirty = false
	}
	return err
}

func (b *Buffer) LineCount() int {
	return len(b.lines)
}

func (b *Buffer) GetLine(n int) string {
	if n < 0 || n >= len(b.lines) {
		return ""
	}
	return b.lines[n]
}

/*
InsertChar inserts a single rune at the given position. Returns the new cursor position
after insertion. Clamps column to line length to handle out-of-bounds positions gracefully.
*/
func (b *Buffer) InsertChar(pos Position, ch rune) Position {
	if pos.Line >= len(b.lines) {
		return pos
	}

	line := b.lines[pos.Line]
	if pos.Col > len(line) {
		pos.Col = len(line)
	}

	b.lines[pos.Line] = line[:pos.Col] + string(ch) + line[pos.Col:]
	b.dirty = true
	return Position{Line: pos.Line, Col: pos.Col + 1}
}

func (b *Buffer) InsertNewline(pos Position) Position {
	if pos.Line >= len(b.lines) {
		return pos
	}

	line := b.lines[pos.Line]
	if pos.Col > len(line) {
		pos.Col = len(line)
	}

	before := line[:pos.Col]
	after := line[pos.Col:]

	b.lines[pos.Line] = before
	newLines := append([]string{}, b.lines[:pos.Line+1]...)
	newLines = append(newLines, after)
	newLines = append(newLines, b.lines[pos.Line+1:]...)
	b.lines = newLines

	b.dirty = true
	return Position{Line: pos.Line + 1, Col: 0}
}

func (b *Buffer) DeleteChar(pos Position) Position {
	if pos.Line >= len(b.lines) {
		return pos
	}

	line := b.lines[pos.Line]
	if pos.Col > 0 && pos.Col <= len(line) {
		b.lines[pos.Line] = line[:pos.Col-1] + line[pos.Col:]
		b.dirty = true
		return Position{Line: pos.Line, Col: pos.Col - 1}
	} else if pos.Col == 0 && pos.Line > 0 {
		prevLine := b.lines[pos.Line-1]
		b.lines[pos.Line-1] = prevLine + line
		b.lines = append(b.lines[:pos.Line], b.lines[pos.Line+1:]...)
		b.dirty = true
		return Position{Line: pos.Line - 1, Col: len(prevLine)}
	}

	return pos
}

/*
DeleteSelection removes text within the selection range, handling both single-line
and multi-line deletions. Merges partial lines when deleting across line boundaries.
Returns cursor position at the start of the deleted range.
*/
func (b *Buffer) DeleteSelection(sel Selection) Position {
	start, end := sel.Start(), sel.End()

	if start.Line == end.Line {
		line := b.lines[start.Line]
		b.lines[start.Line] = line[:start.Col] + line[end.Col:]
	} else {
		startLine := b.lines[start.Line][:start.Col]
		endLine := b.lines[end.Line][end.Col:]

		b.lines[start.Line] = startLine + endLine
		if end.Line+1 < len(b.lines) {
			b.lines = append(b.lines[:start.Line+1], b.lines[end.Line+1:]...)
		} else {
			b.lines = b.lines[:start.Line+1]
		}
	}

	b.dirty = true
	return start
}

func (b *Buffer) GetSelectedText(sel Selection) string {
	start, end := sel.Start(), sel.End()

	if start.Line == end.Line {
		line := b.GetLine(start.Line)
		if start.Col >= len(line) || end.Col > len(line) {
			return ""
		}
		return line[start.Col:end.Col]
	}

	var result strings.Builder
	for i := start.Line; i <= end.Line && i < len(b.lines); i++ {
		line := b.GetLine(i)
		if i == start.Line {
			if start.Col < len(line) {
				result.WriteString(line[start.Col:])
			}
			result.WriteString("\n")
		} else if i == end.Line {
			if end.Col > 0 && end.Col <= len(line) {
				result.WriteString(line[:end.Col])
			}
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (b *Buffer) IsDirty() bool {
	return b.dirty
}

func (b *Buffer) GetFilename() string {
	if b.filename == "" {
		return "[No Name]"
	}
	return b.filename
}