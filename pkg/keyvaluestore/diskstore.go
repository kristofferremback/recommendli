package keyvaluestore

import (
	"context"
	"fmt"
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

func (c *DiskStore) Get(ctx context.Context, key string, out interface{}) error {
	// File doesn't exist is fine
	if _, err := os.Stat(c.filename(key)); err != nil {
		return nil
	}
	file, err := os.Open(c.filename(key))
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()
	if err := c.serializer.Deserialize(file, &out); err != nil {
		return fmt.Errorf("deserializing data: %w", err)
	}
	return nil
}

func (c *DiskStore) Put(ctx context.Context, key string, data interface{}) error {
	if err := mkdirp(c.dir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	file, err := os.OpenFile(c.filename(key), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()
	if err := c.serializer.Serialize(file, data); err != nil {
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
