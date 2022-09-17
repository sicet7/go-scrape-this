package utils

import (
	"runtime"
	"sync"
	"time"
)

type MemoryUsage struct {
	stats    runtime.MemStats
	lastRead time.Time
	lock     sync.Mutex
}

func NewMemoryUsage() MemoryUsage {
	return MemoryUsage{}
}

func (m *MemoryUsage) Get() *runtime.MemStats {
	if time.Since(m.lastRead) <= time.Second {
		return &m.stats
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if time.Since(m.lastRead) <= time.Second {
		return &m.stats
	}
	runtime.ReadMemStats(&m.stats)
	m.lastRead = time.Now()
	return &m.stats
}
