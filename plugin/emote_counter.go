package plugin

import "fmt"
import "time"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type emoteCountMap map[string]int
type emoteStatsMap map[string]*emoteCountMap

type EmoteCounterPlugin struct {
	channelPlugin

	bot   *bot.Kabukibot
	db    *bot.DatabaseStruct
	em    bot.EmoteManager
	log   bot.Logger
	stats emoteStatsMap
	foo   chan bool
}

func NewEmoteCounterPlugin() *EmoteCounterPlugin {
	return &EmoteCounterPlugin{newChannelPlugin(), nil, nil, nil, nil, nil, nil}
}

func (self *EmoteCounterPlugin) Key() string {
	return "emote_counter"
}

func (self *EmoteCounterPlugin) Permissions() []string {
	return []string{"use_emote_counter"}
}

func (self *EmoteCounterPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot = bot
	self.db = bot.Database()
	self.em = bot.EmoteManager()
	self.log = bot.Logger()
	self.stats = make(emoteStatsMap)
	self.foo = make(chan bool)

	d.OnCommand(self.onGlobalCommand, nil)

	go self.syncer()
}

func (self *EmoteCounterPlugin) Load(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	chanName := c.Name

	_, exists := self.stats[chanName]
	if exists {
		delete(self.stats, chanName)
	}

	rows, err := self.db.Query("SELECT emote, counter FROM emote_counter WHERE channel = ?", chanName)
	if err != nil {
		self.log.Fatal("Could not query emote counter: %s", err.Error())
	}
	defer rows.Close()

	newMap := make(emoteCountMap)
	rowCount := 0

	for rows.Next() {
		var emote string
		var counter int

		if err := rows.Scan(&emote, &counter); err != nil {
			self.log.Fatal(err.Error())
		}

		newMap[emote] = counter
		rowCount = rowCount + 1
	}

	if err := rows.Err(); err != nil {
		self.log.Fatal(err.Error())
	}

	self.stats[chanName] = &newMap

	self.log.Debug("Loaded %d counted emotes for #%s.", rowCount, chanName)

	self.addChannelListeners(c, listenerList{
		d.OnCommand(self.onChannelCommand, c),
		d.OnTextMessage(self.onText, c),
	})
}

func (self *EmoteCounterPlugin) Unload(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	_, exists := self.stats[c.Name]
	if exists {
		delete(self.stats, c.Name)
	}

	self.removeChannelListeners(c)
}

func (self *EmoteCounterPlugin) onGlobalCommand(cmd bot.Command) {
}

func (self *EmoteCounterPlugin) onChannelCommand(cmd bot.Command) {
}

func (self *EmoteCounterPlugin) onText(msg twitch.TextMessage) {
	before := time.Now()
	emotes := self.em.FindEmotesInMessage(msg)

	fmt.Printf("      > %s\n", time.Now().Sub(before).String())

	if len(emotes) > 0 {
		fmt.Printf("      > found emotes: %v\n", emotes, msg.User().EmoteSet)
	}
}

func (self *EmoteCounterPlugin) syncer() {
	for {
		<-time.After(15 * time.Minute)
		self.syncAll()
	}
}

func (self *EmoteCounterPlugin) syncAll() {
	for channel, _ := range self.stats {
		self.syncChannel(channel)
	}
}

func (self *EmoteCounterPlugin) syncChannel(channel string) {
	self.log.Debug("Syncing emote counter for #%s.", channel)

	_, err := self.db.Exec("DELETE FROM emote_counter WHERE channel = ?", channel)
	if err != nil {
		self.log.Fatal("Could not delete emote counter data from the database: " + err.Error())
	}

	stats, exists := self.stats[channel]
	if exists {
		for emote, counter := range *stats {
			_, err := self.db.Exec("INSERT INTO emote_counter (channel, emote, counter) VALUES (?, ?, ?)", channel, emote, counter)
			if err != nil {
				self.log.Fatal("Could not insert emote counter data: " + err.Error())
			}
		}
	}
}
