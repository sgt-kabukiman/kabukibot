package bot

import "bufio"
import "encoding/json"
import "fmt"
import "net/http"
import "regexp"
// import "runtime"
import "sort"
import "strings"
// import "time"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type EmoteManager interface {
	FindEmotesInMessage(msg twitch.TextMessage) emoteList
	UpdateEmotes() error
}

type emoteList []string

type emoteManager struct {
	emotes struct {
		global       emoteList
		subscribers  map[int]*emoteList
		frankerfacez map[string]*emoteList
	}
	regexes struct {
		channels    map[string]*regexp.Regexp
		subscribers map[string]*regexp.Regexp
	}
}

type twitchEmoticons struct {
	Emoticons []struct {
		Regex  string
		Images []struct {
			EmoticonSet *int `json:"emoticon_set"` // use a pointer because this can be null
		}
	}
}

func NewEmoteManager() EmoteManager {
	em := emoteManager{}
	em.reset()

	return &em
}

func (self *emoteManager) reset() {
	self.emotes.subscribers  = make(map[int]*emoteList)
	self.emotes.frankerfacez = make(map[string]*emoteList)

	self.regexes.channels    = make(map[string]*regexp.Regexp)
	self.regexes.subscribers = make(map[string]*regexp.Regexp)
}

func (self *emoteManager) FindEmotesInMessage(msg twitch.TextMessage) emoteList {
	// no emotes fetched yet
	if len(self.emotes.global) == 0 || len(msg.Text()) < 3 {
		return make(emoteList, 0)
	}

	chanName := msg.Channel().Name

	// Search for any channel-global emotes
	chanRegex, exists := self.regexes.channels[chanName]
	if !exists {
		chanRegex = self.buildChannelRegex(chanName)
		self.regexes.channels[chanName] = chanRegex
	}

	var result emoteList

	emotes := chanRegex.FindAllString(msg.Text(), -1)
	if emotes != nil {
		result = emotes
	}

	// If the user is a subscriber to any channel, as also search for his
	// subscriber emotes. In order to minimize the memory footprint, we are
	// storing the regexes per emotesets (JSON-encoded) value, so if two or
	// more users are subscribed to the same channels, they will both use
	// the same regex object.

	set := msg.User().EmoteSet

	if len(set) > 0 {
		sort.Ints(set)

		setKey := fmt.Sprintf("%v", set)

		// Search for any channel-global emotes
		setRegex, exists := self.regexes.subscribers[setKey]
		if !exists {
			setRegex = self.buildEmoteSetRegex(set)
			self.regexes.subscribers[setKey] = setRegex
		}

		if setRegex != nil {
			emotes = setRegex.FindAllString(msg.Text(), -1)

			if emotes != nil {
				for _, emote := range emotes {
					result = append(result, emote)
				}
			}
		}
	}

	return result
}

func (self *emoteManager) buildChannelRegex(channel string) *regexp.Regexp {
	// calculate initial capacity

	capacity := len(self.emotes.global)

	ffzGlobalEmotes, hasFfzGlobal := self.emotes.frankerfacez["global"]
	if hasFfzGlobal {
		capacity += len(*ffzGlobalEmotes)
	}

	ffzChanEmotes, hasFfzChan := self.emotes.frankerfacez[channel]
	if hasFfzChan {
		capacity += len(*ffzChanEmotes)
	}

	codes := make([]string, 0, capacity)

	// collect all relevant, user-independent emotes

	for _, code := range self.emotes.global {
		codes = append(codes, code)
	}

	if hasFfzGlobal {
		for _, code := range *ffzGlobalEmotes {
			codes = append(codes, code)
		}
	}

	if hasFfzChan {
		for _, code := range *ffzChanEmotes {
			codes = append(codes, code)
		}
	}

	// build the regex

	expression := fmt.Sprintf(`\b(%s)\b`, strings.Join(codes, "|"))

	return regexp.MustCompile(expression)
}

func (self *emoteManager) buildEmoteSetRegex(set []int) *regexp.Regexp {
	codes := make([]string, 0)

	for _, setID := range set {
		emotes, exists := self.emotes.subscribers[setID]
		if exists {
			for _, emote := range *emotes {
				codes = append(codes, emote)
			}
		}
	}

	if len(codes) == 0 {
		return nil
	}

	expression := fmt.Sprintf(`\b(%s)\b`, strings.Join(codes, "|"))

	return regexp.MustCompile(expression)
}

func (self *emoteManager) UpdateEmotes() error {
	// reset()

	err := self.updateTwitchEmotes()
	if err != nil {
		return err
	}

	// stamp("fetched twitch emotes")

	err = self.updateFrankerFaceZEmotes()

	// stamp("fetched ffz emotes")

	return err
}

func (self *emoteManager) updateTwitchEmotes() error {
	// fetch URL
	response, err := http.Get("https://api.twitch.tv/kraken/chat/emoticons")
	if err != nil {
		return fmt.Errorf("Could not fetch Twitch emotes: %s", err.Error())
	}

	// parse JSON
	emoteData := twitchEmoticons{}

	if json.NewDecoder(response.Body).Decode(&emoteData) != nil {
		return fmt.Errorf("Received invalid JSON for Twitch emotes.")
	}

	isIrregular := regexp.MustCompile(`[?$\[\]()\\;]`)

	self.reset()

	for _, emoteStruct := range emoteData.Emoticons {
		emoteCode := emoteStruct.Regex

		if isIrregular.MatchString(emoteCode) {
			continue
		}

		for _, image := range emoteStruct.Images {
			emoteSetID := image.EmoticonSet

			var list *emoteList

			if emoteSetID == nil {
				list = &self.emotes.global
			} else {
				l, exists := self.emotes.subscribers[*emoteSetID]
				if exists {
					list = l
				} else {
					newList := make(emoteList, 0)

					self.emotes.subscribers[*emoteSetID] = &newList
					list = &newList
				}
			}

			*list = append(*list, emoteCode)
		}
	}

	return nil
}

func (self *emoteManager) updateFrankerFaceZEmotes() error {
	// fetch URL
	response, err := http.Get("http://frankerfacez.com/users.txt")
	if err != nil {
		return fmt.Errorf("Could not fetch FFZ emotes: %s", err.Error())
	}

	self.emotes.frankerfacez = make(map[string]*emoteList)

	// read text file line by line
	defer response.Body.Close()

	scanner  := bufio.NewScanner(response.Body)
	lastChan := ""

	for scanner.Scan() {
		line := scanner.Text()

		if string(line[0]) == "." {
			*self.emotes.frankerfacez[lastChan] = append(*self.emotes.frankerfacez[lastChan], line[1:])
		} else {
			_, exists := self.emotes.frankerfacez[line]
			if !exists {
				newList := make(emoteList, 0)
				self.emotes.frankerfacez[line] = &newList
			}

			lastChan = line
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// var last = time.Now()

// func stamp(name string) {
// 	now  := time.Now()
// 	diff := now.Sub(last)

// 	last = now

// 	var mem runtime.MemStats
// 	runtime.ReadMemStats(&mem)

// 	fmt.Printf("%s: %s (res = %.3f MiB, alloc = %.3f MiB, total alloc = %.3f MiB)\n", name, diff.String(), float64(mem.Sys) / (1024*1024), float64(mem.Alloc) / (1024*1024), float64(mem.TotalAlloc) / (1024*1024))
// }

// func reset() {
// 	last = time.Now()
// 	stamp("RESET")
// }
