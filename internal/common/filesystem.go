package common

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

const (
	DefaultFilePermissions fs.FileMode = 0o600
	DefaultDirPermissions  fs.FileMode = 0o755
)

type LockFile interface {
	Commit() error
	Rollback() error
}

type lockFile struct {
	filename     string
	lockFilename string
}

func (l *lockFile) Commit() error {
	return os.Rename(l.lockFilename, l.filename)
}

func (l *lockFile) Rollback() error {
	return os.Remove(l.lockFilename)
}

type FileSystem interface {
	GetLocalExecutable(name string) (string, error)

	FileExists(filename string) (bool, error)
	WriteFile(filename string, content []byte) error
	WriteLockFileFunc(filename string, writeFunc func(io.Writer) error) (LockFile, error)
	DeleteFile(filename string) (bool, error)
	ReadFileLines(filename string) ([]string, error)
}

type fileSystem struct{}

func NewFileSystem() FileSystem {
	return &fileSystem{}
}

func (f *fileSystem) createLeadingDirs(filename string) error {
	parentDir := path.Dir(filename)
	err := os.MkdirAll(parentDir, DefaultDirPermissions)
	if err != nil {
		return fmt.Errorf("error creating parent directories: %w", err)
	}

	return nil
}

func (f *fileSystem) GetLocalExecutable(name string) (string, error) {
	thisExePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get path to current executable: %w", err)
	}
	exeDir := filepath.Dir(thisExePath)
	if err != nil {
		return "", fmt.Errorf("failed to get parent dir of current executable: %w", err)
	}

	programPath := filepath.Join(exeDir, name)
	programExists, err := f.FileExists(programPath)
	if err != nil {
		return "", fmt.Errorf("could not determine whether path to '%s' exists: %w", name, err)
	} else if !programExists {
		return "", fmt.Errorf("could not find path to '%s'", name)
	}

	return programPath, nil
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
	err := f.createLeadingDirs(filename)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, content, DefaultFilePermissions)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}

func (f *fileSystem) WriteLockFileFunc(filename string, writeFunc func(io.Writer) error) (LockFile, error) {
	err := f.createLeadingDirs(filename)
	if err != nil {
		return nil, err
	}

	lockFilename := filename + ".lock"
	lock, err := os.OpenFile(lockFilename, os.O_WRONLY|os.O_CREATE, DefaultFilePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	lockFile := &lockFile{filename: filename, lockFilename: lockFilename}

	err = writeFunc(lock)
	if err != nil {
		// Try to close & rollback - don't worry about errors, we're already failing.
		lock.Close()
		lockFile.Rollback()
		return nil, err
	}

	err = lock.Close()
	if err != nil {
		// Try to rollback - don't worry about errors, we're already failing.
		lockFile.Rollback()
		return nil, fmt.Errorf("failed to close lock file: %w", err)
	}

	return lockFile, nil
}

func (f *fileSystem) DeleteFile(filename string) (bool, error) {
	err := os.Remove(filename)
	if err == nil {
		return true, nil
	}

	pathErr, ok := err.(*os.PathError)
	if ok && pathErr.Err == syscall.ENOENT {
		return false, nil
	} else {
		return false, err
	}
}

func (f *fileSystem) ReadFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		pathErr, ok := err.(*os.PathError)
		if ok && pathErr.Err == syscall.ENOENT {
			// If the file doesn't exist, return empty result rather than an
			// error
			return []string{}, nil
		} else {
			return nil, err
		}
	}

	var l []string
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l = append(l, scanner.Text())
	}

	return l, nil
}
