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
	log            Logger
	acl            *ACL
	workers        []pluginWorkerStruct
}

func newChannelWorker(channel string, bot *Kabukibot) *channelWorker {
	return &channelWorker{
		channel:        channel,
		inputChannel:   make(chan twitch.IncomingMessage, 10),
		leaveSignal:    make(chan struct{}),
		shutdownSignal: make(chan struct{}),
		alive:          make(chan struct{}),
		log:            bot.Logger(),
		acl:            NewACL(channel, bot.OpUsername(), bot.Logger(), bot.Database()),
		workers:        make([]pluginWorkerStruct, 0),
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
	// remember, defers are executed in reverse order
	defer close(self.inputChannel)
	defer close(self.alive)

	// initialize ACL
	self.acl.loadData()

	// initialize plugin workers

	// endless worker loop
	for {
		select {
		case newMsg := <-self.inputChannel:
			// we just left the channel
			_, okay := newMsg.(twitch.PartMessage)
			if okay {
				fmt.Printf("[%s] IS LEAVING THE BUILDING!\n", self.channel)
				return
			}

			// determine the plugins to hand this message to
			for _, worker := range self.workers {
				if !worker.Enabled {
					continue
				}

				switch msg := newMsg.(type) {
				case twitch.TextMessage:
					asserted, okay := worker.Worker.(textMessageWorker)
					if okay {
						asserted.HandleTextMessage(&msg)
					}

				case twitch.RoomStateMessage:
					asserted, okay := worker.Worker.(roomStateMessageWorker)
					if okay {
						asserted.HandleRoomStateMessage(&msg)
					}

				case twitch.ClearChatMessage:
					asserted, okay := worker.Worker.(clearChatMessageWorker)
					if okay {
						asserted.HandleClearChatMessage(&msg)
					}

				case twitch.SubscriberNotificationMessage:
					asserted, okay := worker.Worker.(subNotificationMessageWorker)
					if okay {
						asserted.HandleSubscriberNotificationMessage(&msg)
					}
				}
			}

		case <-self.leaveSignal:
			// for worker in workers {
			// 	worker.leave()
			// }

			return

		case <-self.shutdownSignal:
			// for worker in workers {
			// 	worker.shutdown()
			// }

			return
		}
	}
}
