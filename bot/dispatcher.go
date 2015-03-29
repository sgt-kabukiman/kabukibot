package bot

// a listener function is just any function (so, actually, `func(msg interface{})`)
type listenerFunc interface{}

// maps events to lists of listeners
type listenerMap map[string][]listenerFunc

// the exported interface to the dispatching
type Dispatcher struct {
	listeners listenerMap
}

// the struct returned when adding new listeners;
// do not store a reference to the list in which this listeners is placed,
// as the list moved around when listeners are removed.
type Listener struct {
	dispatcher *Dispatcher
	event      string
	position   int
}

func (l *Listener) Remove() bool {
	// if l.dispatcher != nil {
	// 	list, exists := l.dispatcher.listeners[l.event]

	// 	if exists && l.position >= 0 && l.position < len(list) {
	// 		l.dispatcher.listeners[l.event] = append(list[:l.position], list[(l.position + 1):]...)
	// 		l.dispatcher = nil

	// 		return true
	// 	}
	// }

	panic("Implement me!")
	return false
}

type walker func(listener interface{})

func NewDispatcher() *Dispatcher {
	return &Dispatcher{make(map[string][]listenerFunc)}
}

func (d *Dispatcher) OnEveryMessage(f MessageHandlerFunc)   Listener { return d.on("MESSAGE", f)   }
func (d *Dispatcher) OnTextMessage(f TextHandlerFunc)       Listener { return d.on("TEXT", f)      }
func (d *Dispatcher) OnTwitchMessage(f TwitchHandlerFunc)   Listener { return d.on("TWITCH", f)    }
func (d *Dispatcher) OnModeMessage(f ModeHandlerFunc)       Listener { return d.on("MODE", f)      }
func (d *Dispatcher) OnCommandMessage(f CommandHandlerFunc) Listener { return d.on("COMMAND", f)   }
func (d *Dispatcher) OnProcessed(f ProcessedHandlerFunc)    Listener { return d.on("PROCESSED", f) }
func (d *Dispatcher) OnResponse(f ResponseHandlerFunc)      Listener { return d.on("RESPONSE", f)  }

func (d *Dispatcher) HandleMessage(msg Message)               { d.handle("MESSAGE",   func(listener interface{}) { listener.(MessageHandlerFunc)(msg) }) }
func (d *Dispatcher) HandleTextMessage(msg TextMessage)       { d.handle("TEXT",      func(listener interface{}) { listener.(TextHandlerFunc)(msg)    }) }
func (d *Dispatcher) HandleTwitchMessage(msg TwitchMessage)   { d.handle("TWITCH",    func(listener interface{}) { listener.(TwitchHandlerFunc)(msg)  }) }
func (d *Dispatcher) HandleModeMessage(msg ModeMessage)       { d.handle("MODE",      func(listener interface{}) { listener.(ModeHandlerFunc)(msg)    }) }
func (d *Dispatcher) HandleCommandMessage(msg CommandMessage) { d.handle("COMMAND",   func(listener interface{}) { listener.(CommandHandlerFunc)(msg) }) }
func (d *Dispatcher) HandleProcessed(msg Message)             { d.handle("PROCESSED", func(listener interface{}) { listener.(MessageHandlerFunc)(msg) }) }

// private helpers

func (d *Dispatcher) on(event string, f listenerFunc) Listener {
	l, exists := d.listeners[event]

	if !exists {
		l = make([]listenerFunc, 0)
	}

	d.listeners[event] = append(l, f)

	return Listener{d, event, len(d.listeners[event]) - 1}
}

func (d *Dispatcher) handle(event string, visitor walker) {
	l, exists := d.listeners[event]

	if !exists {
		return
	}

	for _, listener := range l {
		visitor(listener)
	}
}
