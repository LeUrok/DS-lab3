package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrTimeBasedCircuitOpen = errors.New("circuitbreaker is open")

type WindowConfig struct {
	WindowSize   int
	MaxFailures  float64
	MinRequests  int
	ResetTimeout time.Duration
}

type oneSecond struct {
	timestamp time.Time
	requests  int
	failures  int
}

type TimestampBreaker struct {
	mu       sync.Mutex
	cfg      WindowConfig
	state    State
	segments []oneSecond
	openedAt time.Time
}

func NewTimestampBreaker(cfg WindowConfig) *TimestampBreaker {
	segments := make([]oneSecond, cfg.WindowSize)
	now := time.Now()

	for i := 0; i < len(segments); i++ {
		segments[i] = oneSecond{timestamp: now.Add(-time.Duration(cfg.WindowSize) * time.Second)} 
	}

	return &TimestampBreaker{
		cfg:      cfg,
		state:    StateClosed,
		segments: segments,
	}
}

func (b *TimestampBreaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *TimestampBreaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	switch b.state {
	case StateClosed:
		return nil
	case StateOpen:
		if now.Sub(b.openedAt) >= b.cfg.ResetTimeout {
			b.state = StateHalfOpen
			return nil
		}
		return ErrTimeBasedCircuitOpen
	case StateHalfOpen:
		return ErrTimeBasedCircuitOpen
	}
	return nil
}

func (b *TimestampBreaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.record(time.Now(), false)

	if b.state == StateHalfOpen {
		b.state = StateClosed
	}
}

func (b *TimestampBreaker) record(timeNow time.Time, failed bool) {
	ind := int(timeNow.Unix()) % b.cfg.WindowSize 

	if timeNow.Sub(b.segments[ind].timestamp) > time.Duration(b.cfg.WindowSize)*time.Second {
		b.segments[ind] = oneSecond{timestamp: timeNow}
	}

	b.segments[ind].requests++
	if failed {
		b.segments[ind].failures++
	}

}

func (b *TimestampBreaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.record(time.Now(), true)

	if b.state == StateHalfOpen {
		b.state = StateOpen
		b.openedAt = time.Now()
		return
	}

	requests, failures := b.stats()
	if requests >= b.cfg.MinRequests &&
		float64(failures)/float64(requests) >= b.cfg.MaxFailures {
		b.state = StateOpen
		b.openedAt = time.Now()
		
	}
}

func (b *TimestampBreaker) stats() (requests, failures int) {
	now := time.Now()
	window := time.Duration(b.cfg.WindowSize) * time.Second

	for i := 0; i < len(b.segments); i++ {
		if now.Sub(b.segments[i].timestamp) <= window {
			requests += b.segments[i].requests
			failures += b.segments[i].failures
		}
	}
	return
}
