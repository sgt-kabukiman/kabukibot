package bot

import (
	"time"
)

type queueItem func()
type messageQueue []queueItem

func (q *messageQueue) push(item queueItem) {
	*q = append(*q, item)
}

func (q *messageQueue) pop() queueItem {
	head := (*q)[0]
	*q    = (*q)[1:]

	return head
}

func (q *messageQueue) len() int {
	return len(*q)
}

type SendQueue struct {
	queue  messageQueue
	delay  time.Duration
}

func NewSendQueue(delay time.Duration) *SendQueue {
	return &SendQueue{make([]queueItem, 0), delay}
}

func (s *SendQueue) Push(item queueItem) {
	s.queue.push(item)
}

func (s* SendQueue) Worker() {
	for {
		if (s.queue.len() > 0) {
			s.queue.pop()()
		}

		<-time.After(s.delay)
	}
}
