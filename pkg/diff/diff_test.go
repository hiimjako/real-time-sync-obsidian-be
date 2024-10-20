package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		update   string
		expected []DiffChunk
	}{
		{
			name:   "compute remove chunk",
			text:   "hello world!",
			update: "hello!",
			expected: []DiffChunk{
				{
					Position: 5,
					Type:     DiffRemove,
					Text:     " world",
					Len:      6,
				},
			},
		},
		{
			name:   "compute remove chunk 2",
			text:   " ",
			update: "",
			expected: []DiffChunk{
				{
					Position: 0,
					Type:     DiffRemove,
					Text:     " ",
					Len:      1,
				},
			},
		},
		{
			name:   "compute add chunk",
			text:   "hello!",
			update: "hello world!",
			expected: []DiffChunk{
				{
					Position: 5,
					Type:     DiffAdd,
					Text:     " world",
					Len:      6,
				},
			},
		},
		{
			name:   "compute add chunk 2",
			text:   "h",
			update: "he",
			expected: []DiffChunk{
				{
					Position: 1,
					Type:     DiffAdd,
					Text:     "e",
					Len:      1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ComputeDiff(tt.text, tt.update))
		})
	}
}

func TestApplyDiff(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "add a chunk",
			text:     "hello!",
			expected: "hello world!",
		},
		{
			name:     "add a chunk from empty string",
			text:     "",
			expected: " world",
		},
		{
			name:     "add a chunk from 0",
			text:     "",
			expected: "test",
		},
		{
			name:     "remove a chunk",
			text:     "hello world!",
			expected: "helloworld!",
		},
		{
			name:     "remove a chunk from 0",
			text:     "test",
			expected: "",
		},
		{
			name:     "add in middle of word",
			text:     "wold",
			expected: "world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.text
			diffs := ComputeDiff(tt.text, tt.expected)
			for _, d := range diffs {
				s = ApplyDiff(s, d)
			}
			assert.Equal(t, tt.expected, s)
		})
	}

	t.Run("remove a chunk from empty string", func(t *testing.T) {
		assert.Equal(t, "", ApplyDiff("", DiffChunk{
			Type:     DiffRemove,
			Len:      4,
			Text:     "test",
			Position: 10,
		}))
	})
}
