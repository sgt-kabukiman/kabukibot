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
		case "plugin":
			test.pluginCommand(t, testBot, lineNr, parts[1:])
		case "connect":
			test.connectCommand(t, testBot, lineNr)
		case "join":
			test.joinCommand(t, testBot, lineNr, parts[1:])
		case "wait":
			test.waitCommand(t, testBot, lineNr, parts[1:])
		case "<":
			test.sendCommand(t, testBot, lineNr, line, tc)
		case ">":
			test.receiveCommand(t, testBot, lineNr, line, tc)
		case "silence":
			test.silenceCommand(t, testBot, lineNr, lastLine, tc)
		}

		lastLine = line
	}

	// shutdown
	testBot.Shutdown()
}

func (test *Tester) pluginCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, args []string) {
	plugin := args[0]

	builder, exists := test.pluginBuilders[plugin]
	if !exists {
		t.Errorf("[line %d] plugin could not be found: %s", lineNr, plugin)
	}

	bot.AddPlugin(builder())
}

func (test *Tester) connectCommand(t *testing.T, bot *bot.Kabukibot, lineNr int) {
	err := bot.Connect()
	if err != nil {
		t.Errorf("[line %d] could not connect: %s", lineNr, err.Error())
	}

	go bot.Work()

	// wait a bit for everything to settle, especially the join on the bot channel
	<-time.After(50 * time.Millisecond)
}

func (test *Tester) joinCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, args []string) {
	<-bot.Join(args[0])
	<-time.After(50 * time.Millisecond)
}

func (test *Tester) waitCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, args []string) {
	duration := 50 * time.Millisecond

	if len(args) > 0 {
		d, err := time.ParseDuration(args[0])
		if err != nil {
			t.Error(err)
		}

		duration = d
	}

	<-time.After(duration)
}

func (test *Tester) sendCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, line string, client *fakeClient) {
	matched := injectedMessage.FindStringSubmatch(line)
	if len(matched) != 4 {
		t.Errorf("[line %d] invalid line: '%s'", lineNr, line)
	}

	client.incoming <- twitch.TextMessage{
		Channel: matched[1],
		User:    twitch.User{Name: matched[2]},
		Text:    matched[3],
	}
}

func (test *Tester) receiveCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, line string, client *fakeClient) {
	timeout := time.After(50 * time.Millisecond)
	matched := expectedMessage.FindStringSubmatch(line)
	if len(matched) != 4 {
		t.Errorf("[line %d] invalid line: '%s'", lineNr, line)
	}

	select {
	case actual := <-client.outgoing:
		asserted, okay := actual.(twitch.TextMessage)
		if !okay {
			t.Errorf("[line %d] expected to receive '%s', but did not get a text message. Got %t instead.", lineNr, line, actual)
		}

		if asserted.Channel != matched[1] {
			t.Errorf("[line %d] expected to message in %s, but got one in %s instead.", lineNr, matched[1], asserted.Channel)
		}

		regex := regexp.MustCompile("^" + matched[3] + "$")
		if !regex.MatchString(asserted.Text) {
			t.Errorf("[line %d] expected match `%s`, but got '%s' instead.", lineNr, matched[3], asserted.Text)
		}

	case <-timeout:
		t.Errorf("[line %d] expected to receive '%s', but got no message at all.", lineNr, line)
	}
}

func (test *Tester) silenceCommand(t *testing.T, bot *bot.Kabukibot, lineNr int, lastLine string, client *fakeClient) {
	timeout := time.After(100 * time.Millisecond)

	select {
	case l := <-client.outgoing:
		t.Errorf("[line %d] expected silence in response to '%s', but got a message: '%#v'", lineNr, lastLine, l)
	case <-timeout:
	}
}
