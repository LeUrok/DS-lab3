package circuitbreaker

import (
	"testing"
	"time"
)

func TestBreaker_StartsClosed(t *testing.T) {
	b := New(Config{MaxFailures: 3, ResetTimeout: time.Second})
	if b.State() != StateClosed {
		t.Fatalf("expected CLOSED, got %s", b.State())
	}
}

func TestBreaker_OpensAfterMaxFailures(t *testing.T) {
	b := New(Config{MaxFailures: 3, ResetTimeout: time.Second})

	for i := 0; i < 3; i++ {
		if err := b.Allow(); err != nil {
			t.Fatalf("expected Allow to pass at i=%d, got %v", i, err)
		}
		b.Failure()
	}

	if b.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", b.State())
	}

	if err := b.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreaker_SuccessResetsCounter(t *testing.T) {
	b := New(Config{MaxFailures: 3, ResetTimeout: time.Second})

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	b.Allow()
	b.Success() 

	b.Allow()
	b.Failure()

	if b.State() != StateClosed {
		t.Fatalf("expected CLOSED, got %s", b.State())
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	b := New(Config{MaxFailures: 2, ResetTimeout: 50 * time.Millisecond})

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	if b.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", b.State())
	}

	time.Sleep(60 * time.Millisecond)

	if err := b.Allow(); err != nil {
		t.Fatalf("expected Allow to pass in HALF-OPEN, got %v", err)
	}
	if b.State() != StateHalfOpen {
		t.Fatalf("expected HALF-OPEN, got %s", b.State())
	}
}

func TestBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	b := New(Config{MaxFailures: 2, ResetTimeout: 50 * time.Millisecond})

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	time.Sleep(60 * time.Millisecond)

	b.Allow()
	b.Success()

	if b.State() != StateClosed {
		t.Fatalf("expected CLOSED, got %s", b.State())
	}
}

func TestBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	b := New(Config{MaxFailures: 2, ResetTimeout: 50 * time.Millisecond})

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	time.Sleep(60 * time.Millisecond)

	b.Allow()
	b.Failure() 

	if b.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", b.State())
	}
}

func TestBreaker_HalfOpenOnlyOneRequest(t *testing.T) {
	b := New(Config{MaxFailures: 2, ResetTimeout: 50 * time.Millisecond})

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	time.Sleep(60 * time.Millisecond)

	if err := b.Allow(); err != nil {
		t.Fatalf("first Allow should pass, got %v", err)
	}
	if err := b.Allow(); err != ErrCircuitOpen {
		t.Fatalf("second Allow should be rejected, got %v", err)
	}
}