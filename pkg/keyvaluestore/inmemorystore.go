package keyvaluestore

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var _ KV = (*MemoryStore)(nil)

type MemoryStore struct {
	data map[string]interface{}
	mux  *sync.RWMutex
}

func InMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]interface{}), mux: &sync.RWMutex{}}
}

func (m *MemoryStore) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	outPtr := reflect.ValueOf(out)
	if outPtr.Kind() != reflect.Ptr {
		return false, fmt.Errorf("out must be a pointer, got non-pointer type %s", outPtr.Kind())
	}

	data, exists := m.data[key]
	if !exists {
		return false, nil
	}
	var copied reflect.Value
	if reflect.ValueOf(data).Kind() == reflect.Ptr {
		copied = reflect.ValueOf(data).Elem()
	} else {
		copied = reflect.ValueOf(data)
	}

	if copied.Kind() != outPtr.Elem().Kind() {
		return false, fmt.Errorf("type mismatch: cannot assign %s into %s", copied.Kind(), outPtr.Elem().Kind())
	}

	outPtr.Elem().Set(copied)
	return true, nil
}

func (m *MemoryStore) Put(ctx context.Context, key string, data interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.data[key] = data
	return nil
}
