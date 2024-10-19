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
					Position: 4,
					Type:     DiffRemove,
					Text:     " world",
					Len:      6,
				},
			},
		},
		{
			name:   "compute add chunk",
			text:   "hello!",
			update: "hello world!",
			expected: []DiffChunk{
				{
					Position: 4,
					Type:     DiffAdd,
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
				Position: 4,
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
				Position: 4,
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
				Position: 4,
				Type:     DiffRemove,
				Text:     " world",
				Len:      6,
			},
			text:     "",
			expected: "",
		},
		{
			name: "remove a chunk from single word",
			diff: DiffChunk{
				Position: 0,
				Type:     DiffRemove,
				Text:     "a",
				Len:      1,
			},
			text:     "a",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ApplyDiff(tt.text, tt.diff))
		})
	}
}
