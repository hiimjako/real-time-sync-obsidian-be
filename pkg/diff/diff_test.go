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
					Position: 4,
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
					Position: 0,
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
		diff     DiffChunk
		text     string
		expected string
	}{
		{
			name: "add a chunk",
			diff: DiffChunk{
				Position: 4,
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
			name: "add a chunk from 0",
			diff: DiffChunk{
				Position: 0,
				Type:     DiffAdd,
				Text:     "test",
				Len:      4,
			},
			text:     "",
			expected: "test",
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
			name: "remove a chunk from 0",
			diff: DiffChunk{
				Position: 0,
				Type:     DiffRemove,
				Text:     "test",
				Len:      4,
			},
			text:     "test",
			expected: "",
		},
		{
			name: "add in middle of word",
			diff: DiffChunk{
				Position: 2,
				Type:     DiffAdd,
				Text:     "l",
				Len:      1,
			},
			text:     "wor",
			expected: "worl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ApplyDiff(tt.text, tt.diff))
		})
	}
}
