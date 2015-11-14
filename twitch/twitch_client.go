package twitch

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sorcix/irc"
)

// a message on the queue, this is not what the outside world sees
type queueItem struct {
	message irc.Message
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
	incoming chan Message

	// list of ougtoing messages (sent by us)
	outgoing chan queueItem
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
		incoming:         make(chan Message, 50),
		outgoing:         make(chan queueItem, 50),
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

func (client *TwitchClient) Incoming() <-chan Message {
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
	client.Send(irc.Message{
		Command: irc.PASS,
		Params:  []string{client.password},
	})

	client.Send(irc.Message{
		Command: irc.NICK,
		Params:  []string{client.username},
	})

	client.Send(irc.Message{
		Command: irc.USER,
		Params:  []string{"kabukibot", "8", "*", client.username},
	})

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

func (client *TwitchClient) Send(msg irc.Message) <-chan bool {
	signal := make(chan bool, 1)
	outgoing := queueItem{msg, signal}

	// if queue is not full, then
	client.outgoing <- outgoing
	// else
	// 	signal <- false // means "not sent"
	// 	close(signal)

	return signal
}

func (client *TwitchClient) sender() {
	for {
		select {
		case msg := <-client.outgoing:
			fmt.Println("< " + msg.message.String())
			client.writer.Encode(&msg.message)

			// signal to the one who sent the message that it was in fact sent
			msg.signal <- true
			close(msg.signal)

		case <-client.stopSending:
			break
		}
	}

	close(client.stoppedSending)
}

func (client *TwitchClient) receiver() {
	reading := make(chan struct{})

	// a buffer between the raw irc input from the net and the goroutine channels
	buffer := make(chan string, 10)

	// fork a reader loop, which could block and needs special handling as it's not a channel
	// (but it will pump its messages into a channel)
	go func() {
		for {
			select {
			case <-client.stopReceiving:
				break

			default:
				// set a 5min timeout
				client.conn.SetDeadline(time.Now().Add(300 * time.Second))

				line, err := client.reader.ReadString('\n')
				if err != nil {
					break
				}

				buffer <- line
			}
		}

		close(reading)
	}()

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

			// hand it over to the message handler
			handler, ok := client.handlers[msg.Command]
			if ok {
				go handler(msg, tags)
			}

		case <-client.stopReceiving:
			break
		}
	}

	close(client.stoppedReceiving)
}
