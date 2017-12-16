package main

import (
	"fmt"
	"sync"
)

type SectionState struct {
	state string     `json:"state"`
	mux   sync.Mutex `json:"-"`
}

func newSectionState() SectionState {
	return SectionState{
		state: "",
	}
}

func (s *SectionState) getState() string {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.state
}

func (s *SectionState) trySetState(state string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.state == "" {
		s.state = state
		return nil
	}

	return fmt.Errorf("Already has state: \"%S\"", s.state)
}

func (s *SectionState) deleteState() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state = ""
}

type Sections struct {
	t   map[string]SectionState `json:"table"`
	mux sync.Mutex              `json:"-"`
}

func LoadSections(fileName string) (Sections, error) {
	bytes, err = ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error: Can't load from %s\n    %v", fileName, err)
	}

	var sections Sections
	if err := json.Unmarshal(bytes, &sections); err != nil {
		return nil, fmt.Errorf("Error: Can't parse JSON\n    %v", err)
	}

	return sections, nil
}

func (ss *Sections) Save(fileName string) error {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	bytes, err := json.Marshal(ss)
	if err != nil {
		return fmt.Errorf("Error: Can't encode to JSON\n    %v", err)
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error: Can't open file %s\n   %v", fileName, err)
	}
	defer file.Close()

	if err := file.Write(bytes); err != nil {
		return fmt.Errorf("Error: Can't write to file\n   %v", err)
	}

	return nil
}
