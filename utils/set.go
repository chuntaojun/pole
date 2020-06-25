package utils

import "sync"

type void struct{}

var member void

type Set struct {
	container	map[interface{}]void
}

func NewSet() Set {
	return Set{
		container:make(map[interface{}]void),
	}
}

func (s *Set) Range(f func(value interface{})) {
	for v, _ := range s.container {
		f(v)
	}
}

func (s *Set) Add(value interface{}) {
	s.container[value] = member
}

func (s *Set) Remove(value interface{}) {
	delete(s.container, value)
}

func (s *Set) Size() int {
	return len(s.container)
}

func (s *Set) ToSlice() []interface{} {
	result := make([]interface{}, len(s.container))

	for v, _ := range s.container {
		result = append(result, v)
	}

	return result
}

type SyncSet struct {
	container sync.Map
}

func NewSyncSet() *SyncSet {
	return &SyncSet{
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

