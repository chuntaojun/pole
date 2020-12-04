package common

import "sync"

const (
	PoleContextKey = "pole_context"
)

type PoleContext struct {
	lock     sync.RWMutex
	metadata map[string]interface{}
}
