package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiff(t *testing.T) {
	text := "hello world!"
	newText := "hello!"

	diff := ComputeDiff(text, newText)
	assert.Equal(t, []DiffChunk{
		{
			Position: 5,
			Type:     DiffRemove,
			Text:     " world",
			Len:      6,
		},
	}, diff)

	diff = ComputeDiff(newText, text)
	assert.Equal(t, []DiffChunk{
		{
			Position: 5,
			Type:     DiffAdd,
			Text:     " world",
			Len:      6,
		},
	}, diff)
}

func TestApplyDiff(t *testing.T) {
	t.Run("diff add", func(t *testing.T) {
		text := "hello!"
		diffAdd := DiffChunk{
			Position: 5,
			Type:     DiffAdd,
			Text:     " world",
			Len:      6,
		}
		updatedText := ApplyDiff(text, diffAdd)
		assert.Equal(t, "hello world!", updatedText)
	})

	t.Run("diff remove", func(t *testing.T) {
		text := "hello world!"
		diffRemove := DiffChunk{
			Position: 5,
			Type:     DiffRemove,
			Text:     " world",
			Len:      6,
		}
		updatedText := ApplyDiff(text, diffRemove)
		assert.Equal(t, "hello!", updatedText)
	})
}
