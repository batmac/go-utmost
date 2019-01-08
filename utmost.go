// Package utmost lets you limit the number of simultaneous goroutines with some observability.
package utmost

import (
	"sync"
)

const (
	// DefaultUtmost is used if you ask a negative limit
	DefaultUtmost = 100
)

// Telemetry contains vars used for observability only
// structcheck doesn't handle embedded structs yet.
//nolint:structcheck
type telemetry struct {
	dispensed int
	inUse     int
	maxInUse  int
	m         sync.RWMutex
}

// TicketsMachine is our base type
type TicketsMachine struct {
	telemetry
	dispenser chan int
	limit     int
	wg        sync.WaitGroup
}

// Wait blocks until the WaitGroup counter is zero.
func (t *TicketsMachine) Wait() {
	t.wg.Wait()
}

// New returns a initialized *TicketsMachine
func New(limit int) *TicketsMachine {
	// "int" is enough ((limitint=int(^uint(0) >> 1)))
	if limit < 1 {
		limit = DefaultUtmost
	}
	t := &TicketsMachine{
		dispenser: make(chan int, limit),
		limit:     limit,
	}
	for i := 0; i < limit; i++ {
		t.dispenser <- i
	}
	return t
}

// Go executes your function as a goroutine, and blocks awaiting a free ticket if needed
func (t *TicketsMachine) Go(goroutine func()) {
	ticket := <-t.dispenser
	t.m.Lock()
	t.inUse++
	t.dispensed++
	if t.maxInUse < t.inUse {
		t.maxInUse = t.inUse
	}
	t.m.Unlock()
	t.wg.Add(1)
	go func() {
		defer func() {
			// inUse must be decremented before giving our ticket back to avoid
			// this racy example:
			// (routine A) t.dispenser <- ticket
			// (routine B) ticket := <-t.dispenser
			// (routine B) Lock ... inUse++ ... Unlock
			// < here, InUse() (and so MaxInUse() too) = limit+1 ! >
			// (routine A) Lock ... inUse--
			t.m.Lock()
			t.inUse--
			t.m.Unlock()
			t.dispenser <- ticket
			t.wg.Done()
		}()
		goroutine()
	}()
}

// Limit returns the limit of simultaneous dispensed tickets
func (t *TicketsMachine) Limit() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.limit
}

// Dispensed returns the total number of dispensed tickets
func (t *TicketsMachine) Dispensed() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.dispensed
}

// MaxInUse returns the max simultaneous used tickets
func (t *TicketsMachine) MaxInUse() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.maxInUse
}

// InUse returns the currently used tickets
func (t *TicketsMachine) InUse() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.inUse
}
