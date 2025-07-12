package client

import (
	"sync"
)

type RateLimiter struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{}
	rl.cond = sync.NewCond(&rl.mu)
	return rl
}

func (rl *RateLimiter) Pause() {
	rl.mu.Lock()
	rl.paused = true
	rl.mu.Unlock()
}

func (rl *RateLimiter) Resume() {
	rl.mu.Lock()
	rl.paused = false
	rl.cond.Broadcast()
	rl.mu.Unlock()
}

func (rl *RateLimiter) WaitIfPaused() {
	rl.mu.Lock()
	for rl.paused {
		rl.cond.Wait()
	}
	rl.mu.Unlock()
}
