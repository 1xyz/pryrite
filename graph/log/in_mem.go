package log

import (
	"fmt"
	"sync"
)

type inMemLog struct {
	list []*ResultLogEntry
	lock sync.RWMutex
}

func (l *inMemLog) Len() (int, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return len(l.list), nil
}

func (l *inMemLog) Append(entry *ResultLogEntry) error {
	l.lock.Lock()
	l.list = append(l.list, entry)
	l.lock.Unlock()
	return nil
}

func (l *inMemLog) Each(cb func(int, *ResultLogEntry) bool) error {
	l.lock.RLock()
	defer l.lock.RUnlock()
	for i, entry := range l.list {
		if !cb(i, entry) {
			return nil
		}
	}
	return nil
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
func (l *inMemLog) Find(id string) (*ResultLogEntry, error) {
	var found *ResultLogEntry
	err := l.Each(func(_ int, entry *ResultLogEntry) bool {
		if entry.ID == id {
			found = entry
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, ErrResultLogEntryNotFound
	}
	return found, nil
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
