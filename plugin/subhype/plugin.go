package subhype

import "github.com/sgt-kabukiman/kabukibot/bot"

type pluginStruct struct {
	dict *bot.Dictionary
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{nil}
}

func (self *pluginStruct) Name() string {
	return "subhype"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.dict = bot.Dictionary()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		dict:    self.dict,
		message: self.dict.Get(subhypeKey(channel.Name())),
	}
}
