// Package gopherboat provides a gopher and a boat, duh
package gopherboat

import "time"

// Signal represents a signal that can be sent to a boat.
type Signal struct{}

// ProcessFunc ...
type ProcessFunc func(chan Signal)

// tippedBoat contains the panic message of tipping the boat,
// tippedBoat panics are the only panics that are recovered in a boat,
// since we can't block out all other panics!
type tippedBoat string

// New creates a new gopher boat, the boat is on a tight budget so
// it can only hold one gopher, if that gopher invites his friends
// the chance of the boat causing the end of the world greatly incresses.
func New(f ProcessFunc) *Boat {
	boat := &Boat{
		Signals: make(chan Signal),
		Panic: make(chan interface{}, 1)
	}

	go func() {
		defer func() {
			e := recover()
			if tipMsg, ok := e.(tippedBoat); ok {
				return
			}

			panic(e)
		}()

		// fuck you rob pike it can't be done

		f(boat.Signals)
	}()

	return boat
}

// Boat is a handle on a running boat
type Boat struct {
	Signals chan Signal
}

// Tip tips over the boat drowning its gophers, a delay can
// be provided to give them time to jump out before the
// boat is tipped over.
func (b *Boat) Tip(delay time.Duration) {
}

// Send Sends a signal to the boat
func (b *Boat) Send() {
}
