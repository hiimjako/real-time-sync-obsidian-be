package filestorage

import (
	"crypto/sha256"
	"fmt"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type Storage interface {
	// PersistChunk adds a chunk to the provided filepath
	PersistChunk(string, diff.DiffChunk) error
	// CreateObject creates an object to the provided filepath
	CreateObject(string, []byte) (string, error)
}

func CalculateHash(content []byte) string {
	hash := sha256.New()
	hash.Write(content)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum
}
