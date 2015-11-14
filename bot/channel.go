package bot

import (
	"fmt"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type channelWorker struct {
	channel        string
	inputChannel   chan twitch.IncomingMessage // should be buffered to sth. like 100msg
	leaveSignal    chan struct{}               // to be sent (= closed) when we LEAVE the channel on purpose
	shutdownSignal chan struct{}               // to be sent when we just shutdown the bot
	alive          chan struct{}               // is sent by the worker when the goroutine is ending
}

func newChannelWorker(channel string) *channelWorker {
	return &channelWorker{
		channel:        channel,
		inputChannel:   make(chan twitch.IncomingMessage, 10),
		leaveSignal:    make(chan struct{}),
		shutdownSignal: make(chan struct{}),
		alive:          make(chan struct{}),
	}
}

func (self *channelWorker) Input() chan<- twitch.IncomingMessage {
	return self.inputChannel
}

func (self *channelWorker) Alive() <-chan struct{} {
	return self.alive
}

func (self *channelWorker) Leave() <-chan struct{} {
	close(self.leaveSignal)

	return self.alive
}

func (self *channelWorker) Shutdown() <-chan struct{} {
	close(self.shutdownSignal)

	return self.alive
}

func (self *channelWorker) Work() {
	// endless worker loop
	for {
		select {
		case newMsg := <-self.inputChannel:
			// we just left the channel
			_, okay := newMsg.(twitch.PartMessage)
			if okay {
				fmt.Printf("[%s] IS LEAVING THE BUILDING!\n", self.channel)
				break
			}

			fmt.Printf("[%s] %+v\n", self.channel, newMsg)

		case <-self.leaveSignal:
			// for worker in workers {
			// 	worker.leave()
			// }

			break // out of the endless loop

		case <-self.shutdownSignal:
			// for worker in workers {
			// 	worker.shutdown()
			// }

			break // out of the endless loop
		}
	}

	close(self.alive)
	close(self.inputChannel)
}
