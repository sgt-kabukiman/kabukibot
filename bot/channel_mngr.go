package bot

import "log"
import "strings"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type channelMap map[string]*twitch.Channel

type ChannelManager struct {
	db       *DatabaseStruct
	channels channelMap
}

func NewChannelManager(db *DatabaseStruct) *ChannelManager {
	return &ChannelManager{db, make(channelMap)}
}

func (self *ChannelManager) Channels() *channelMap {
	return &self.channels
}

func (self *ChannelManager) Channel(name string) (c *twitch.Channel, ok bool) {
	c, ok = self.channels[strings.TrimLeft(name, "#")]
	return
}

func (self *ChannelManager) loadChannels() {
	rows, err := self.db.Query("SELECT * FROM channel")
	if err != nil {
		log.Fatal("Could not query the channels: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}

		cn := twitch.NewChannel(name)
		self.channels[name] = cn
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func (self *ChannelManager) addChannel(channel string) *twitch.Channel {
	channel = twitch.NormaliseChannelName(channel)

	cn, ok := self.channels[channel]
	if ok {
		return cn
	}

	_, err := self.db.Query("INSERT INTO channel VALUES (?)", channel)
	if err != nil {
		log.Fatal("Could not add channel #" + channel + " to the database: " + err.Error())
	}

	cn = twitch.NewChannel(channel)
	self.channels[channel] = cn

	return cn
}

func (self *ChannelManager) removeChannel(channel string) {
	channel = twitch.NormaliseChannelName(channel)

	_, ok := self.channels[channel]
	if !ok {
		return
	}

	_, err := self.db.Query("DELETE FROM channel WHERE name = ?", channel)
	if err != nil {
		log.Fatal("Could not remove channel #" + channel + " from the database: " + err.Error())
	}

	delete(self.channels, channel)
}
