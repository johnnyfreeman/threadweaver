package editor

/*
Position represents a cursor location in the buffer using zero-indexed line and column.
The Col field is byte-based, not rune-based, simplifying buffer operations.
*/
type Position struct {
	Line int
	Col  int
}

/*
Selection implements anchor-head selection model where Anchor is the starting point
and Head is the cursor position. This allows directional selections and maintains
selection intent during cursor movement. Empty selections (Anchor == Head) represent
just the cursor position.
*/
type Selection struct {
	Anchor Position
	Head   Position
}

func NewSelection(pos Position) Selection {
	return Selection{
		Anchor: pos,
		Head:   pos,
	}
}

func (s Selection) IsEmpty() bool {
	return s.Anchor == s.Head
}

/*
Start returns the earlier position in document order, regardless of selection direction.
Used for buffer operations that need normalized ranges.
*/
func (s Selection) Start() Position {
	if s.Anchor.Line < s.Head.Line || (s.Anchor.Line == s.Head.Line && s.Anchor.Col < s.Head.Col) {
		return s.Anchor
	}
	return s.Head
}

func (s Selection) End() Position {
	if s.Anchor.Line > s.Head.Line || (s.Anchor.Line == s.Head.Line && s.Anchor.Col > s.Head.Col) {
		return s.Anchor
	}
	return s.Head
}

func (s Selection) Contains(pos Position) bool {
	start, end := s.Start(), s.End()
	if pos.Line < start.Line || pos.Line > end.Line {
		return false
	}
	if pos.Line == start.Line && pos.Col < start.Col {
		return false
	}
	if pos.Line == end.Line && pos.Col > end.Col {
		return false
	}
	return true
}

/*
ExtendTo moves the head while keeping anchor fixed, used in visual mode
to grow/shrink selections as the cursor moves.
*/
func (s *Selection) ExtendTo(pos Position) {
	s.Head = pos
}

func (s *Selection) Collapse() {
	s.Anchor = s.Head
}