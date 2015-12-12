package test

import (
	"time"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type fakeClient struct {
	incoming chan twitch.IncomingMessage
	outgoing chan twitch.OutgoingMessage
	ready    chan struct{}
}

func (c *fakeClient) Connect() error {
	close(c.ready)
	return nil
}

func (c *fakeClient) Disconnect() error {
	close(c.incoming)
	return nil
}

func (c *fakeClient) Incoming() <-chan twitch.IncomingMessage {
	return c.incoming
}

func (c *fakeClient) Ready() <-chan struct{} {
	return c.ready
}

func (c *fakeClient) QueueLen() int {
	return 0
}

func (c *fakeClient) MessagesSent() uint64 {
	return 0
}

func (c *fakeClient) MessagesReceived() uint64 {
	return 0
}

func (c *fakeClient) Send(msg twitch.OutgoingMessage) <-chan bool {
	asserted, okay := msg.(twitch.JoinMessage)
	if okay {
		// respond to a JOIN with a JOIN
		c.incoming <- asserted
	} else {
		asserted2, okay := msg.(twitch.PartMessage)
		if okay {
			// respond to a PART with a PART, but wait a bit because Kabukibot doesn't
			// like it if the signal for "i left the channel" comes before it even had
			// a chance to process the "i sent the PART request" event.
			go func() {
				<-time.After(100 * time.Millisecond)
				c.incoming <- asserted2
			}()
		} else {
			// send all other messages
			c.outgoing <- msg
		}
	}

	cn := make(chan bool, 1)
	cn <- true
	close(cn)

	return cn
}
