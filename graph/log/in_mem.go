package log

import (
	"fmt"
	"sync"
)

type inMemLog struct {
	list []*ResultLogEntry
	lock sync.RWMutex
}

func (l *inMemLog) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return len(l.list)
}

func (l *inMemLog) Append(entry *ResultLogEntry) {
	l.lock.Lock()
	l.list = append(l.list, entry)
	l.lock.Unlock()
}

func (l *inMemLog) Each(cb func(int, *ResultLogEntry) bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	for i, entry := range l.list {
		if !cb(i, entry) {
			return
		}
	}
}

func (l *inMemLog) EachFromEnd(cb func(int, *ResultLogEntry) bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	for i := len(l.list) - 1; i >= 0; i-- {
		entry := l.list[i]
		if !cb(i, entry) {
			return
		}
	}
}

// Find scans the slice of BlockExecutionResults by ID
func (l *inMemLog) Find(id string) (*ResultLogEntry, bool) {
	var found *ResultLogEntry
	l.Each(func(_ int, entry *ResultLogEntry) bool {
		if entry.ID == id {
			found = entry
			return false
		}
		return true
	})
	return found, found != nil
}

type inMemLogIndex struct {
	sync.Map
}

func newInMemLogIndex() *inMemLogIndex {
	return &inMemLogIndex{}
}

func (i *inMemLogIndex) Append(entry *ResultLogEntry) error {
	blockResults, _ := i.LoadOrStore(entry.NodeID, &inMemLog{})
	blockResults.(*inMemLog).Append(entry)
	return nil
}

func (i *inMemLogIndex) Get(nodeID string) (ResultLog, error) {
	e, ok := i.Load(nodeID)
	if ok {
		return e.(*inMemLog), nil
	}
	return nil, fmt.Errorf("no log found for nodeID=%s", nodeID)
}
