package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewQueue_MatchesInterface(t *testing.T) {
	q := NewConcurrentQueue()
	assert.NotNil(t, q)

	mq, ok := q.(*ConcurrentQueue)
	assert.True(t, ok)
	assert.NotNil(t, mq.q)
}

func TestConcurrentQueue_Enqueue_ItemsAreAddedToList(t *testing.T) {
	q := NewConcurrentQueue()
	mq, _ := q.(*ConcurrentQueue)

	q.Enqueue("Garbanzo Beans")
	q.Enqueue("Falafal")
	assert.Equal(t, 2, mq.q.Len())
}

func TestConcurrentQueue_Dequeue_ItemsAreDeQueuedInOrder(t *testing.T) {
	q := NewConcurrentQueue()
	entries := []string{"5", "4", "3", "2", "1"}
	expectedEntries := make([]string, 0, 5)
	for _, e := range entries {
		q.Enqueue(e)
		expectedEntries = append(expectedEntries, e)
	}

	for _, expectedEntry := range expectedEntries {
		actualEntry, ok := q.Dequeue()
		assert.True(t, ok)
		assert.Equal(t, expectedEntry, actualEntry)
	}
}

func TestConcurrentQueue_Dequeue_EmptyQReturnsNotOk(t *testing.T) {
	q := NewConcurrentQueue()
	_, ok := q.Dequeue()
	assert.False(t, ok)
}

func TestConcurrentQueue_WaitForItem_DequeuesItemsInOrder(t *testing.T) {
	q := NewConcurrentQueue()
	entries := []string{"5", "4", "3", "2", "1"}
	expectedEntries := make([]string, 0, 5)
	for _, e := range entries {
		q.Enqueue(e)
		expectedEntries = append(expectedEntries, e)
	}

	for i := 0; i < len(expectedEntries); i++ {
		actualEntry := q.WaitForItem()
		expectedEntry := expectedEntries[i]
		assert.Equal(t, expectedEntry, actualEntry)
	}
}

func TestConcurrentQueue_WaitForItem_DeQueuesItemsAcrossGoRoutines(t *testing.T) {
	q := NewConcurrentQueue()
	entries := []string{"5", "4", "3", "2", "1"}

	go func(entries []string) {
		for _, e := range entries {
			q.Enqueue(e)
		}
	}(entries)

	for i := 0; i < len(entries); i++ {
		actualEntry := q.WaitForItem()
		expectedEntry := entries[i]
		assert.Equal(t, expectedEntry, actualEntry)
	}
}
