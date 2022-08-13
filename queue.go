package gux

import (
	"sync"
)

type (
	QueueListener struct {
		IsConnected bool
		Signal      chan int
	}
	Queue struct {
		queue     []interface{}
		listeners map[string]*QueueListener
		limit     int
		sync.RWMutex
	}
)

// NewQueue instantiates a new Queue struct
func NewQueue(queueLimit int) *Queue {
	return &Queue{
		queue:     []interface{}{},
		listeners: make(map[string]*QueueListener),
		limit:     queueLimit,
	}
}

// Pop returns the first item in the queue and deletes it from the queue
func (q *Queue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()
	var item interface{}
	if len(q.queue) > 0 {
		item = q.queue[0]
		q.queue = q.queue[1:]
	} else {
		q.queue = []interface{}{}
	}
	return item
}

// Push adds a new item to the back of the queue
func (q *Queue) Push(v interface{}) {
	q.Lock()
	defer q.Unlock()
	if len(q.queue) == q.limit {
		q.queue = q.queue[1:]
	}
	q.queue = append(q.queue, v)
	newListenerMap := make(map[string]*QueueListener)
	for n, l := range q.listeners {
		if !l.IsConnected {
			close(l.Signal)
			continue
		}
		l.Signal <- len(q.queue) + 1 // + 1 because then the subscriber can know when the channel is closed (if they receive 0)
		newListenerMap[n] = l
	}
	q.listeners = newListenerMap
}

// Subscribe returns a channel which will have signals sent when a new item is pushed as well as an unsub function
func (q *Queue) Subscribe(name string) (chan int, func(), error) {
	q.Lock()
	defer q.Unlock()
	if _, ok := q.listeners[name]; ok {
		return nil, nil, ErrAlreadySubscribed
	}
	q.listeners[name] = &QueueListener{IsConnected: true, Signal: make(chan int, 2)}
	unsub := func() {
		q.Lock()
		defer q.Unlock()
		q.listeners[name].IsConnected = false
	}
	return q.listeners[name].Signal, unsub, nil
}
