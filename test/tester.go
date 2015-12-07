package test

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type pluginBuilder func() bot.Plugin

type Tester struct {
	file           io.Reader
	config         *bot.Configuration
	db             *sqlx.DB
	pluginBuilders map[string]pluginBuilder
}

func NewTester(file io.Reader, config *bot.Configuration, db *sqlx.DB) *Tester {
	return &Tester{
		file:           file,
		config:         config,
		db:             db,
		pluginBuilders: make(map[string]pluginBuilder),
	}
}

func (test *Tester) AddPlugin(name string, builder pluginBuilder) {
	test.pluginBuilders[name] = builder
}

var injectedMessage = regexp.MustCompile(`< \[(#[a-z0-9_]+)\] ([$%&@!~+]*[a-z0-9_]+): (.+)$`)
var expectedMessage = regexp.MustCompile(`> \[(#[a-z0-9_]+)\] ([$%&@!~+]*[a-z0-9_]+): (.+)$`)

func (test *Tester) WipeDatabase() {
	rows, _ := test.db.Queryx("SHOW TABLES")
	for rows.Next() {
		var table string
		rows.Scan(&table)

		test.db.MustExec("DELETE FROM `" + table + "` WHERE 1")
	}
}

func (test *Tester) Run(t *testing.T) {
	log := &fakeLog{}
	tc := &fakeClient{
		incoming: make(chan twitch.IncomingMessage),
		outgoing: make(chan twitch.OutgoingMessage, 10),
		ready:    make(chan struct{}),
	}

	testBot, _ := bot.NewKabukibot(tc, log, test.db, test.config)

	lineNr := 0
	lastLine := ""

	scanner := bufio.NewScanner(test.file)
	for scanner.Scan() {
		line := scanner.Text()
		lineNr++

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)

		switch parts[0] {
		// load a plugin
		case "plugin":
			builder, exists := test.pluginBuilders[parts[1]]
			if !exists {
				t.Errorf("[line %d] plugin could not be found: %s", lineNr, parts[1])
			}

			testBot.AddPlugin(builder())

		// connect to the chat, join initial chats
		case "connect":
			err := testBot.Connect()
			if err != nil {
				t.Errorf("[line %d] could not connect: %s", lineNr, err.Error())
			}

			go testBot.Work()

			// wait a bit for everything to settle, especially the join on the bot channel
			<-time.After(50 * time.Millisecond)

		// join a channel
		case "join":
			<-testBot.Join(parts[1])
			<-time.After(50 * time.Millisecond)

		// wait a bit (for operations that naturally take a few milliseconds)
		case "wait":
			duration := 50 * time.Millisecond

			if len(parts) > 1 {
				d, err := time.ParseDuration(parts[1])
				if err != nil {
					t.Error(err)
				}

				duration = d
			}

			<-time.After(duration)

		// inject a message, aka the bot receives a message
		case "<":
			matched := injectedMessage.FindStringSubmatch(line)
			if len(matched) != 4 {
				t.Errorf("[line %d] invalid line: '%s'", lineNr, line)
			}

			tc.incoming <- twitch.TextMessage{
				Channel: matched[1],
				User:    twitch.User{Name: matched[2]},
				Text:    matched[3],
			}

		// assert a response
		case ">":
			timeout := time.After(50 * time.Millisecond)
			matched := expectedMessage.FindStringSubmatch(line)
			if len(matched) != 4 {
				t.Errorf("[line %d] invalid line: '%s'", lineNr, line)
			}

			select {
			case actual := <-tc.outgoing:
				asserted, okay := actual.(twitch.TextMessage)
				if !okay {
					t.Errorf("[line %d] expected to receive '%s', but did not get a text message. Got %t instead.", lineNr, line, actual)
				}

				if asserted.Channel != matched[1] {
					t.Errorf("[line %d] expected to message in %s, but got one in %s instead.", lineNr, matched[1], asserted.Channel)
				}

				regex := regexp.MustCompile(matched[3])
				if !regex.MatchString(asserted.Text) {
					t.Errorf("[line %d] expected match `%s`, but got '%s' instead.", lineNr, matched[3], asserted.Text)
				}

			case <-timeout:
				t.Errorf("[line %d] expected to receive '%s', but got no message at all.", lineNr, line)
			}

		// expect silence
		case "silence":
			timeout := time.After(100 * time.Millisecond)

			select {
			case l := <-tc.outgoing:
				t.Errorf("[line %d] expected silence in response to '%s', but got a message: '%#v'", lineNr, lastLine, l)
			case <-timeout:
			}
		}

		lastLine = line
	}

	// shutdown
	testBot.Shutdown()
}
