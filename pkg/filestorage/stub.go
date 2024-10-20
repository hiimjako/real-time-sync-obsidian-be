package filestorage

import (
	"sync"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type FileStorageStub struct {
	files map[string]string
	mu    sync.Mutex
}

func NewStorageStub() *FileStorageStub {
	return &FileStorageStub{
		files: map[string]string{},
	}
}

func (s *FileStorageStub) PersistChunk(file string, d diff.DiffChunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[file] = diff.ApplyDiff(s.files[file], d)
	return nil
}

func (s *FileStorageStub) Files() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	cpy := make(map[string]string, len(s.files))

	for k, v := range s.files {
		cpy[k] = v
	}

	return cpy
}
