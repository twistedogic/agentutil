package lsp

import "sync"

// VersionedMap is a thread-safe map that tracks a version number on every write.
// It enables readers to detect changes without locking.
type VersionedMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
	ver  uint64
}

// NewVersionedMap creates a new VersionedMap.
func NewVersionedMap[K comparable, V any]() *VersionedMap[K, V] {
	return &VersionedMap[K, V]{data: make(map[K]V)}
}

// Set sets the value for key and increments the version.
func (m *VersionedMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.ver++
}

// Get returns the value for key.
func (m *VersionedMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

// Delete deletes the key.
func (m *VersionedMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	m.ver++
}

// Version returns the current version number.
func (m *VersionedMap[K, V]) Version() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ver
}

// Copy returns a snapshot of all key-value pairs.
func (m *VersionedMap[K, V]) Copy() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[K]V, len(m.data))
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// Seq2 returns all key-value pairs as a map.
func (m *VersionedMap[K, V]) Seq2() map[K]V {
	return m.Copy()
}