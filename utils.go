package main

import "sync"

type SynchronizedIntSet struct {
	smu sync.Mutex
	set map[int]bool
}

func (s *SynchronizedIntSet) addElement(elem int) bool {

	// lock `visited` set to check its value
	s.smu.Lock()

	// release lock at the end of function call
	defer s.smu.Unlock()

	// check if not is already visited
	if _, ok := s.set[elem]; ok {
		// element is already in set -> return false
		return false
	}

	// mark node as 'visited' to avoid running processing multiple times
	s.set[elem] = true

	// element was successfully added to set -> return true
	return true
}
