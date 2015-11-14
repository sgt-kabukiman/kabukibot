package twitch

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sorcix/irc"
)

// buffer at most this many messages before dropping messages
// (this applies to OUTGOING messages)
const queueSize = 50

// a message on the queue, this is not what the outside world sees
type queueItem struct {
	message OutgoingMessage
	signal  chan bool
}

type TwitchClient struct {
	server   string
	username string
	password string

	// connection handling
	conn   net.Conn
	reader *bufio.Reader
	writer *irc.Encoder

	// handlers for incoming messages
	handlers map[string]HandlerFunc

	// time between two regular messages are sent
	delay time.Duration

	// this signal is sent when the client has sent the CAP REQ commands
	ready chan struct{}

	// this signal is sent when we disconnected
	alive chan struct{}

	// these are fired when .Disconnect() is called
	stopSending   chan struct{}
	stopReceiving chan struct{}

	// this is fired when .sender() / .receiver() stop
	stoppedSending   chan struct{}
	stoppedReceiving chan struct{}

	// on this channel incoming messages from the network are sent
	incoming chan IncomingMessage

	// list of ougtoing messages (sent by us)
	outgoing   chan queueItem
	queueLen   int
	queueSize  int
	queueMutex sync.Mutex
}

func NewTwitchClient(server string, username string, password string, delay time.Duration) *TwitchClient {
	client := &TwitchClient{
		server:           server,
		username:         username,
		password:         password,
		delay:            delay,
		conn:             nil,
		reader:           nil,
		writer:           nil,
		ready:            make(chan struct{}),
		alive:            make(chan struct{}),
		stopReceiving:    make(chan struct{}),
		stopSending:      make(chan struct{}),
		stoppedReceiving: make(chan struct{}),
		stoppedSending:   make(chan struct{}),
		incoming:         make(chan IncomingMessage, 50),
		outgoing:         make(chan queueItem, queueSize+10), // make a bit room so we never block, even when reaching queueSize
		queueSize:        queueSize,
		queueLen:         0,
		queueMutex:       sync.Mutex{},
	}

	// setup vital message listeners
	client.setupHandlers()

	return client
}

func (client *TwitchClient) Ready() <-chan struct{} {
	return client.ready
}

func (client *TwitchClient) Alive() <-chan struct{} {
	return client.alive
}

func (client *TwitchClient) Incoming() <-chan IncomingMessage {
	return client.incoming
}

func (client *TwitchClient) Connect() error {
	conn, err := net.Dial("tcp", client.server)
	if err != nil {
		return err
	}

	client.conn = conn
	client.reader = bufio.NewReader(conn) // we manually read to properly handle tags
	client.writer = irc.NewEncoder(conn)

	// start working on the queue
	go client.sender()

	// start receiving
	go client.receiver()

	// send login info before anything else
	client.Send(RawMessage{irc.Message{
		Command: irc.PASS,
		Params:  []string{client.password},
	}})

	client.Send(RawMessage{irc.Message{
		Command: irc.NICK,
		Params:  []string{client.username},
	}})

	client.Send(RawMessage{irc.Message{
		Command: irc.USER,
		Params:  []string{"kabukibot", "8", "*", client.username},
	}})

	return nil
}

func (client *TwitchClient) Disconnect() error {
	// stop the sender/receiver and wait for them to stop (maybe it will drain the
	// outgoing queue, maybe it won't, but let's give it time)
	close(client.stopReceiving)
	<-client.stoppedReceiving

	close(client.stopSending)
	<-client.stoppedSending

	// for all intents and purposes, we are not alive anymore
	close(client.alive)

	// close the IRC connection
	return client.conn.Close()
}

func (client *TwitchClient) Send(msg OutgoingMessage) <-chan bool {
	signal := make(chan bool, 1)

	client.queueMutex.Lock()

	// silenty drop the message so our queue doesn't grow infinitely
	if client.queueLen >= client.queueSize {
		signal <- false
		close(signal)
	} else {
		client.outgoing <- queueItem{msg, signal}
		client.queueLen++
	}

	client.queueMutex.Unlock()

	return signal
}

func (client *TwitchClient) sender() {
	for {
		select {
		case msg := <-client.outgoing:
			ircMsg := msg.message.IrcMessage()
			fmt.Println("< " + ircMsg.String())
			client.writer.Encode(ircMsg)

			// signal to the one who sent the message that it was in fact sent
			msg.signal <- true
			close(msg.signal)

			client.queueMutex.Lock()
			client.queueLen--
			client.queueMutex.Unlock()

			// wait a bit
			<-time.After(client.delay)

		case <-client.stopSending:
			close(client.stoppedSending)
			return
		}
	}
}

func (client *TwitchClient) receiver() {
	reading := make(chan struct{})

	// a buffer between the raw irc input from the net and the goroutine channels
	buffer := make(chan string, 10)

	// fork a reader loop, which could block and needs special handling as it's not a channel
	// (but it will pump its messages into a channel)
	go func() {
		defer close(reading)

		for {
			select {
			case <-client.stopReceiving:
				return

			default:
				// set a 5min timeout
				client.conn.SetDeadline(time.Now().Add(300 * time.Second))

				line, err := client.reader.ReadString('\n')
				if err != nil {
					return
				}

				buffer <- line
			}
		}
	}()

	defer close(client.stoppedReceiving)

	for {
		select {
		case rawLine := <-buffer:
			fmt.Println("> " + strings.TrimSpace(rawLine))

			// if the message begins with a '@', we have some tags (IRCv3). The default
			// IRC decoder will not have properly detected it and mangled its output.
			// We fix that by manually splitting the tags from the rest of the message
			// and parse each part individually.
			tags := make(irc.Tags)
			msg := &irc.Message{}

			if strings.HasPrefix(rawLine, "@") {
				parts := strings.SplitN(rawLine, " ", 2)

				tags = irc.ParseTags(strings.TrimPrefix(parts[0], "@"))
				msg = irc.ParseMessage(parts[1])
			} else {
				msg = irc.ParseMessage(rawLine)
			}

			// hand it over to the message handler;
			// this could be done in goroutines by simply doing "go handler(...)",
			// but then we could interpret messages out-of-order. There are enough
			// buffers and goroutines already, so forking here is not really
			// needed anyway.
			handler, ok := client.handlers[msg.Command]
			if ok {
				handler(msg, tags)
			}

		case <-client.stopReceiving:
			return
		}
	}
}
