package utils

import "sync"

type SafeCounter struct {
	c  int64
	mu sync.Mutex
}

func MakeSafeCounter() *SafeCounter {
	return &SafeCounter{c: 0}
}

func (c *SafeCounter) Add(delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c += delta
}

func (c *SafeCounter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c
}
