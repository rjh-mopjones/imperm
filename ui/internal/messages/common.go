package messages

import "time"

// Common message types shared across packages

// ErrMsg represents an error message
type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string {
	return e.Err.Error()
}

// TickMsg represents a timer tick
type TickMsg time.Time

// ClearStatusMsg signals to clear the status message
type ClearStatusMsg struct{}
