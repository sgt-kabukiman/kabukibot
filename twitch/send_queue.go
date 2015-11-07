package twitch

import (
	"time"
)

type QueueItem func()

type SendQueue interface {
	Push(QueueItem)
	Worker()
}

func NewSendQueue(delay time.Duration) SendQueue {
	return &sendQueue{make([]QueueItem, 0), delay}
}

type messageQueue []QueueItem

func (q *messageQueue) push(item QueueItem) {
	*q = append(*q, item)
}

func (q *messageQueue) pop() QueueItem {
	head := (*q)[0]
	*q = (*q)[1:]

	return head
}

func (q *messageQueue) len() int {
	return len(*q)
}

type sendQueue struct {
	queue messageQueue
	delay time.Duration
}

func (s *sendQueue) Push(item QueueItem) {
	s.queue.push(item)
}

func (s *sendQueue) Worker() {
	for {
		if s.queue.len() > 0 {
			s.queue.pop()()
		}

		<-time.After(s.delay)
	}
}
