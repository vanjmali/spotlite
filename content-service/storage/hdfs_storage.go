package storage

import (
	"fmt"
	"io"
	"path"
	"time"

	"github.com/colinmarc/hdfs/v2"
	"github.com/vanjmali/spotlite/common-lib/utils"
)

type HDFSStorage struct {
	c    *hdfs.Client
	base string
}

func NewHDFSStorage() (*HDFSStorage, error) {
	uri := utils.MustGetEnv("HDFS_URI")
	base := utils.MustGetEnv("HDFS_AUDIO_BASE")

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

	now := time.Now().UnixNano()
	finalPath = path.Join(s.base, fmt.Sprintf("%s-%d%s", songID, now, ext))
	tmpPath := finalPath + fmt.Sprintf(".uploading-%d", now)

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
