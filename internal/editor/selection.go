package editor

type Position struct {
	Line int
	Col  int
}

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

func (s *Selection) ExtendTo(pos Position) {
	s.Head = pos
}

func (s *Selection) Collapse() {
	s.Anchor = s.Head
}