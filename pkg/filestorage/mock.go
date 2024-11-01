package filestorage

import (
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/stretchr/testify/mock"
)

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) PersistChunk(p string, d diff.DiffChunk) error {
	args := m.Called(p, d)
	return args.Error(0)
}

func (m *MockFileStorage) CreateObject(c []byte) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}
