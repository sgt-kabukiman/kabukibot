package plugin

import (
	"bufio"
	"io"
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
	. "github.com/smartystreets/goconvey/convey"
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

func (c *fakeClient) Send(msg twitch.OutgoingMessage) <-chan bool {
	asserted, okay := msg.(twitch.JoinMessage)
	if okay {
		// respond to a JOIN with a JOIN
		c.incoming <- asserted
	} else {
		// send all other messages
		c.outgoing <- msg
	}

	cn := make(chan bool)
	close(cn)

	return cn
}

type fakeLog struct{}

func (f *fakeLog) SetLevel(int)                   {}
func (f *fakeLog) Debug(string, ...interface{})   {}
func (f *fakeLog) Info(string, ...interface{})    {}
func (f *fakeLog) Warning(string, ...interface{}) {}
func (f *fakeLog) Warn(string, ...interface{})    {}
func (f *fakeLog) Error(string, ...interface{})   {}
func (f *fakeLog) Fatal(string, ...interface{})   {}

type pluginBuilder func() bot.Plugin

type tester struct {
	file           io.Reader
	pluginBuilders map[string]pluginBuilder
}

func NewTester(file io.Reader) *tester {
	return &tester{
		file:           file,
		pluginBuilders: make(map[string]pluginBuilder),
	}
}

func (test *tester) AddPlugin(name string, builder pluginBuilder) {
	test.pluginBuilders[name] = builder
}

func (test *tester) Run() {
	broadcaster := "owner"
	testChannel := "#" + broadcaster
	operator := "op"

	log := &fakeLog{}
	config := &bot.Configuration{
		CommandPrefix: "k_",
		Operator:      operator,
		Account: struct {
			Username string
			Password string
		}{
			Username: "bot",
			Password: "",
		},
		Database: struct {
			DSN string `yaml:"DSN"`
		}{
			DSN: "develop:develop@/kabukibot_dev",
		},
	}

	tc := &fakeClient{
		incoming: make(chan twitch.IncomingMessage),
		outgoing: make(chan twitch.OutgoingMessage, 10),
		ready:    make(chan struct{}),
	}

	testBot, err := bot.NewKabukibot(tc, log, config)
	So(err, ShouldBeNil)

	scanner := bufio.NewScanner(test.file)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "plugin ") {
			pluginName := strings.TrimPrefix(line, "plugin ")
			builder, exists := test.pluginBuilders[pluginName]
			So(exists, ShouldBeTrue)

			testBot.AddPlugin(builder())
		} else if line == "connect" {
			err = testBot.Connect()
			So(err, ShouldBeNil)

			go testBot.Work()

			// wait a bit for everything to settle, especially the join on the bot channel
			<-time.After(100 * time.Millisecond)

			<-testBot.Join(testChannel)
		} else if strings.HasPrefix(line, "< ") {
			parts := strings.SplitN(line, " ", 3)
			msg := twitch.TextMessage{
				Channel: testChannel,
				Text:    parts[2],
				User:    twitch.User{Name: parts[1]},
			}

			tc.incoming <- msg
		} else if strings.HasPrefix(line, "> ") {
			parts := strings.SplitN(line, " ", 2)
			expected := parts[1]
			actual := <-tc.outgoing

			asserted, okay := actual.(twitch.TextMessage)
			So(okay, ShouldBeTrue)
			So(asserted.Text, ShouldEqual, expected)
		}
	}

	So(scanner.Err(), ShouldBeNil)

	// shutdown
	testBot.Shutdown()
}
