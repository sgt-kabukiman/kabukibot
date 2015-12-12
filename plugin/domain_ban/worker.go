package domain_ban

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dchest/validator"
	"github.com/jmoiron/sqlx"
	"github.com/mvdan/xurls"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

var nilTimeout = 0 * time.Second
var minTimeout = 1 * time.Second
var maxTimeout = 365 * 24 * time.Hour

type ban struct {
	Type    string
	Timeout time.Duration
	Counter int
}

type worker struct {
	plugin.NilWorker

	channel     string
	acl         *bot.ACL
	db          *sqlx.DB
	bans        map[string]ban
	syncing     chan struct{}
	stopSyncing chan struct{}
	mutex       sync.RWMutex
}

type domainBanDbStruct struct {
	Domain  string
	Bantype string
	Counter int
}

func (self *worker) Enable() {
	list := make([]domainBanDbStruct, 0)
	self.db.Select(&list, "SELECT domain, bantype, counter FROM domain_ban WHERE channel = ? ORDER BY domain", self.channel)

	self.bans = make(map[string]ban)

	for _, item := range list {
		b := ban{
			Type:    item.Bantype,
			Timeout: nilTimeout,
			Counter: item.Counter,
		}

		// type is either "ban" or "timeout:N", with N being the number of seconds to time out
		if b.Type != "ban" {
			parts := strings.Split(b.Type, ":")

			if len(parts) == 2 && parts[0] == "timeout" {
				b.Type = "timeout"
				b.Timeout = *bot.ParseDuration(parts[1], &minTimeout, &maxTimeout)
			} else {
				b.Type = "ban"
			}
		}

		self.bans[item.Domain] = b
	}

	// if we for some reason are already syncing, stop now
	if self.syncing != nil {
		self.Disable()
	}

	self.syncing = make(chan struct{})
	self.stopSyncing = make(chan struct{})

	go self.worker()
}

func (self *worker) Disable() {
	close(self.stopSyncing)
	<-self.syncing
}

func (self *worker) Permissions() []string {
	return []string{"configure_domain_bans"}
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	cmd := msg.Command()
	if cmd == "ban_domain" || cmd == "unban_domain" || cmd == "banned_domains" {
		if self.acl.IsAllowed(msg.User, "configure_domain_bans") {
			if cmd == "banned_domains" {
				self.bannedDomains(sender)
			} else {
				args := msg.Arguments()
				if len(args) == 0 {
					sender.Respond("no domain given.")
					return
				}

				domain := strings.ToLower(args[0])
				if validator.ValidateDomainByResolvingIt(domain) != nil {
					sender.Respond("this domain name seems invalid to me.")
					return
				}

				if cmd == "ban_domain" {
					self.banDomain(domain, args[1:], sender)
				} else {
					self.unbanDomain(domain, sender)
				}
			}
		}

		msg.SetProcessed()
	} else {
		self.textMessage(msg, sender)
	}
}

func (self *worker) bannedDomains(sender bot.Sender) {
	var bans []string

	self.mutex.RLock()

	for domain, ban := range self.bans {
		var t string

		if ban.Type == "timeout" {
			t = fmt.Sprintf("%s (%s t/o)", domain, ban.Timeout.String())
		} else {
			t = fmt.Sprintf("%s (ban)", domain)
		}

		bans = append(bans, t)
	}

	self.mutex.RUnlock()

	if len(bans) == 0 {
		sender.Respond("no domains are banned yet.")
	} else {
		sender.Respond("the following domains are forbidden: " + bot.HumanJoin(bans, ", "))
	}
}

func (self *worker) banDomain(domain string, args []string, sender bot.Sender) {
	bantype := "ban"
	timeout := nilTimeout

	if len(args) >= 2 && (args[0] == "timeout" || args[0] == "to") {
		parsed := bot.ParseDuration(strings.Join(args[1:], ""), nil, nil)
		if parsed == nil || *parsed < minTimeout {
			sender.Respond("invalid timeout time given. Expected a value like 50s or 3d.")
			return
		}

		if *parsed > maxTimeout {
			sender.Respond(fmt.Sprintf("invalid timeout time given. Use values smaller than %s.", maxTimeout.String()))
			return
		}

		// round to seconds, just in case
		bantype = "timeout"
		timeout = time.Duration(parsed.Seconds() * float64(time.Second))
	}

	self.mutex.Lock()

	b, exists := self.bans[domain]
	if !exists {
		b = ban{Counter: 0}
	}

	b.Type = bantype
	b.Timeout = timeout

	self.bans[domain] = b

	self.mutex.Unlock()

	if timeout > 0 {
		sender.Respond(fmt.Sprintf("links to %s will be timed out for %s.", domain, bot.FormatDuration(timeout, true)))
	} else {
		sender.Respond(fmt.Sprintf("links to %s will be *banned*.", domain))
	}

	// let the worker take care of writing this to the database
}

func (self *worker) unbanDomain(domain string, sender bot.Sender) {
	self.mutex.Lock()

	b, exists := self.bans[domain]
	if !exists {
		self.mutex.Unlock()
		sender.Respond(domain + " was not banned in the first place.")
		return
	}

	if b.Type == "timeout" {
		sender.Respond(fmt.Sprintf("links to %s will no longer be timed out.", domain))
	} else {
		sender.Respond(fmt.Sprintf("links to %s will no longer be banned.", domain))
	}

	delete(self.bans, domain)
	self.mutex.Unlock()

	// let the worker take care of writing this to the database
}

func (self *worker) textMessage(msg *bot.TextMessage, sender bot.Sender) {
	if len(self.bans) == 0 || msg.IsFromBroadcaster() || msg.IsFromOperator() || msg.IsFromBot() {
		return
	}

	user := msg.User
	t := user.Type

	if t == twitch.Moderator || t == twitch.GlobalModerator || t == twitch.TwitchStaff || t == twitch.TwitchAdmin {
		return
	}

	links := xurls.Relaxed.FindAllString(msg.Text, -1)
	if len(links) == 0 {
		return
	}

	var evilDomains []string

	self.mutex.RLock()

	for _, link := range links {
		link = strings.ToLower(link)

		_, exists := self.bans[link]
		if exists {
			evilDomains = append(evilDomains, link)
		} else {
			if !strings.Contains(link, "://") {
				link = "http://" + link
			}

			parsed, err := url.Parse(link)
			if err == nil && len(parsed.Host) > 0 {
				_, exists = self.bans[parsed.Host]
				if exists {
					evilDomains = append(evilDomains, parsed.Host)
				}
			}
		}
	}

	self.mutex.RUnlock()

	if len(evilDomains) == 0 {
		return
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	// find the most severe sentence
	action := ban{Type: "timeout"}
	worstDomain := ""

	for _, domain := range evilDomains {
		ban, _ := self.bans[domain]

		// nothing can be more severe than a permanent ban
		if ban.Type == "ban" {
			action = ban
			worstDomain = domain
			break
		} else if ban.Timeout > action.Timeout {
			action = ban
			worstDomain = domain
		}
	}

	// kick the offender
	name := strings.ToLower(msg.User.Name)

	if action.Type == "ban" {
		sender.Ban(name)
		sender.Respond("posting that link was a bad idea and got you permanently banned.")
	} else {
		sender.Timeout(name, int(action.Timeout.Seconds()))
		sender.Respond(fmt.Sprintf(
			"posting that link was a bad idea and got you timed out for %s.",
			bot.FormatDuration(action.Timeout, true),
		))
	}

	// and count that we hit one
	action.Counter++
	self.bans[worstDomain] = action
}

func (self *worker) worker() {
	defer close(self.syncing)

	for {
		select {
		case <-time.After(5 * time.Minute):
			self.sync()

		case <-self.stopSyncing:
			self.sync()
			return
		}
	}
}

func (self *worker) sync() {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	self.db.Exec("DELETE FROM domain_ban WHERE channel = ?", self.channel)

	for domain, ban := range self.bans {
		t := ban.Type

		if t == "timeout" {
			t += ":" + bot.FormatDuration(ban.Timeout, false)
		}

		self.db.Exec("INSERT INTO domain_ban (channel, domain, bantype, counter) VALUES (?, ?, ?)", self.channel, domain, t, ban.Counter)
	}
}
