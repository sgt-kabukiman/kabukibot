package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/plugin/blacklist"
	"github.com/sgt-kabukiman/kabukibot/twitch"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPlugins(t *testing.T) {
	// load configuration
	config, err := bot.LoadConfiguration("config-test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	// connect to database
	db, err := sqlx.Connect("mysql", config.Database.DSN)
	if err != nil {
		t.Fatalf(err.Error())
	}

	filepath.Walk("plugin", func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".test") {
			return nil
		}

		rel := strings.Replace(path[7:], "\\", "/", -1)

		Convey(rel, t, func() {
			file, err := os.Open(path)
			So(err, ShouldBeNil)
			defer file.Close()

			tester := newTester(file, config, db)
			initTester(tester)
			tester.WipeDatabase()
			tester.Run()
		})

		return nil
	})
}

func shouldMatchRegexp(actual interface{}, expected ...interface{}) string {
	expr := expected[0].(string)
	regex := regexp.MustCompile(expr)

	if regex.MatchString(actual.(string)) {
		return ""
	}

	return "Response '" + (actual.(string)) + "' should have matched `" + expr + "`."
}

func initTester(t *tester) {
	t.AddPlugin("blacklist", func() bot.Plugin {
		return blacklist.NewPlugin()
	})

	t.AddPlugin("log", func() bot.Plugin {
		return plugin.NewLogPlugin()
	})

	t.AddPlugin("ping", func() bot.Plugin {
		return plugin.NewPingPlugin()
	})

	t.AddPlugin("join", func() bot.Plugin {
		return plugin.NewJoinPlugin()
	})

	t.AddPlugin("acl", func() bot.Plugin {
		return plugin.NewACLPlugin()
	})

	t.AddPlugin("plugin_control", func() bot.Plugin {
		return plugin.NewPluginControlPlugin()
	})

	t.AddPlugin("speedruncom", func() bot.Plugin {
		return plugin.NewSpeedrunComPlugin()
	})

	t.AddPlugin("echo", func() bot.Plugin {
		return plugin.NewEchoPlugin()
	})

	t.AddPlugin("sysinfo", func() bot.Plugin {
		return plugin.NewSysInfoPlugin()
	})

	t.AddPlugin("dictionary", func() bot.Plugin {
		return plugin.NewDictionaryPlugin()
	})

	t.AddPlugin("domain_ban", func() bot.Plugin {
		return plugin.NewDomainBanPlugin()
	})

	t.AddPlugin("banhammer_bot", func() bot.Plugin {
		return plugin.NewBanhammerBotPlugin()
	})

	t.AddPlugin("emote_counter", func() bot.Plugin {
		return plugin.NewEmoteCounterPlugin()
	})

	t.AddPlugin("subhype", func() bot.Plugin {
		return plugin.NewSubHypePlugin()
	})

	t.AddPlugin("troll", func() bot.Plugin {
		return plugin.NewTrollPlugin()
	})

	t.AddPlugin("monitor", func() bot.Plugin {
		return plugin.NewMonitorPlugin()
	})

	t.AddPlugin("custom_commands", func() bot.Plugin {
		return plugin.NewCustomCommandsPlugin()
	})
}

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
	config         *bot.Configuration
	db             *sqlx.DB
	pluginBuilders map[string]pluginBuilder
}

func newTester(file io.Reader, config *bot.Configuration, db *sqlx.DB) *tester {
	return &tester{
		file:           file,
		config:         config,
		db:             db,
		pluginBuilders: make(map[string]pluginBuilder),
	}
}

func (test *tester) AddPlugin(name string, builder pluginBuilder) {
	test.pluginBuilders[name] = builder
}

var injectedMessage = regexp.MustCompile(`< \[(#[a-z0-9_]+)\] ([$%&@!~+]*[a-z0-9_]+): (.+)$`)
var expectedMessage = regexp.MustCompile(`> \[(#[a-z0-9_]+)\] ([$%&@!~+]*[a-z0-9_]+): (.+)$`)

func (test *tester) WipeDatabase() {
	rows, _ := test.db.Queryx("SHOW TABLES")
	for rows.Next() {
		var table string
		rows.Scan(&table)

		test.db.MustExec("DELETE FROM `" + table + "` WHERE 1")
	}
}

func (test *tester) Run() {
	log := &fakeLog{}
	tc := &fakeClient{
		incoming: make(chan twitch.IncomingMessage),
		outgoing: make(chan twitch.OutgoingMessage, 10),
		ready:    make(chan struct{}),
	}

	testBot, err := bot.NewKabukibot(tc, log, test.db, test.config)
	So(err, ShouldBeNil)

	scanner := bufio.NewScanner(test.file)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)

		switch parts[0] {
		// load a plugin
		case "plugin":
			builder, exists := test.pluginBuilders[parts[1]]
			So(exists, ShouldBeTrue)
			testBot.AddPlugin(builder())

		// connect to the chat, join initial chats
		case "connect":
			err = testBot.Connect()
			So(err, ShouldBeNil)

			go testBot.Work()

			// wait a bit for everything to settle, especially the join on the bot channel
			<-time.After(50 * time.Millisecond)

		// join a channel
		case "join":
			<-testBot.Join(parts[1])
			<-time.After(50 * time.Millisecond)

		// inject a message, aka the bot receives a message
		case "<":
			matched := injectedMessage.FindStringSubmatch(line)
			So(matched, ShouldHaveLength, 4)

			tc.incoming <- twitch.TextMessage{
				Channel: matched[1],
				User:    twitch.User{Name: matched[2]},
				Text:    matched[3],
			}

		// assert a response
		case ">":
			timeout := time.After(250 * time.Millisecond)
			matched := expectedMessage.FindStringSubmatch(line)
			So(matched, ShouldHaveLength, 4)

			select {
			case actual := <-tc.outgoing:
				asserted, okay := actual.(twitch.TextMessage)
				So(okay, ShouldBeTrue)
				So(asserted.Channel, ShouldEqual, matched[1])
				So(asserted.Text, shouldMatchRegexp, matched[3])

			case <-timeout:
				So(true, ShouldBeFalse)
			}

		// expect silence
		case "silence":
			timeout := time.After(250 * time.Millisecond)
			received := false

			select {
			case msg := <-tc.outgoing:
				fmt.Printf("bad: %#v\n", msg)
				received = true
			case <-timeout:
				received = false
			}

			So(received, ShouldBeFalse)
		}
	}

	So(scanner.Err(), ShouldBeNil)

	// shutdown
	testBot.Shutdown()
}
