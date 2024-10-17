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
	Position int
	Type     Operation
	Text     string
	Len      int
}

func ComputeDiff(oldText, newText string) []DiffChunk {
	var diffChunks []DiffChunk

	dmp := diffmatchpatch.New()

	idx := 1
	diffs := dmp.DiffMain(oldText, newText, true)
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffAdd,
				Position: idx,
				Text:     diff.Text,
				Len:      len(diff.Text),
			})
			idx += len(diff.Text) - 1
		case diffmatchpatch.DiffDelete:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffRemove,
				Position: idx,
				Text:     diff.Text,
				Len:      len(diff.Text),
			})
		case diffmatchpatch.DiffEqual:
			idx += len(diff.Text) - 1
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
		return text[:diff.Position] + diff.Text + text[diff.Position:]
	case DiffRemove:
		if text == "" {
			return ""
		}
		return text[:diff.Position] + text[diff.Position+diff.Len:]
	}
	panic("not reachable")
}
