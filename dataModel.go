package main

import (
	"fmt"
	"sync"
)

type SectionState struct {
	state string     `json:"state"`
	mux   sync.Mutex `json:"-"`
}

func newSectionState() {
	return SectionState{
		state: "",
		mux:   make(sync.Mutex),
	}
}

func (s *SectionState) getState() string {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.state
}

func (s *SectionState) trySetState(string state) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state == "" {
		s.state = state
		return nil
	}

	return fmt.Eprint("Already has state")
}

func (s *SectionState) deleteState() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state = ""
}
