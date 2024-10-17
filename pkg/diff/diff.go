package diff

import (
	"fmt"

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
}

func ComputeDiff(newText, oldText string) []DiffChunk {
	var diffChunks []DiffChunk

	dmp := diffmatchpatch.New()

	idx := 0
	diffs := dmp.DiffMain(newText, oldText, true)
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffAdd,
				Position: idx,
				Text:     diff.Text,
			})
			idx += len(diff.Text) - 1
		case diffmatchpatch.DiffDelete:
			diffChunks = append(diffChunks, DiffChunk{
				Type:     DiffRemove,
				Position: idx,
				Text:     diff.Text,
			})
		case diffmatchpatch.DiffEqual:
			idx += len(diff.Text) - 1
		}
	}

	fmt.Println(diffs)
	fmt.Println(dmp.DiffPrettyText(diffs))

	return diffChunks
}

