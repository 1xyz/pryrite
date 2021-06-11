package queue

// Queue - An interface to queues
type Queue interface {
	// Enqueue an element to the queue
	Enqueue(item interface{})

	// Dequeue an element from the queue
	Dequeue() (interface{}, bool)

	// Len returns length of the underlying queue
	Len() int

	// WaitForItem waut until an item in the underlying queue is ready
	// Once an item is available. Dequeue this item and return it
	WaitForItem() interface{}
}
