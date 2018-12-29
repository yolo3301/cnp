package file

import (
	"io"
	"os"
	"path/filepath"
)

type Storage struct {
	root string
}

func NewStorage(root string) (*Storage, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}
	return &Storage{root}, nil
}

func (s *Storage) Read(path string, w io.Writer) error {
	f, err := os.Open(filepath.Join(s.root, path))
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Write(path string, r io.Reader) error {
	f, err := os.Create(filepath.Join(s.root, path))
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}
