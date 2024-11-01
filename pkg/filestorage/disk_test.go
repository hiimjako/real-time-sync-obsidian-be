package filestorage

import (
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistChunk(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		diffs    [][]diff.DiffChunk
	}{
		{
			name:     "compute remove chunk in new file",
			expected: "hello",
			diffs: [][]diff.DiffChunk{
				diff.ComputeDiff("hello", ""),
				diff.ComputeDiff("", "he__llo"),
				diff.ComputeDiff("he__llo", "hello"),
			},
		},
		{
			name:     "compute add chunk in new file",
			expected: "hello world!",
			diffs: [][]diff.DiffChunk{
				diff.ComputeDiff("", "hello"),
				diff.ComputeDiff("hello", "hello!"),
				diff.ComputeDiff("hello!", "hello world!"),
			},
		},
	}

	dir := t.TempDir()
	disk := NewDisk(dir)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileId := uuid.New().String()
			for _, d := range tt.diffs {
				for _, d2 := range d {
					assert.NoError(t, disk.PersistChunk(fileId, d2))
				}
			}

			filePath := path.Join(dir, fileId)
			fileContent, err := os.ReadFile(filePath)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(fileContent))
		})
	}
}

func TestCreateObject(t *testing.T) {
	dir := t.TempDir()
	d := NewDisk(dir)

	content := []byte("bar")
	p, err := d.CreateObject(content)
	assert.NoError(t, err)

	fileContent, err := os.ReadFile(p)
	assert.NoError(t, err)

	assert.Equal(t, content, fileContent)
}
