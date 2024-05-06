package indexables

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type Range struct {
	Start Position
	End   Position
}

func NewRange(startLine uint, startChar uint, endLine uint, endChar uint) Range {
	return Range{
		Start: NewPosition(startLine, startChar),
		End:   NewPosition(endLine, endChar),
	}
}

func NewRangeFromTreeSitterPositions(start sitter.Point, end sitter.Point) Range {
	return Range{
		Start: NewPositionFromTreeSitterPoint(start),
		End:   NewPositionFromTreeSitterPoint(end),
	}
}

func (r Range) HasPosition(position Position) bool {
	line := uint(position.Line)
	ch := uint(position.Character)

	if line >= r.Start.Line && line <= r.End.Line {
		// Exactly same line
		if line == r.Start.Line && line == r.End.Line {
			// Must be inside character ranges
			if ch >= r.Start.Character && ch <= r.End.Character {
				return true
			}
		} else {
			return true
		}
	}

	return false
}
