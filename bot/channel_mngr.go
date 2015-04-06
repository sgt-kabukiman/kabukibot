package bot

import "log"
import "strings"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type channelMap map[string]*twitch.Channel

type channelManager struct {
	db       *DatabaseStruct
	channels channelMap
}

func NewChannelManager(db *DatabaseStruct) *channelManager {
	return &channelManager{db, make(channelMap)}
}

func (self *channelManager) Channels() *channelMap {
	return &self.channels
}

func (self *channelManager) Channel(name string) (c *twitch.Channel, ok bool) {
	c, ok = self.channels[strings.TrimLeft(name, "#")]
	return
}

func (self *channelManager) loadChannels() {
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

func (self *channelManager) addChannel(channel *twitch.Channel) {
	name := channel.Name

	_, ok := self.channels[name]
	if ok {
		return
	}

	_, err := self.db.Query("INSERT INTO channel VALUES (?)", name)
	if err != nil {
		log.Fatal("Could not add channel #" + name + " to the database: " + err.Error())
	}

	self.channels[name] = channel
}

func (self *channelManager) removeChannel(channel *twitch.Channel) {
	name := channel.Name

	_, ok := self.channels[name]
	if !ok {
		return
	}

	_, err := self.db.Query("DELETE FROM channel WHERE name = ?", name)
	if err != nil {
		log.Fatal("Could not remove channel #" + name + " from the database: " + err.Error())
	}

	delete(self.channels, name)
}
