package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	RootDir string
}

type Object struct {
	Key         string
	Size        int64
	ContentType string
}

type LocalStore struct {
	rootDir string
}

func NewLocal(cfg Config) (*LocalStore, error) {

	if cfg.RootDir == "" {
		return nil, fmt.Errorf("storage root directory is required")
	}

	if err := os.MkdirAll(cfg.RootDir, 0o750); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}

	rootDir, err := filepath.Abs(cfg.RootDir)

	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}

	return &LocalStore{rootDir: rootDir}, nil
}

func (s *LocalStore) Put(ctx context.Context, key string, content io.Reader, contentType string) (Object, error) {

	if content == nil {
		return Object{}, fmt.Errorf("storage content is required")
	}

	path, err := s.safePath(key)

	if err != nil {
		return Object{}, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return Object{}, fmt.Errorf("create object directory: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".upload-*")

	if err != nil {
		return Object{}, fmt.Errorf("create temporary object: %w", err)
	}

	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	written, copyErr := copyWithContext(ctx, temporary, content)
	closeErr := temporary.Close()

	if copyErr != nil {
		return Object{}, copyErr
	}

	if closeErr != nil {
		return Object{}, fmt.Errorf("close temporary object: %w", closeErr)
	}

	if err := os.Rename(temporaryPath, path); err != nil {
		return Object{}, fmt.Errorf("commit object: %w", err)
	}

	return Object{Key: key, Size: written, ContentType: contentType}, nil
}

func (s *LocalStore) Open(key string) (io.ReadCloser, error) {
	path, err := s.safePath(key)

	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("open object: %w", err)
	}

	return file, nil
}

func (s *LocalStore) Delete(key string) error {
	path, err := s.safePath(key)

	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func (s *LocalStore) safePath(key string) (string, error) {

	if s == nil || s.rootDir == "" {
		return "", fmt.Errorf("storage is not configured")
	}

	cleanKey := filepath.Clean(filepath.FromSlash(key))

	if key == "" || cleanKey == "." || filepath.IsAbs(cleanKey) || strings.HasPrefix(cleanKey, ".."+string(os.PathSeparator)) || cleanKey == ".." {
		return "", fmt.Errorf("invalid storage key")
	}

	path := filepath.Join(s.rootDir, cleanKey)

	if !strings.HasPrefix(path, s.rootDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid storage key")
	}

	return path, nil
}

func copyWithContext(ctx context.Context, destination io.Writer, source io.Reader) (int64, error) {
	buffer := make([]byte, 32*1024)
	var written int64

	for {

		select {
		case <-ctx.Done():
			return written, fmt.Errorf("write object: %w", ctx.Err())
		default:
		}

		read, readErr := source.Read(buffer)

		if read > 0 {
			count, writeErr := destination.Write(buffer[:read])
			written += int64(count)

			if writeErr != nil {
				return written, fmt.Errorf("write object: %w", writeErr)
			}

			if count != read {
				return written, io.ErrShortWrite
			}
		}

		if readErr == io.EOF {
			return written, nil
		}

		if readErr != nil {
			return written, fmt.Errorf("read object: %w", readErr)
		}
	}
}
