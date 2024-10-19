package storage

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type Disk struct {
	basepath string
}

func NewDisk(basepath string) Disk {
	return Disk{
		basepath: basepath,
	}
}

func (d Disk) PersistChunk(filePath string, chunk diff.DiffChunk) error {
	filePath = path.Join(d.basepath, filePath)
	switch chunk.Type {
	case diff.DiffAdd:
		return addBytesFromFile(filePath, chunk.Position, chunk.Text)
	case diff.DiffRemove:
		return removeBytesFromFile(filePath, chunk.Position, chunk.Len)
	}
	return fmt.Errorf("diff type %v not supported", chunk.Type)
}

func addBytesFromFile(filePath string, start int64, str string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(string(str)))
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func removeBytesFromFile(filePath string, start, length int64) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(start+length, 0)
	if err != nil {
		return err
	}

	remainingData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(remainingData)
	if err != nil {
		return err
	}

	err = file.Truncate(start + int64(len(remainingData)))
	if err != nil {
		return err
	}

	return nil
}
