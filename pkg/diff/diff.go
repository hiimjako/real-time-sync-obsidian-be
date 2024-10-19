package diff

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Operation int8

const (
	DiffRemove Operation = -1
	DiffAdd    Operation = 1
)

type DiffChunk struct {
	Type Operation
	// Position indicates the position immediately after the last valid character, where the diff should start being applied.
	Position int64
	Text     string
	Len      int64
}

func ComputeDiff(oldText, newText string) []DiffChunk {
	var diffChunks []DiffChunk

	dmp := diffmatchpatch.New()

	var idx int64
	diffs := dmp.DiffMain(oldText, newText, true)
	for _, diff := range diffs {
		l := int64(len(diff.Text))
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffAdd,
				Position: idx,
				Text:     diff.Text,
				Len:      l,
			})
			idx += int64(l) - 1
		case diffmatchpatch.DiffDelete:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffRemove,
				Position: idx,
				Text:     diff.Text,
				Len:      l,
			})
		case diffmatchpatch.DiffEqual:
			idx += l - 1
		}
	}

	return diffChunks
}

func ApplyDiff(text string, diff DiffChunk) string {
	switch diff.Type {
	case DiffAdd:
		if text == "" {
			return diff.Text
		}

		if diff.Position == 0 {
			return text[diff.Position:] + diff.Text
		}

		return text[:diff.Position+1] + diff.Text + text[diff.Position+1:]
	case DiffRemove:
		if text == "" {
			return ""
		}

		if diff.Position == 0 {
			return text[diff.Len:]
		}

		return text[:diff.Position+1] + text[diff.Position+diff.Len+1:]
	}
	panic("not reachable")
}
