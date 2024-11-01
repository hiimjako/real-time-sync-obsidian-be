package filestorage

import (
	"crypto/sha256"
	"fmt"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type Storage interface {
	// PersistChunk adds a chunk to the provided filepath
	PersistChunk(string, diff.DiffChunk) error
	// CreateObject creates an object and returns the path
	CreateObject([]byte) (string, error)
	// DeleteObject deletes an object
	DeleteObject(string) error
}

func GenerateHash(content []byte) string {
	hash := sha256.New()
	hash.Write(content)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum
}
