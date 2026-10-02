package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	}
	return "UNKNOWN"
}

var ErrCircuitOpen = errors.New("circuitbreaker is open")

type Config struct {
	MaxFailures  int
	ResetTimeout time.Duration
}

type Breaker struct {
	mu          sync.Mutex
	cfg         Config
	state       State
	failures    int
	lastFailure time.Time
	openedAt    time.Time
}

func New(cfg Config) *Breaker {
	return &Breaker{
		cfg:   cfg,
		state: StateClosed,
	}
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *Breaker) Allow() error {
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

		return ErrCircuitOpen
	case StateHalfOpen:
		return ErrCircuitOpen
	}

	return nil
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.state = StateClosed
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++
	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		if b.failures >= b.cfg.MaxFailures {
			b.state = StateOpen
			b.openedAt = time.Now()
		}
	case StateHalfOpen:
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}
