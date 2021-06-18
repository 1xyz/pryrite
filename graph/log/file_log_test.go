package log

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestFsLogIndex_New(t *testing.T) {
	_, err := NewResultLogIndex(IndexFileSystem)
	assert.Nil(t, err)
}

func TestFsLogIndex_Append(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	e := newTestLogEntry()
	err := index.Append(e)
	assert.Nil(t, err)
}

func TestFsLogIndex_Get_NotFound(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	_, err := index.Get("abcde")
	assert.Equal(t, ErrResultLogNotFound, err)
}

func TestFsLogIndex_Get(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	entry := newTestLogEntry()
	if err := index.Append(entry); err != nil {
		t.FailNow()
	}

	log, err := index.Get(entry.NodeID)
	assert.Nil(t, err)
	assert.NotNil(t, log)
}

func TestFsLog_Find(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	entries := appendNTestLogEntries(t, 10, index)
	log, err := index.Get(entries[0].NodeID)
	if err != nil {
		t.FailNow()
	}

	actual, err := log.Find(entries[0].ID)
	assert.Nil(t, err)
	assert.NotNil(t, actual)

	assert.Equal(t, entries[0].ID, actual.ID)
	assert.Equal(t, entries[0].NodeID, actual.NodeID)
	assert.NotEqual(t, entries[1].ID, actual.ID)
}

func TestFsLog_Find_NotFound(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	entries := appendNTestLogEntries(t, 10, index)
	log, err := index.Get(entries[0].NodeID)
	if err != nil {
		t.FailNow()
	}

	actual, err := log.Find("hello")
	assert.Equal(t, ErrResultLogEntryNotFound, err)
	assert.Nil(t, actual)
}

func TestFsLog_Len(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	entries := appendNTestLogEntries(t, 10, index)
	log, err := index.Get(entries[0].NodeID)
	if err != nil {
		t.FailNow()
	}

	n, err := log.Len()
	assert.Nil(t, err)
	assert.Equal(t, n, 10)
}

func TestFsLog_Each(t *testing.T) {
	index := newFsIndex(t)
	defer removeDir(t, index)

	entries := appendNTestLogEntries(t, 10, index)
	log, err := index.Get(entries[0].NodeID)
	if err != nil {
		t.FailNow()
	}

	idx := 0
	err = log.Each(func(i int, entry *ResultLogEntry) bool {
		assert.Equal(t, idx, i)
		assert.Equal(t, entries[idx].ID, entry.ID)
		idx++
		return true
	})
	assert.Nil(t, err)
}

func appendNTestLogEntries(t *testing.T, n int, index *fsLogIndex) []*ResultLogEntry {
	entries := make([]*ResultLogEntry, 0)
	for i := 0; i < n; i++ {
		entry := newTestLogEntry()
		entry.ID = fmt.Sprintf("log-%d", i)
		if err := index.Append(entry); err != nil {
			t.FailNow()
		}
		entries = append(entries, entry)
	}
	return entries
}

func newTestLogEntry() *ResultLogEntry {
	return &ResultLogEntry{
		ID:          "log1",
		ExecutionID: "ex1",
		NodeID:      "node1",
		BlockID:     "block1",
		RequestID:   "req1",
		ExitStatus:  "",
		Stdout:      "",
		Stderr:      "",
		ExecutedAt:  nil,
		ExecutedBy:  "",
		State:       "",
		Err:         nil,
		Content:     "",
	}
}

func newFsIndex(t *testing.T) *fsLogIndex {
	dir, err := ioutil.TempDir("", "fslog")
	if err != nil {
		assert.Failf(t, "ioutil.TempDir", "err = %v", err)
	}
	t.Logf("newfsIndex: dir = %s", dir)

	logIndex, err := newFSLogIndex(dir)
	if err != nil {
		assert.Failf(t, "NewResultLogIndex", "unexpected err is not nil %v", err)
	}
	return logIndex
}

func removeDir(t *testing.T, index *fsLogIndex) {
	if err := os.RemoveAll(index.dir); err != nil {
		assert.Failf(t, "removeDir: os.RemoveAll(%s) err = %v", index.dir, err)
	}
}
