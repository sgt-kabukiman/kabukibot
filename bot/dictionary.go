package bot

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

type dict map[string]string

// The Dictionary is a glorified string/string map that's kept in sync with a database table.
type Dictionary struct {
	db    *sqlx.DB
	log   Logger
	data  dict
	mutex sync.RWMutex
}

func NewDictionary(db *sqlx.DB, log Logger) *Dictionary {
	return &Dictionary{db, log, make(dict), sync.RWMutex{}}
}

func (self *Dictionary) Keys() []string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	list := make([]string, len(self.data))
	idx := 0

	for key, _ := range self.data {
		list[idx] = key
		idx = idx + 1
	}

	return list
}

func (self *Dictionary) Add(key string, value string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	_, exists := self.data[key]
	if !exists {
		self.data[key] = value

		_, err := self.db.Exec("INSERT INTO dictionary (keyname, value) VALUES (?, ?)", key, value)
		if err != nil {
			self.log.Fatal("Could not add dictionary entry '" + key + "' to the database: " + err.Error())
		}

		self.log.Debug("Added dictionary entry '%s' as '%s'.", key, value)
	}
}

func (self *Dictionary) Set(key string, value string) {
	self.mutex.Lock()

	_, exists := self.data[key]
	if !exists {
		self.mutex.Unlock()
		self.Add(key, value)
		return
	}

	self.data[key] = value

	_, err := self.db.Exec("UPDATE dictionary SET value = ? WHERE keyname = ?", value, key)
	if err != nil {
		self.log.Fatal("Could not update dictionary entry '" + key + "' in the database: " + err.Error())
	}

	self.log.Debug("Updated dictionary entry '%s' with '%s'.", key, value)
	self.mutex.Unlock()
}

func (self *Dictionary) Get(key string) string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	value, exists := self.data[key]
	if exists {
		return value
	}

	return ""
}

func (self *Dictionary) Has(key string) bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	_, exists := self.data[key]
	return exists
}

func (self *Dictionary) Delete(key string) *Dictionary {
	_, exists := self.data[key]
	if exists {
		delete(self.data, key)

		_, err := self.db.Exec("DELETE FROM dictionary WHERE keyname = ?", key)
		if err != nil {
			self.log.Fatal("Could not remove dictionary entry '" + key + "' from the database: " + err.Error())
		}

		self.log.Debug("Deleted dictionary entry '%s'.", key)
	}

	return self
}

type dictRow struct {
	Keyname string
	Value   string
}

func (self *Dictionary) load() {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	list := make([]dictRow, 0)
	self.db.Select(&list, "SELECT * FROM dictionary ORDER BY keyname")

	for _, item := range list {
		self.data[item.Keyname] = item.Value
	}

	self.log.Debug("Loaded %d dictionary entries.", len(list))
}
