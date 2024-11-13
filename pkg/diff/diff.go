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
	Type Operation `json:"type"`
	// Position indicates the position immediately after the last valid character, inclusive.
	Position int64  `json:"position"`
	Text     string `json:"text"`
	Len      int64  `json:"len"`
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
			idx += l
		case diffmatchpatch.DiffDelete:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffRemove,
				Position: idx,
				Text:     diff.Text,
				Len:      l,
			})
		case diffmatchpatch.DiffEqual:
			idx += l
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

		return text[:diff.Position] + diff.Text + text[diff.Position:]
	case DiffRemove:
		if text == "" {
			return ""
		}

		if diff.Position == 0 {
			return text[diff.Len:]
		}

		return text[:diff.Position] + text[diff.Position-1+diff.Len+1:]
	}
	panic("not reachable")
}
