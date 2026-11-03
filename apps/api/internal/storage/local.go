package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStore struct {
	base string
}

func NewLocal(base string) (*LocalStore, error) {
	if base == "" {
		base = "./uploads"
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{base: base}, nil
}

func (s *LocalStore) path(key string) string {
	return filepath.Join(s.base, filepath.FromSlash(key))
}

func (s *LocalStore) Put(key string, r io.Reader, _ int64, _ string) error {
	p := s.path(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStore) Get(key string) (io.ReadCloser, string, error) {
	p := s.path(key)
	f, err := os.Open(p)
	if err != nil {
		return nil, "", fmt.Errorf("open: %w", err)
	}
	return f, "application/octet-stream", nil
}
