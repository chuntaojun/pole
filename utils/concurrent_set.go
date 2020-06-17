package utils

import "sync"

type void struct{}

var member void

type SyncSet struct {
	container sync.Map
}

func NewSyncSet() SyncSet {
	return SyncSet{
		container:sync.Map{},
	}
}

func (s *SyncSet) Range(f func(value interface{})) {
	s.container.Range(func(key, value interface{}) bool {
		f(key)
		return true
	})
}

func (s *SyncSet) Add(value interface{}) {
	s.container.Store(value, member)
}

func (s *SyncSet) Remove(value interface{}) {
	s.container.Delete(value)
}

