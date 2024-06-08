package session

import (
	"fmt"
	"goskydarks/specs"
)

// Session struct implements the session service, used for overall session control
// such as start time or resuming from saved state
type Session struct {
}

func NewSession() (*Session, error) {
	return &Session{}, nil
}

func (s *Session) PrepareForCapture(config specs.CaptureSpecs) error {
	fmt.Printf("session.PrepareForCapture %#v\n", config)
	return nil
}
