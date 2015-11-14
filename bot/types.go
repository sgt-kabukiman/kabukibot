package bot

// type Command interface {
// 	twitch.Message

// 	Command() string
// 	Args() []string
// }

// type CommandHandlerFunc func(Command)

// type command struct {
// 	twitch.Message

// 	cmd  string
// 	args []string
// }

// func (cmd *command) Command() string { return cmd.cmd }
// func (cmd *command) Args() []string  { return cmd.args }

// type Response interface {
// 	ResponseTo() twitch.Message
// 	Channel() *twitch.Channel
// 	Text() string
// }

// type ResponseHandlerFunc func(Response)

// type response struct {
// 	to      twitch.Message
// 	channel *twitch.Channel
// 	text    string
// }

// func (r *response) ResponseTo() twitch.Message { return r.to }
// func (r *response) Channel() *twitch.Channel   { return r.channel }
// func (r *response) Text() string               { return r.text }

// type Plugin interface {
// 	Setup(*Kabukibot, Dispatcher)
// }

// type GlobalPlugin interface {
// 	Plugin
// }

// type ChannelPlugin interface {
// 	Plugin

// 	Key() string
// 	Permissions() []string
// 	Load(*twitch.Channel, *Kabukibot, Dispatcher)
// 	Unload(*twitch.Channel, *Kabukibot, Dispatcher)
// }
