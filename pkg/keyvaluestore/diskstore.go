package keyvaluestore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type DiskStore struct {
	dir        string
	serializer Serializer
}

func JSONDiskStore(dir string) *DiskStore {
	return &DiskStore{dir: dir, serializer: &JSONSerializer{}}
}

func (d *DiskStore) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	// File doesn't exist is fine
	if _, err := os.Stat(d.filename(key)); err != nil {
		return false, nil
	}
	file, err := os.Open(d.filename(key))
	if err != nil {
		return false, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()
	// Sometimes the file in Replit is empty
	if err := d.serializer.Deserialize(file, &out); err != nil && errors.Is(err, io.ErrUnexpectedEOF) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("deserializing data: %w", err)
	}
	return true, nil
}

func (d *DiskStore) Put(ctx context.Context, key string, data interface{}) error {
	if err := mkdirp(d.dir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	file, err := os.OpenFile(d.filename(key), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()
	if err := d.serializer.Serialize(file, data); err != nil {
		return fmt.Errorf("serializing data: %w", err)
	}
	return nil
}

func (c *DiskStore) filename(key string) string {
	return filepath.Join(c.dir, key)
}

func mkdirp(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	return nil
}
