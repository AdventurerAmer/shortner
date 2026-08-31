package health

import (
	"sync/atomic"
)

type Checker struct {
	ready atomic.Bool
}

func (c *Checker) IsReady() bool {
	return c.ready.Load()
}

func (c *Checker) IsNotReady() bool {
	return c.ready.Load() == false
}

func (c *Checker) Ready() {
	c.ready.Store(true)
}

func (c *Checker) Shutdown() {
	c.ready.Store(false)
}
