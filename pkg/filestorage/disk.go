package filestorage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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

func (d Disk) CreateObject(content []byte) (string, error) {
	id := uuid.New().String()
	diskPath := path.Join(d.basepath, path.Join(strings.Split(id, "-")...))

	_, err := os.Stat(diskPath)
	if os.IsExist(err) {
		return "", err
	}

	dir := filepath.Dir(diskPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	file, err := os.Create(diskPath)
	if err != nil {
		return "", nil
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return "", nil
	}

	return diskPath, nil
}

func (d Disk) DeleteObject(diskPath string) error {
	_, err := os.Stat(diskPath)
	if !os.IsExist(err) {
		return nil
	}

	err = os.Remove(diskPath)

	return err
}

func (d Disk) PersistChunk(filePath string, chunk diff.DiffChunk) error {
	filePath = path.Join(d.basepath, filePath)
	switch chunk.Type {
	case diff.DiffAdd:
		return addBytesToFile(filePath, chunk.Position, chunk.Text)
	case diff.DiffRemove:
		return removeBytesFromFile(filePath, chunk.Position, chunk.Len)
	}
	return fmt.Errorf("diff type %v not supported", chunk.Type)
}

func addBytesToFile(filePath string, start int64, str string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}

	remainder, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	_, err = file.Seek(start, 0)
	if err != nil {
		return err
	}

	_, err = file.WriteString(str)
	if err != nil {
		return err
	}

	_, err = file.Write(remainder)
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
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
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
