package main

import (
	"container/list"
)

type ListenerMap map[string]*list.List

type Dispatcher struct {
	listeners ListenerMap
}

type Listener struct {
	list    *list.List
	element *list.Element
}

func (l *Listener) Remove() {
	if l.element != nil {
		l.list.Remove(l.element)
		l.element = nil
	}
}

type walker func(handler *list.Element)

func NewDispatcher() *Dispatcher {
	return &Dispatcher{make(map[string]*list.List)}
}

func (d *Dispatcher) OnEveryMessage(f MessageHandlerFunc) Listener { return d.on("MESSAGE", f)   }
func (d *Dispatcher) OnTextMessage(f TextHandlerFunc)     Listener { return d.on("TEXT", f)      }
func (d *Dispatcher) OnTwitchMessage(f TwitchHandlerFunc) Listener { return d.on("TWITCH", f)    }
func (d *Dispatcher) OnModeMessage(f ModeHandlerFunc)     Listener { return d.on("MODE", f)      }
func (d *Dispatcher) OnCommand(f CommandHandlerFunc)      Listener { return d.on("COMMAND", f)   }
func (d *Dispatcher) OnProcessed(f ProcessedHandlerFunc)  Listener { return d.on("PROCESSED", f) }
func (d *Dispatcher) OnResponse(f ResponseHandlerFunc)    Listener { return d.on("RESPONSE", f)  }

func (d *Dispatcher) HandleMessage(msg Message) {
	d.handle("MESSAGE", func(handler *list.Element) {
		callback := handler.Value.(MessageHandlerFunc)
		callback(&msg)
	})
}

func (d *Dispatcher) HandleTextMessage(msg TextMessage) {
	d.handle("TEXT", func(handler *list.Element) {
		callback := handler.Value.(TextHandlerFunc)
		callback(&msg)
	})
}

func (d *Dispatcher) HandleTwitchMessage(msg TwitchMessage) {
	d.handle("TWITCH", func(handler *list.Element) {
		callback := handler.Value.(TwitchHandlerFunc)
		callback(&msg)
	})
}

func (d *Dispatcher) HandleModeMessage(msg ModeMessage) {
	d.handle("MODE", func(handler *list.Element) {
		callback := handler.Value.(ModeHandlerFunc)
		callback(&msg)
	})
}

func (d *Dispatcher) HandleCommand(command string, args []string, msg Message) {
	d.handle("COMMAND", func(handler *list.Element) {
		callback := handler.Value.(CommandHandlerFunc)
		callback(command, args, &msg)
	})
}

func (d *Dispatcher) HandleProcessed(msg Message) {
	d.handle("PROCESSED", func(handler *list.Element) {
		callback := handler.Value.(ProcessedHandlerFunc)
		callback(&msg)
	})
}

// private helpers

func (d *Dispatcher) on(event string, f interface{}) Listener {
	l, exists := d.listeners[event]

	if !exists {
		l = list.New()
		d.listeners[event] = l
	}

	return Listener{l, l.PushBack(f)}
}

func (d *Dispatcher) handle(event string, visitor walker) {
	l, exists := d.listeners[event]

	if !exists {
		return
	}

	for handler := l.Front(); handler != nil; handler = handler.Next() {
		visitor(handler)
	}
}
