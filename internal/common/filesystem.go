package common

import (
	"errors"
	"fmt"
	"os"
	"path"
)

type FileSystem interface {
	FileExists(filename string) (bool, error)
	WriteFile(filename string, content []byte) error
}

type fileSystem struct{}

func NewFileSystem() *fileSystem {
	return &fileSystem{}
}

func (f *fileSystem) FileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("error checking for file existence with 'stat': %w", err)
	}
}

func (f *fileSystem) WriteFile(filename string, content []byte) error {
	// Get filename parent path
	parentDir := path.Dir(filename)
	err := os.MkdirAll(parentDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating parent directories: %w", err)
	}

	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}
