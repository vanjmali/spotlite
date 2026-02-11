package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/colinmarc/hdfs/v2"
)

type HDFSStorage struct {
	c    *hdfs.Client
	base string
}

func NewHDFSStorage() (*HDFSStorage, error) {
	uri := os.Getenv("HDFS_URI")
	if uri == "" {
		return nil, errors.New("HDFS_URI is not set")
	}
	base := os.Getenv("HDFS_AUDIO_BASE")
	if base == "" {
		base = "/spotlite/audio"
	}

	c, err := hdfs.New(uri)
	if err != nil {
		return nil, err
	}

	return &HDFSStorage{c: c, base: base}, nil
}

func (s *HDFSStorage) EnsureBaseDir() error {
	return s.c.MkdirAll(s.base, 0o755)
}

func (s *HDFSStorage) UploadSongAudio(songID string, r io.Reader, ext string) (finalPath string, size int64, err error) {
	if ext == "" {
		ext = ".bin"
	}

	finalPath = path.Join(s.base, fmt.Sprintf("%s%s", songID, ext))
	tmpPath := finalPath + fmt.Sprintf(".uploading-%d", time.Now().UnixNano())

	w, err := s.c.Create(tmpPath)
	if err != nil {
		return "", 0, err
	}

	n, copyErr := io.Copy(w, r)
	closeErr := w.Close()
	if copyErr != nil {
		_ = s.c.Remove(tmpPath)
		return "", 0, copyErr
	}
	if closeErr != nil {
		_ = s.c.Remove(tmpPath)
		return "", 0, closeErr
	}

	if err := s.c.Rename(tmpPath, finalPath); err != nil {
		_ = s.c.Remove(tmpPath)
		return "", 0, err
	}
	return finalPath, n, nil
}

func (s *HDFSStorage) Open(p string) (io.ReadCloser, error) {
	return s.c.Open(p)
}

func (s *HDFSStorage) Close() error {
	return s.c.Close()
}

func (s *HDFSStorage) Remove(path string) error {
	return s.c.Remove(path)
}
