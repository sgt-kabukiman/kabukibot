package bot

import "log"

type dict map[string]string

// The Dictionary is a glorified string/string map that's kept in sync with a database table.
type Dictionary struct {
	db   *DatabaseStruct
	data dict
}

func NewDictionary(db *DatabaseStruct) *Dictionary {
	return &Dictionary{db, make(dict)}
}

func (self *Dictionary) Keys() []string {
	list := make([]string, len(self.data))
	idx  := 0

	for key, _ := range self.data {
		list[idx] = key
		idx       = idx + 1
	}

	return list
}

func (self *Dictionary) Add(key string, value string) *Dictionary {
	_, exists := self.data[key]
	if !exists {
		self.data[key] = value

		_, err := self.db.Exec("INSERT INTO dictionary (keyname, value) VALUES (?, ?)", key, value)
		if err != nil {
			log.Fatal("Could not add dictionary entry '" + key + "' to the database: " + err.Error())
		}
	}

	return self
}

func (self *Dictionary) Set(key string, value string) *Dictionary {
	_, exists := self.data[key]
	if !exists {
		return self.Add(key, value)
	}

	_, err := self.db.Exec("UPDATE dictionary SET value = ? WHERE keyname = ?", value, key)
	if err != nil {
		log.Fatal("Could not update dictionary entry '" + key + "' in the database: " + err.Error())
	}

	return self
}

func (self *Dictionary) Get(key string) string {
	value, exists := self.data[key]
	if exists {
		return value
	}

	return ""
}

func (self *Dictionary) Has(key string) bool {
	_, exists := self.data[key]
	return exists
}

func (self *Dictionary) Delete(key string) *Dictionary {
	_, exists := self.data[key]
	if exists {
		delete(self.data, key)

		_, err := self.db.Exec("DELETE FROM dictionary WHERE keyname = ?", key)
		if err != nil {
			log.Fatal("Could not remove dictionary entry '" + key + "' from the database: " + err.Error())
		}
	}

	return self
}

func (self *Dictionary) load() {
	rows, err := self.db.Query("SELECT * FROM dictionary")
	if err != nil {
		log.Fatal("Could not query the dictionary: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			log.Fatal(err)
		}

		self.data[key] = value
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
