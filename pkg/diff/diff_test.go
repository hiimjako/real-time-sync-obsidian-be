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
	tests := []struct {
		name     string
		diff     DiffChunk
		text     string
		expected string
	}{
		{
			name: "add a chunk",
			diff: DiffChunk{
				Position: 5,
				Type:     DiffAdd,
				Text:     " world",
				Len:      6,
			},
			text:     "hello!",
			expected: "hello world!",
		},
		{
			name: "add a chunk from empty string",
			diff: DiffChunk{
				Position: 5,
				Type:     DiffAdd,
				Text:     " world",
				Len:      6,
			},
			text:     "",
			expected: " world",
		},
		{
			name: "remove a chunk",
			diff: DiffChunk{
				Position: 5,
				Type:     DiffRemove,
				Text:     " world",
				Len:      6,
			},
			text:     "hello world!",
			expected: "hello!",
		},
		{
			name: "remove a chunk from empty string",
			diff: DiffChunk{
				Position: 5,
				Type:     DiffRemove,
				Text:     " world",
				Len:      6,
			},
			text:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ApplyDiff(tt.text, tt.diff))
		})
	}
}
