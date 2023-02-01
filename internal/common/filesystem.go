package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
)

type FileSystem interface {
	FileExists(filename string) (bool, error)
	WriteFile(filename string, content []byte) error
	ReadFileLines(filename string) ([]string, error)
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
	err := os.MkdirAll(parentDir, 0o755)
	if err != nil {
		return fmt.Errorf("error creating parent directories: %w", err)
	}

	err = os.WriteFile(filename, content, 0o644)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}

func (f *fileSystem) ReadFileLines(filename string) ([]string, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}

	var l []string
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l = append(l, scanner.Text())
	}

	return l, nil
}
