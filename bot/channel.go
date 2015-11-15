package bot

import (
	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

// this is what's visible to plugins
type Channel interface {
	Name() string
	Alive() <-chan struct{}
	Plugins() []Plugin
	ACL() *ACL
	EnablePlugin(string) bool
	DisablePlugin(string) bool
}

type channelWorker struct {
	channel        string
	inputChannel   chan twitch.IncomingMessage // should be buffered to sth. like 100msg
	leaveSignal    chan struct{}               // to be sent (= closed) when we LEAVE the channel on purpose
	shutdownSignal chan struct{}               // to be sent when we just shutdown the bot
	alive          chan struct{}               // is sent by the worker when the goroutine is ending
	database       *sqlx.DB
	log            Logger
	acl            *ACL
	workers        []pluginWorkerStruct
	sender         Sender
}

type pluginRow struct {
	Plugin string `db:"plugin"`
}

func newChannelWorker(channel string, bot *Kabukibot) *channelWorker {
	workers := make([]pluginWorkerStruct, 0)

	cw := &channelWorker{
		channel:        channel,
		inputChannel:   make(chan twitch.IncomingMessage, 10),
		leaveSignal:    make(chan struct{}),
		shutdownSignal: make(chan struct{}),
		alive:          make(chan struct{}),
		database:       bot.Database(),
		log:            bot.Logger(),
		acl:            NewACL(channel, bot.OpUsername(), bot.Logger(), bot.Database()),
		workers:        nil,
		sender:         newSenderStruct(bot.twitch, channel),
	}

	// find out what plugins have been enabled for the channel
	list := make([]pluginRow, 0)
	bot.Database().Select(&list, "SELECT plugin FROM plugin WHERE channel = ?", channel)

	for _, plugin := range bot.Plugins() {
		name := plugin.Name()
		enabled := (name == "")

		for _, l := range list {
			if l.Plugin == name {
				enabled = true
				break
			}
		}

		workers = append(workers, pluginWorkerStruct{
			Plugin:  plugin,
			Worker:  plugin.CreateWorker(cw),
			Enabled: enabled,
		})
	}

	cw.workers = workers

	return cw
}

func (self *channelWorker) Name() string {
	return self.channel
}

func (self *channelWorker) Plugins() []Plugin {
	result := make([]Plugin, 0)

	for _, worker := range self.workers {
		if worker.Enabled {
			result = append(result, worker.Plugin)
		}
	}

	return result
}

func (self *channelWorker) ACL() *ACL {
	return self.acl
}

func (self *channelWorker) EnablePlugin(name string) bool {
	worker := self.findWorker(name)

	if worker == nil || worker.Enabled {
		return false
	}

	worker.Enabled = true
	worker.Worker.Enable()

	self.database.Exec("INSERT INTO plugin (channel, plugin) VALUES (?, ?)", self.channel, name)

	return true
}

func (self *channelWorker) DisablePlugin(name string) bool {
	worker := self.findWorker(name)

	if worker == nil || !worker.Enabled {
		return false
	}

	worker.Enabled = false
	worker.Worker.Disable()

	self.database.Exec("DELETE FROM plugin WHERE channel = ? AND plugin = ?", self.channel, name)

	return true
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

	// enable workers
	for _, worker := range self.workers {
		if worker.Enabled {
			worker.Worker.Enable()
		}
	}

	// endless worker loop
	for {
		select {
		case newMsg := <-self.inputChannel:
			// we just left the channel
			_, okay := newMsg.(twitch.PartMessage)
			if okay {
				self.partWorkers()
				return
			}

			// determine the plugins to hand this message to
			for _, worker := range self.workers {
				if !worker.Enabled {
					continue
				}

				switch msg := newMsg.(type) {
				case TextMessage:
					asserted, okay := worker.Worker.(textMessageWorker)
					if okay {
						asserted.HandleTextMessage(&msg, self.sender)
					}

				case twitch.RoomStateMessage:
					asserted, okay := worker.Worker.(roomStateMessageWorker)
					if okay {
						asserted.HandleRoomStateMessage(&msg, self.sender)
					}

				case twitch.ClearChatMessage:
					asserted, okay := worker.Worker.(clearChatMessageWorker)
					if okay {
						asserted.HandleClearChatMessage(&msg, self.sender)
					}

				case twitch.SubscriberNotificationMessage:
					asserted, okay := worker.Worker.(subNotificationMessageWorker)
					if okay {
						asserted.HandleSubscriberNotificationMessage(&msg, self.sender)
					}
				}
			}

		case <-self.leaveSignal:
			self.partWorkers()
			return

		case <-self.shutdownSignal:
			self.shutdownWorkers()
			return
		}
	}
}

func (self *channelWorker) partWorkers() {
	for _, worker := range self.workers {
		worker.Worker.Part()
	}
}

func (self *channelWorker) shutdownWorkers() {
	for _, worker := range self.workers {
		worker.Worker.Shutdown()
	}
}

func (self *channelWorker) findWorker(pluginName string) *pluginWorkerStruct {
	for idx, ws := range self.workers {
		if ws.Plugin.Name() == pluginName {
			return &self.workers[idx]
		}
	}

	return nil
}
