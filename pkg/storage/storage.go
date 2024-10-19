package storage

import "github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"

type Storage interface {
	PersistChunk(string, diff.DiffChunk) error
}
