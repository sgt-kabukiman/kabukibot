package emote_counter

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type emoteCountMap map[string]int

type worker struct {
	plugin.NilWorker

	channel     string
	acl         *bot.ACL
	db          *sqlx.DB
	stats       emoteCountMap
	queue       chan *bot.TextMessage
	syncing     chan struct{}
	stopSyncing chan struct{}
	mutex       sync.RWMutex
}

type emoteDbStruct struct {
	Emote   string
	Counter int
}

func (self *worker) Enable() {
	list := make([]emoteDbStruct, 0)
	self.db.Select(&list, "SELECT emote, counter FROM emote_counter WHERE channel = ?", self.channel)

	self.stats = make(emoteCountMap)

	for _, item := range list {
		self.stats[item.Emote] = item.Counter
	}

	// if we for some reason are already syncing, stop now
	if self.syncing != nil {
		self.Disable()
	}

	self.syncing = make(chan struct{})
	self.stopSyncing = make(chan struct{})

	go self.worker()
}

func (self *worker) Disable() {
	close(self.stopSyncing)
	<-self.syncing
}

func (self *worker) Permissions() []string {
	return []string{"use_emote_counter"}
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	if msg.IsCommand("top_emotes") {
		self.handleTopEmotesCommand(msg, sender)
		msg.SetProcessed()
	} else if msg.IsCommand("emote_count") {
		self.handleEmoteCountCommand(msg, sender)
		msg.SetProcessed()
	} else if msg.IsCommand("reset_emote_counter") {
		self.handleResetCommand(msg, sender)
		msg.SetProcessed()
	} else {
		self.handleRegularText(msg, sender)
	}
}

func (self *worker) handleTopEmotesCommand(msg *bot.TextMessage, sender bot.Sender) {
	if !self.acl.IsAllowed(msg.User, "use_emote_counter") {
		return
	}

	max := 5
	args := msg.Arguments()

	if len(args) > 0 {
		value, err := strconv.Atoi(args[0])
		if err != nil {
			if value > 20 {
				max = 20
			} else if value < 1 {
				max = 1
			} else {
				max = value
			}
		}
	}

	top := self.topEmotes(max)

	if len(top) == 0 {
		sender.Respond("no emotes have been counted yet.")
		return
	}

	output := make([]string, 0, len(top))

	for idx, emote := range top {
		output[idx] = fmt.Sprintf("%s (%s x)", emote.emote, humanize.FormatInteger("#,###.", emote.count))
	}

	if max == 1 {
		sender.Respond("this channel's top emote is: " + output[0])
	} else {
		sender.Respond(fmt.Sprintf("this channel's top %d emotes are: %s", len(output), bot.HumanJoin(output, ", ")))
	}
}

func (self *worker) handleEmoteCountCommand(msg *bot.TextMessage, sender bot.Sender) {
	if !self.acl.IsAllowed(msg.User, "use_emote_counter") {
		return
	}

	args := msg.Arguments()

	if len(args) == 0 {
		sender.Respond("you did not give any emote.")
		return
	}

	emote := args[0]
	count, _ := self.stats[emote]

	if count == 0 {
		sender.Respond(emote + " has not yet been used or does not even exist.")
	} else if count == 1 {
		sender.Respond(emote + " has been used once.")
	} else {
		sender.Respond(fmt.Sprintf("%s has been used %s times.", emote, humanize.FormatInteger("#,###.", count)))
	}
}

func (self *worker) handleResetCommand(msg *bot.TextMessage, sender bot.Sender) {
	if !self.acl.IsAllowed(msg.User, "use_emote_counter") {
		return
	}

	self.mutex.Lock()
	self.stats = make(emoteCountMap)
	self.mutex.Unlock()

	self.sync()

	sender.Respond("the emote counter has been reset.")
}

func (self *worker) handleRegularText(msg *bot.TextMessage, sender bot.Sender) {
	// anticipating that finding emotes will be complex when taking FFZ into account, we put
	// this into the worker goroutine
	self.queue <- msg
}

func (self *worker) worker() {
	defer close(self.syncing)

	for {
		select {
		case <-time.After(5 * time.Minute):
			self.sync()

		case msg := <-self.queue:
			self.countEmotes(msg)

		case <-self.stopSyncing:
			self.sync()
			return
		}
	}
}

func (self *worker) sync() {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	self.db.Exec("DELETE FROM emote_counter WHERE channel = ?", self.channel)

	for emote, counter := range self.stats {
		self.db.Exec("INSERT INTO emote_counter (channel, emote, counter) VALUES (?, ?, ?)", self.channel, emote, counter)
	}
}

func (self *worker) countEmotes(msg *bot.TextMessage) {
	self.mutex.Lock()

	for _, occurences := range msg.User.Emotes {
		// find the first instance, so we know what emote we have to deal with
		emote := msg.Text[occurences[0].FirstChar : occurences[0].LastChar+1]

		count, _ := self.stats[emote]
		self.stats[emote] = count + len(occurences)
	}

	self.mutex.Unlock()
}

type emoteFlat struct {
	emote string
	count int
}

type emoteSorter []emoteFlat

func (a emoteSorter) Len() int {
	return len(a)
}

func (a emoteSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a emoteSorter) Less(i, j int) bool {
	return a[i].count > a[j].count
}

func (self *worker) topEmotes(max int) []emoteFlat {
	self.mutex.RLock()

	result := make([]emoteFlat, 0, len(self.stats))

	for emote, count := range self.stats {
		result = append(result, emoteFlat{emote, count})
	}

	self.mutex.RUnlock()

	sort.Sort(sort.Reverse(emoteSorter(result)))

	if len(result) > max {
		return result[:max]
	}

	return result
}
