package twitch

// a listener function is just any function (so, actually, `func(msg interface{})`)
type listenerFunc interface{}

// maps events to lists of listeners
type listenerMap map[string][]listenerFunc

// the exported interface to the dispatching
type Dispatcher interface {
	AddListener(event string, f listenerFunc) Listener
	TriggerEvent(event string, visitor walker)

	OnTextMessage(f TextHandlerFunc)     Listener
	OnTwitchMessage(f TwitchHandlerFunc) Listener
	OnJoin(f JoinHandlerFunc)            Listener
	OnPart(f JoinHandlerFunc)            Listener

	HandleTextMessage(msg TextMessage)
	HandleTwitchMessage(msg TwitchMessage)
	HandleJoin(c *Channel)
	HandlePart(c *Channel)
}

type dispatcher struct {
	listeners listenerMap
}

// the struct returned when adding new listeners;
// do not store a reference to the list in which this listeners is placed,
// as the list moved around when listeners are removed.
type Listener struct {
	dispatcher Dispatcher
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

func NewDispatcher() Dispatcher {
	return &dispatcher{make(map[string][]listenerFunc)}
}

func (d *dispatcher) OnTextMessage(f TextHandlerFunc)     Listener { return d.AddListener("TEXT", f)   }
func (d *dispatcher) OnTwitchMessage(f TwitchHandlerFunc) Listener { return d.AddListener("TWITCH", f) }
func (d *dispatcher) OnJoin(f JoinHandlerFunc)            Listener { return d.AddListener("JOIN", f)   }
func (d *dispatcher) OnPart(f JoinHandlerFunc)            Listener { return d.AddListener("PART", f)   }

func (d *dispatcher) HandleTextMessage(msg TextMessage)     { d.TriggerEvent("TEXT",   func(listener interface{}) { listener.(TextHandlerFunc)(msg)   }) }
func (d *dispatcher) HandleTwitchMessage(msg TwitchMessage) { d.TriggerEvent("TWITCH", func(listener interface{}) { listener.(TwitchHandlerFunc)(msg) }) }
func (d *dispatcher) HandleJoin(c *Channel)                 { d.TriggerEvent("JOIN",   func(listener interface{}) { listener.(JoinHandlerFunc)(c)     }) }
func (d *dispatcher) HandlePart(c *Channel)                 { d.TriggerEvent("PART",   func(listener interface{}) { listener.(JoinHandlerFunc)(c)     }) }

// private helpers

func (d *dispatcher) AddListener(event string, f listenerFunc) Listener {
	l, exists := d.listeners[event]

	if !exists {
		l = make([]listenerFunc, 0)
	}

	d.listeners[event] = append(l, f)

	return Listener{d, event, len(d.listeners[event]) - 1}
}

func (d *dispatcher) TriggerEvent(event string, visitor walker) {
	l, exists := d.listeners[event]

	if !exists {
		return
	}

	for _, listener := range l {
		visitor(listener)
	}
}
